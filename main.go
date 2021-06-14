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
	"github.com/kawaz/zunproxy/handlers"
)

func MustValue(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {

	log.Println("zunproxy started")
	log.Printf("cwd: %v", MustValue(os.Getwd()))

	cfg, err := config.Load("zunproxy.cue")
	if err != nil {
		panic(err)
	}
	pp.Println("load config", cfg)

	// ミドルウェア
	middlewares := []handlers.Middleware{
		handlers.NewDumpHandler(cfg.DumpDir),
		handlers.NewRequestBundlerDefault(),
		handlers.NewCacheHandler(memcache.New(cfg.Memcached...)),
	}

	// ハンドラ
	backendUrl, err := url.Parse(cfg.Backend)
	if err != nil {
		panic(fmt.Errorf("could not parse backend: %v", cfg.Backend))
	}
	backendProxy := httputil.NewSingleHostReverseProxy(backendUrl)

	// 起動
	handler := handlers.MultipleHandler(backendProxy, middlewares...)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler)
}
