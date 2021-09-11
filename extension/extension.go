// Package exsention interfaces
package extension

import (
	"net/http"
)

type Extention interface {
	Configurable
	Initializable
	Disposable
}

type Configurable interface {
	// return reference of struct
	Config() interface{}
}

type Initializable interface {
	Init() error
}

type Disposable interface {
	Dispose() error
}

type HttpExtention interface {
	Extention
	Handle(http.Handler) http.Handler
}
