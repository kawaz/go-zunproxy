package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"log"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/goccy/go-json"
)

type Configurable interface {
	Config() interface{}
}

type CacheMiddleware interface {
	Middleware
	Configurable
}

type CacheConfig struct {
	// memcached サーバリスト
	MemcachedServers []string
	// キャッシュの更新期間
	SoftTTL time.Duration
	// キャッシュエントリを削除する期間
	HardTTL time.Duration
	// キャッシュ更新時にバックエンドからの最新レスポンスを待つ時間。バックエンドのレスポンスがこれより遅い場合は古いキャッシュを返す。
	NewResponseWaitLimit time.Duration
	// バックエンドのレスポンスコードに応じたTTL
	ErrorTTL map[int]time.Duration
}

func NewCacheHandler(config *CacheConfig) Middleware {
	mc := memcache.New(config.MemcachedServers...)
	return &CacheHandler{
		MemcachedClient: mc,
		config:          config,
	}
}

type CacheHandler struct {
	MemcachedClient *memcache.Client
	config          *CacheConfig
}

// キャッシュの情報
type CacheInfo struct {
	// Request を表す文字列
	KeySource string
	// memcachedのキー(<250byte)
	Key string
	// キャシュの有効期限
	Expires time.Time
	// キャッシュエントリが作られた
	Created time.Time
	// ボディが更新された
	Updated time.Time
	// 更新回数
	UpCount int
	// 総処理時間
	UpDurations time.Duration
	// ボディのハッシュ b64url(sha256(body))
	BodyHash string
	// キャッシュされたレスポンス
	CachedResponse *CachedResponse
	// 元になった Item を更新用に保持しておく
	mcItem *memcache.Item
}

func NewCacheInfo(item *memcache.Item) (*CacheInfo, error) {
	if item == nil {
		item = &memcache.Item{}
	}
	var ci *CacheInfo
	err := json.Unmarshal(item.Value, &ci)
	if err != nil {
		return nil, fmt.Errorf("could not unmatchal CacheInfo: %v", err)
	}
	ci.mcItem = item
	return ci, nil
}

func (ci *CacheInfo) Bytes() []byte {
	bytes, err := json.Marshal(ci)
	if err != nil {
		panic(fmt.Errorf("could not marchal CacheInfo: %v", err))
	}
	return bytes
}

type CachedResponse struct {
	Code          int
	ContentLength int
	Header        http.Header
	Body          []byte
	Enc           string
	// 元になった Item を更新用に保持しておく
	mcItem *memcache.Item
}

func NewCacheResponse(item *memcache.Item) (*CachedResponse, error) {
	if item == nil {
		item = &memcache.Item{}
	}
	var cr *CachedResponse
	err := json.Unmarshal(item.Value, &cr)
	if err != nil {
		return nil, fmt.Errorf("could not unmatchal CacheInfo: %v", err)
	}
	cr.mcItem = item
	return cr, nil
}

func NewCacheResponseFromRecorder(rr *httptest.ResponseRecorder) *CachedResponse {
	// select enc := rr.Header().Get("Content-Encoding") {
	// case "gzip":

	// }
	return &CachedResponse{
		Code:   rr.Code,
		Header: rr.Header(),
		Body:   rr.Body.Bytes(),
		Enc:    "",
	}
}

func (cr *CachedResponse) WriteTo(w http.ResponseWriter) {
	for k, values := range cr.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(cr.Code)
	w.Write(cr.Body)
}

