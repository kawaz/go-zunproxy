package main

import (
	"fmt"
	"io"
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

	mc := memcache.New(cfg.Memcached...)
	cache := &handlers.CacheHandler{MemcachedClient: mc}

	backendUrl, err := url.Parse(cfg.Backend)
	if err != nil {
		panic(fmt.Errorf("could not parse backend: %v", cfg.Backend))
	}
	backendProxy := httputil.NewSingleHostReverseProxy(backendUrl)

	dump := handlers.NewDumpHandler(cfg.DumpDir)

	bundler := handlers.NewRequestBundlerDefault()

	http.Handle("/",
		dump.Handle(
			bundler.Handle(
				cache.Handle(
					backendProxy))))

	fmt.Fprint(io.Discard, cache, dump, bundler)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
}
