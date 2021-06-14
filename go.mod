module github.com/kawaz/zunproxy

go 1.16

require (
	cuelang.org/go v0.2.2
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/goccy/go-json v0.6.1
	github.com/itchyny/timefmt-go v0.1.3
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/kawaz/go-requestid v0.0.0-20201222065628-3590f9fdeed4
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/oklog/ulid v1.3.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// replace github.com/kawaz/go-requestid => $GOPATH/src/github.com/kawaz/go-requestid
replace github.com/kawaz/go-zunproxy => ./
