package cache

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"log"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/goccy/go-json"
	"github.com/kawaz/go-zunproxy/extension"
	"github.com/kawaz/go-zunproxy/middleware"
)

var _ extension.HttpExtention = &cacheExtension{}

func NewCacheExtension(config *CacheConfig) *cacheExtension {
	return &cacheExtension{
		config: config,
	}
}

type cacheExtension struct {
	extension.NopExtension
	config *CacheConfig
	mc     *memcache.Client
}

// func(ex *cacheExtension)

type CacheConfig struct {
	// memcached サーバリスト
	MemcachedServers []string
	// キャッシュの更新期間
	SoftTTL time.Duration
	// キャッシュエントリを削除する期間
	HardTTL time.Duration
	// バックエンドのレスポンスコードに応じたTTL
	// ErrorTTL map[int]int
	// キャッシュ更新時にバックエンドからの最新レスポンスを待つ時間。バックエンドのレスポンスがこれより遅い場合は古いキャッシュを返す。
	NewResponseTimeout time.Duration
}

func (ext *cacheExtension) Config() interface{} {
	return &ext.config
}

func (ext *cacheExtension) Dispose() error {
	ext.mc = nil
	return nil
}

// キャッシュの情報
type cacheInfo struct {
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
	CachedResponse *cachedResponse
	// 元になった Item を更新用に保持しておく
	mcItem *memcache.Item
}

func newCacheInfo(item *memcache.Item) (*cacheInfo, error) {
	if item == nil {
		item = &memcache.Item{}
	}
	var ci *cacheInfo
	err := json.Unmarshal(item.Value, &ci)
	if err != nil {
		return nil, fmt.Errorf("could not unmatchal CacheInfo: %v", err)
	}
	ci.mcItem = item
	return ci, nil
}

func (ci *cacheInfo) Bytes() []byte {
	bytes, err := json.Marshal(ci)
	if err != nil {
		panic(fmt.Errorf("could not marchal CacheInfo: %v", err))
	}
	return bytes
}

type cachedResponse struct {
	Code          int
	ContentLength int
	Header        http.Header
	Body          []byte
	Enc           string
	// 元になった Item を更新用に保持しておく
	mcItem *memcache.Item
}

func newCacheResponse(item *memcache.Item) (*cachedResponse, error) {
	if item == nil {
		item = &memcache.Item{}
	}
	var cr *cachedResponse
	err := json.Unmarshal(item.Value, &cr)
	if err != nil {
		return nil, fmt.Errorf("could not unmatchal CacheInfo: %v", err)
	}
	cr.mcItem = item
	return cr, nil
}

func newCacheResponseFromRecorder(rr *httptest.ResponseRecorder) *cachedResponse {
	// select enc := rr.Header().Get("Content-Encoding") {
	// case "gzip":

	// }
	return &cachedResponse{
		Code:   rr.Code,
		Header: rr.Header(),
		Body:   rr.Body.Bytes(),
		Enc:    "",
	}
}

func (cr *cachedResponse) WriteTo(w http.ResponseWriter) {
	for k, values := range cr.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(cr.Code)
	w.Write(cr.Body)
}

func (ext *cacheExtension) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ci, err := ext.getCacheInfo(r)
		if err != nil {
			// memcached で何かエラー
			log.Print(err)
			// 普通にキャッシュなしでスルー
			next.ServeHTTP(w, r)
			return
		}
		if ci.CachedResponse != nil && time.Now().Before(ci.Expires) {
			// キャッシュが有効なのですぐ返して終了
			ci.CachedResponse.WriteTo(w)
			return
		}
		// キャッシュ更新は確定
		tsStart := time.Now()
		// 他リクエストが同時にキャッシュ更新するのを避けるためにまずキャッシュのExpiresを伸ばしておく
		// 失敗しててもやることは変わらないので error は無視
		_ = ext.updateCacheInfo(ci)
		// Responseを取り出せるようにしておく
		rec := middleware.NewResponseRecorder(w)
		buf := bytes.NewBuffer([]byte{})
		rec.AddWriter(buf)
		//ついでにハッシュも計算しておく
		hash := sha256.New()
		rec.AddWriter(hash)
		next.ServeHTTP(rec, r)
		isNew := ci.CachedResponse == nil
		// キャッシュを保存
		ci.CachedResponse = &cachedResponse{
			Code:          rec.Code(),
			ContentLength: rec.ContentLength(),
			Header:        rec.Header().Clone(),
			Body:          buf.Bytes(),
		}
		err = ext.updateCacheInfo(ci)
		if err != nil {
			log.Fatalf("could not save CacheInfo: %v", err)
			return
		}
		//ログ
		action := "UPDATE"
		if isNew {
			action = "CREATE"
		}
		log.Printf("%v %v %10s %v %v", action, ci.Key, time.Since(tsStart).Truncate(time.Millisecond), ci.CachedResponse.Code, ci.KeySource)
	})
}

func (ext *cacheExtension) makeCacheKey(prefix string, key string) string {
	hash := sha256.New()
	hash.Write([]byte(key))
	sum := hash.Sum(nil)
	k := prefix + middleware.Base32.EncodeToString(sum)
	if 250 < len(k) {
		return k[:250]
	}
	return k
}

func (ext *cacheExtension) getCacheInfo(r *http.Request) (*cacheInfo, error) {
	rKeySource := r.Method + " " + r.Host + r.URL.Path + "?" + r.URL.RawQuery
	rKey := ext.makeCacheKey("ch/", rKeySource)
	item, err := ext.mc.Get(rKey)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			return nil, fmt.Errorf("could not load CacheInfo: %v", err)
		}
	}
	var ci cacheInfo
	if item != nil {
		ci.mcItem = item
		err = json.Unmarshal(item.Value, &ci)
		if err != nil {
			return nil, fmt.Errorf("could not Unmarchal CacheInfo: %v", err)
		}
	} else {
		ci = cacheInfo{
			KeySource: rKeySource,
			Key:       rKey,
			mcItem:    &memcache.Item{Key: rKey},
		}
	}
	return &ci, nil
}

func (ext *cacheExtension) updateCacheInfo(ci *cacheInfo) error {
	ci.Expires = time.Now().Add(time.Second * 120)
	ciBytes, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("could not marshal CacheInfo: %v", err)
	}
	ci.mcItem.Value = ciBytes
	return ext.mc.Set(ci.mcItem)
}
