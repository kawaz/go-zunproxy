import "time"

Port: 3000

Backend: "http://backend:8080"

#DumpDir: "/tmp/zunproxy-dump/%Y/%m/%d/%H/%M"

Bundler: false

Cache: {
    // memcached サーバリスト
    MemcachedServers: [
        "memcached-1:11211",
        "memcached-2:11211",
    ]

    // キャッシュの更新期間
    SoftTTL: time.ParseDuration("120s")

    // キャッシュエントリを削除する期間
    HardTTL: time.ParseDuration("24h")

    // キャッシュ更新時にバックエンドからの最新レスポンスを待つ時間。バックエンドのレスポンスがこれより遅い場合は古いキャッシュを返す。
    NewResponseWaitLimit: time.ParseDuration("20ms")

    // バックエンドのレスポンスコードに応じたTTL
    ErrorTTL: "404": time.ParseDuration("4s")
    ErrorTTL: "413": time.ParseDuration("0s")
    ErrorTTL: "500": time.ParseDuration("5s")

    // キャッシュする最大レスポンスサイズ(ヘッダやエンコードを含め1MBを超えるとmemcachedに保存できない等のケース対応)
    BytesLimit: 700K
}

