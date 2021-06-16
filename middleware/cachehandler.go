package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"log"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/goccy/go-json"
)

func NewCacheHandler(mc *memcache.Client) Middleware {
	return &CacheHandler{MemcachedClient: mc}
}

type CacheHandler struct {
	MemcachedClient *memcache.Client
}

// キャッシュの情報
type CacheInfo struct {
	// Request を表す文字列
	ReqKeySource string
	// キャシュの有効期限
	Expires time.Time
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
	return &CachedResponse{
		Code:   rr.Code,
		Header: rr.Header(),
		Body:   rr.Body.Bytes(),
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
		if ci.CachedResponse != nil && time.Now().Before(ci.Expires) {
			// キャッシュが有効なのですぐ返して終了
			ci.CachedResponse.WriteTo(w)
			return
		}
		// キャッシュ更新は確定
		// 失敗しててもやることは変わらないので error は無視
		_ = cache.updateCacheInfo(ci)
		// Responseを取り出せるようにしておく
		buf := bytes.NewBuffer([]byte{})
		rec := NewResponseRecorder(w, buf)
		next.ServeHTTP(rec, r)
		isNew := ci.CachedResponse == nil
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
		if isNew {
			log.Printf("CREATE_CACHE: %v", ci.ReqKeySource)
		} else {
			log.Printf("UPDATE_CACHE: %v", ci.ReqKeySource)
		}
	})
}

func (cache *CacheHandler) makeCacheKey(prefix string, key string) string {
	k := prefix + base32.StdEncoding.EncodeToString(sha256.New().Sum([]byte(key)))
	if 250 < len(k) {
		return k[:250]
	}
	return k
}

func (cache *CacheHandler) getCacheInfo(r *http.Request) (*CacheInfo, error) {
	rKeySource := r.Method + " " + r.Host + r.URL.Path + "?" + r.URL.RawQuery
	rKey := cache.makeCacheKey("req:", rKeySource)
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
			ReqKeySource: rKeySource,
			mcItem:       &memcache.Item{Key: rKey},
		}
	}
	return &ci, nil
}

func (cache *CacheHandler) updateCacheInfo(ci *CacheInfo) error {
	ci.Expires = time.Now().Add(time.Second * 120)
	ciBytes, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("could not marshal CacheInfo: %v", err)
	}
	ci.mcItem.Value = ciBytes
	return cache.MemcachedClient.Set(ci.mcItem)
}
