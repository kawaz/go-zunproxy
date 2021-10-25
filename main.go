package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"

	"github.com/k0kubun/pp"
	"github.com/kawaz/go-zunproxy/config"
	"github.com/kawaz/go-zunproxy/middleware"
)

func MustValue(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

func main() {
	cfgfile := flag.String("config", "./zunproxy.cue", "config file")
	flg_version := flag.Bool("v", false, "show version")
	flg_version_long := flag.Bool("V", false, "show version long")
	flg_help := flag.Bool("h", false, "show help")
	flag.Parse()

	if *flg_version {
		fmt.Println(path.Base(os.Args[0]), build)
		os.Exit(0)
	}
	if *flg_version_long {
		fmt.Println(path.Base(os.Args[0]), build)
		json, _ := json.MarshalIndent(build, "", "\t")
		fmt.Println(string(json))
		os.Exit(0)
	}
	if *flg_help {
		flag.Usage()
		os.Exit(0)
	}
	cfg, err := config.Load(*cfgfile)
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
	// 同じリクエストの同時処理を1つに制限して結果を共有する
	if cfg.Bundler {
		bundler := middleware.NewRequestBundlerDefault()
		middlewares = append(middlewares, bundler)
	}
	// レスポンスキャッシュ
	if cfg.Cache != nil {
		cache := middleware.NewCacheHandler(cfg.Cache)
		middlewares = append(middlewares, cache)
	}
	// 壊れたレスポンスをエラーにする奴
	middlewares = append(middlewares, middleware.NewBrokenRewriteGuard())
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
	log.Fatal(http.ListenAndServe(addr, nil))
}
