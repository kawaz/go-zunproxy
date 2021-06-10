package zunproxy

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"log"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/goccy/go-json"
	"github.com/k0kubun/pp"
)

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
	Status int
	Header http.Header
	Body   []byte
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
		Status: rr.Code,
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
	//	w.WriteHeader(cr.Status)
	w.Write(cr.Body)
}

func (cr *CachedResponse) Bytes() []byte {
	bytes, err := json.Marshal(cr)
	if err != nil {
		panic(fmt.Errorf("could not marchal CachedResponse: %v", err))
	}
	return bytes
}

func (cache *CacheHandler) MakeCacheKey(prefix string, key string) string {
	k := prefix + base32.StdEncoding.EncodeToString(sha256.New().Sum([]byte(key)))
	if 250 < len(k) {
		return k[:250]
	}
	return k
}

func (cache *CacheHandler) getCacheInfo(r *http.Request) (*CacheInfo, error) {
	rKeySource := r.Method + " " + r.Host + r.URL.Path + "?" + r.URL.RawQuery
	rKey := cache.MakeCacheKey("req:", rKeySource)
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

func (cache *CacheHandler) FrontHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ci, err := cache.getCacheInfo(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("cache error: %v", err), http.StatusInternalServerError)
			return
		}
		if time.Now().Before(ci.Expires) {
			ci.CachedResponse.WriteTo(w)
			return
		}
		// キャッシュ更新は確定なのでリクエストのコンテキストにCacheInfoをセット
		ctx := context.WithValue(r.Context(), CtxKeyCacheInfo, ci)
		if ci.CachedResponse != nil {
			ci.CachedResponse.WriteTo(w)
			next.ServeHTTP(httptest.NewRecorder(), r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func (cache *CacheHandler) SaveHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ci, ok := r.Context().Value(CtxKeyCacheInfo).(*CacheInfo)
		if !ok {
			// キャッシュ対象じゃないのでスルー
			next.ServeHTTP(w, r)
			return
		}
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)
		cr := &CachedResponse{
			Status: rec.Code,
			Header: rec.Header(),
			Body:   rec.Body.Bytes(),
		}
		cr.WriteTo(w)
		go cache.update(ci, cr)
	})
}

func (cache *CacheHandler) update(ci *CacheInfo, cr *CachedResponse) {
	isNew := ci.CachedResponse == nil
	if cr != nil {
		ci.CachedResponse = cr
		ci.Expires = time.Now().Add(time.Second * 60)
	}
	bytes, err := json.Marshal(ci)
	if err != nil {
		pp.Println("could not marchal CacheInfo", err)
	}
	ci.mcItem.Value = bytes
	//err = cache.MemcachedClient.CompareAndSwap(ci.mcItem)
	err = cache.MemcachedClient.Set(ci.mcItem)
	if err != nil {
		log.Fatalf("could not save to memcached: %v", err)
	}
	if isNew {
		log.Printf("CREATE_CACHE: %v", ci.ReqKeySource)
	} else {
		log.Printf("UPDATE_CACHE: %v", ci.ReqKeySource)
	}
}

type contextKey string

const CtxKeyCacheInfo contextKey = "CacheInfo"
