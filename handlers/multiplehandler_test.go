package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("[h]")) })
var wrapMiddleware1 Middleware = MiddlewareFunc(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<m1"))
		next.ServeHTTP(w, r)
		w.Write([]byte("m1>"))
	})
})
var wrapMiddleware2 Middleware = MiddlewareFunc(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<m2"))
		next.ServeHTTP(w, r)
		w.Write([]byte("m2>"))
	})
})
var wrapMiddleware3 Middleware = MiddlewareFunc(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<m3"))
		next.ServeHTTP(w, r)
		w.Write([]byte("m3>"))
	})
})
var noopMiddleware Middleware = MiddlewareFunc(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
})

func TestMultipleHandler(t *testing.T) {
	// m1 := Middleware
	type args struct {
		h  http.Handler
		ms []Middleware
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no middleware",
			args: args{h: h},
			want: "[h]",
		},
		{
			name: "1 middleware",
			args: args{h: h, ms: []Middleware{wrapMiddleware1}},
			want: "<m1[h]m1>",
		},
		{
			name: "2 middleware",
			args: args{h: h, ms: []Middleware{wrapMiddleware1, wrapMiddleware2}},
			want: "<m1<m2[h]m2>m1>",
		},
		{
			name: "3 middleware",
			args: args{h: h, ms: []Middleware{wrapMiddleware1, wrapMiddleware2, wrapMiddleware3}},
			want: "<m1<m2<m3[h]m3>m2>m1>",
		},
		{
			name: "has Middleware that does not chain to next",
			args: args{h: h, ms: []Middleware{wrapMiddleware1, noopMiddleware, wrapMiddleware2}},
			want: "<m1m1>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			mh := MultipleHandler(tt.args.h, tt.args.ms...)
			mh.ServeHTTP(rec, req)
			got := rec.Body.String()
			if got != tt.want {
				t.Errorf("rec.Body = %v, want %v", got, tt.want)
			}
		})
	}
}
