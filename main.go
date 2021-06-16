package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/k0kubun/pp"
	"github.com/kawaz/zunproxy/config"
	"github.com/kawaz/zunproxy/middleware"
)

func MustValue(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {

	log.Printf("cwd: %v", MustValue(os.Getwd()))

	cfg, err := config.Load("zunproxy.cue")
	if err != nil {
		panic(err)
	}
	pp.Println(cfg)

	// ミドルウェア
	var middlewares []middleware.Middleware
	if cfg.DumpDir != "" {
		dump := middleware.NewDumpHandler(cfg.DumpDir)
		middlewares = append(middlewares, dump)
	}
	if cfg.Bundler {
		bundler := middleware.NewRequestBundlerDefault()
		middlewares = append(middlewares, bundler)
	}
	if 0 < len(cfg.Memcached) {
		cache := middleware.NewCacheHandler(memcache.New(cfg.Memcached...))
		middlewares = append(middlewares, cache)
	}

	// ハンドラ
	backendUrl, err := url.Parse(cfg.Backend)
	if err != nil {
		panic(fmt.Errorf("could not parse backend: %v", cfg.Backend))
	}
	backendProxy := httputil.NewSingleHostReverseProxy(backendUrl)

	// 起動
	handler := middleware.MultipleHandler(backendProxy, middlewares...)
	http.Handle("/", handler)
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("zunproxy start at %v -> %v", addr, cfg.Backend)
	http.ListenAndServe(addr, nil)
}
