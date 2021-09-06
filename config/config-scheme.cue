import "time"

// #Middleware: {
//   Constructor: {
//     Name: string
//     Args: [..._]
//   }
// }

Port: int16 | *3000
Backend: string & =~ "^https?://[a-z0-9_\\.-]+(:[0-9]+)?$"

// ミドルウェア
DumpDir?: string
Bundler?: bool
Memcached?: [
  ...( string & =~ "^[a-z0-9_\\.-]+:[0-9]+$" )
]
// CacheTTL?: time.Duration | *120
// Listeners: [#Listener, ...#Listener]
// Routes: [...#Route]
// Backends: [string]: #Backend

Port: uint16

//EndPoints: #Endpoint

#Listener: {
	Port: uint16
	Middlewares: [...#Middleware]
}

#Middleware: {
	Type: string,
	Disabled: bool | *false,
	Middlewares?: [...#Middleware]
	...
}

#Route: {
	Backend: string
	...
}

#Filter: Matcher: #Matcher

#Matcher: string

//Backends: [_Name=@]: {Name: _Name}
#Backend: {
	Type: string
} & ( #BackendProxy | #BackendMock | #BackendStatic )

#BackendProxy: {
	Type: "Proxy"
	Url:  string
}
#BackendMock: {
	Type: "Mock"
	Routes: [#Method]: {
		Code:  int & >=100 & <600
		Tyep?: "text/plain; charset=UTF-8"
		Body:  string
	}
}
#BackendStatic: {
	Type:         "Static"
	DocumentRoot: string
}

#HealthCheck: #HttpHealthcheck | #TcpHealthcheck | #PingHealthcheck
#HttpHealthcheck: {
	Path:           string
	Code:           int
	HealthyCount:   int | *1
	UhhealtyCount:  int | *2
	Interval:       time.Duration | *"30s"
	Timeout:        time.Duration | *"5s"
	ConnectTimeout: time.Duration | *"3s"
}
#TcpHealthcheck: {
	ConnectTimeout: time.Duration | *"3s"
	Send?:          string | bytes
	Reponse?:       string | bytes
}
#PingHealthcheck: {
	Timeout: time.Duration | *"3s"
}

#Method: *"GET" | "POST" | "HEAD" | "OPTIONS" | "POST" | "PUT" | "DELETE" | "PATCH"
