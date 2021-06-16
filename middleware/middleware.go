package middleware

import (
	"net/http"
)

type Middleware interface {
	Handle(http.Handler) http.Handler
}

type MiddlewareFunc func(http.Handler) http.Handler

func (mf MiddlewareFunc) Handle(next http.Handler) http.Handler {
	return mf(next)
}

type MiddlewareFuncEasy func(w http.ResponseWriter, r *http.Request, next http.Handler)

func (mf MiddlewareFuncEasy) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mf(w, r, next)
	})
}
