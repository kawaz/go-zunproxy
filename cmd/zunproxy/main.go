package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/k0kubun/pp"
	"github.com/kawaz/zunproxy"
)

func main() {

	fmt.Println("zunproxy")
	pp.Println(os.Getwd())

	config, err := Load()
	if err != nil {
		panic(err)
	}
	pp.Println(config)

	mc := memcache.New(config.Memcached...)
	cache := zunproxy.CacheHandler{MemcachedClient: mc}

	backendUrl, err := url.Parse(config.Backend)
	if err != nil {
		panic(fmt.Errorf("could not parse backend: %v", config.Backend))
	}
	backendProxy := httputil.NewSingleHostReverseProxy(backendUrl)

	// http.Handle("/",
	// 	cache.FrontHandler(
	// 		zunproxy.NewDuplicatePreventerDefault().Handler(
	// 			cache.SaveHandler(
	// 				backendProxy))))

	http.Handle("/",
		cache.FrontHandler(
			cache.SaveHandler(
				backendProxy)))

	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