func (cache *CacheHandler) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ci, err := cache.getCacheInfo(r)
		if err != nil {
			// memcached で何かエラー
			log.Print(err)
			// 普通にキャッシュなしでスルー
			next.ServeHTTP(w, r)
			return
		}
		if ci.CachedResponse != nil {
			if time.Now().Before(ci.Expires) {
				// キャッシュが有効なのですぐ返して終了
				ci.CachedResponse.WriteTo(w)
				return
			}
		}

		// キャッシュ更新は確定
		tsStart := time.Now()
		var isNew bool
		var rec ResponseRecorder
		if ci.CachedResponse == nil {
			isNew = true
			rec = NewResponseRecorder(w)
		} else {
			isNew = false
			rec = NewResponseSteeler()
		}

		// バックエンドにリクエストを投げる
		newCache := make(chan *CachedResponse, 1)
		go func() {
			// 他リクエストが同時にキャッシュ更新するのを避けるためにまずキャッシュのExpiresを伸ばしておく
			// 失敗しててもやることは変わらないので error は無視
			_ = cache.updateCacheInfo(ci)
			// Responseを取り出せるようにしておく
			buf := bytes.NewBuffer([]byte{})
			rec.AddWriter(buf)
			//ついでにハッシュも計算しておく
			hash := sha256.New()
			rec.AddWriter(hash)
			next.ServeHTTP(rec, r.Clone(context.Background()))
			// キャッシュを保存
			ci.CachedResponse = &CachedResponse{
				Code:          rec.Code(),
				ContentLength: rec.ContentLength(),
				Header:        rec.Header().Clone(),
				Body:          buf.Bytes(),
			}
			err = cache.updateCacheInfo(ci)
			if err != nil {
				log.Fatalf("could not save CacheInfo: %v", err)
				return
			}
			log.Printf("%v %v %10s %v %v", "UPDATE", ci.Key, time.Since(tsStart).Truncate(time.Millisecond), ci.CachedResponse.Code, ci.KeySource)
			newCache <- ci.CachedResponse
		}()

		// 新規なら更新リクエストが終わったら戻る
		if isNew {
			<-newCache
			log.Printf("%v %v %10s %v %v", "CREATE", ci.Key, time.Since(tsStart).Truncate(time.Millisecond), ci.CachedResponse.Code, ci.KeySource)
			return
		}

		// 更新の場合は、NewResponseWaitLimit 秒以内にバックエンドのレスポンスが帰ってこなければ古いキャッシュを返す
		oldCache := make(chan *CachedResponse, 1)
		go func() {
			time.Sleep(cache.config.NewResponseWaitLimit)
			oldCache <- ci.CachedResponse
		}()

		select {
		case wt := <-oldCache:
			wt.WriteTo(w)
			log.Printf("%v %v %10s %v %v", "OLDRES", ci.Key, time.Since(tsStart).Truncate(time.Millisecond), ci.CachedResponse.Code, ci.KeySource)
			return
		case wt := <-newCache:
			wt.WriteTo(w)
			return
		}
	})
}

func (cache *CacheHandler) makeCacheKey(prefix string, key string) string {
	hash := sha256.New()
	hash.Write([]byte(key))
	sum := hash.Sum(nil)
	k := prefix + Base32.EncodeToString(sum)
	if 250 < len(k) {
		return k[:250]
	}
	return k
}

func (cache *CacheHandler) getCacheInfo(r *http.Request) (*CacheInfo, error) {
	rKeySource := r.Method + " " + r.Host + r.URL.Path + "?" + r.URL.RawQuery
	rKey := cache.makeCacheKey("ch/", rKeySource)
	item, err := cache.MemcachedClient.Get(rKey)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			return nil, fmt.Errorf("could not load CacheInfo: %v", err)
		}
	}
	var ci CacheInfo
	if item != nil {
		ci.mcItem = item
		err = json.Unmarshal(item.Value, &ci)
		if err != nil {
			return nil, fmt.Errorf("could not Unmarchal CacheInfo: %v", err)
		}
	} else {
		ci = CacheInfo{
			KeySource: rKeySource,
			Key:       rKey,
			mcItem:    &memcache.Item{Key: rKey},
		}
	}
	return &ci, nil
}

func (cache *CacheHandler) updateCacheInfo(ci *CacheInfo) error {
	ci.Expires = time.Now().Add(time.Second * time.Duration(cache.config.SoftTTL))
	ciBytes, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("could not marshal CacheInfo: %v", err)
	}
	ci.mcItem.Value = ciBytes
	return cache.MemcachedClient.Set(ci.mcItem)
}
