package handlers

import (
	"net/http"
)

type Middleware interface {
	Handle(http.Handler) http.Handler
}

func MultipleHandler(h http.Handler, ms ...Middleware) http.Handler {
	if len(ms) == 0 {
		return h
	}
	if len(ms) == 1 {
		return ms[0].Handle(h)
	}
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i].Handle(h)
	}
	return h
}
