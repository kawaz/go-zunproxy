package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestMiddlewareFunc_Handle(t *testing.T) {
	type args struct {
		next http.Handler
	}
	type want struct {
		code   int
		body   string
		header http.Header
	}
	tests := []struct {
		name string
		mf   Middleware
		args args
		want want
	}{
		{
			name: "Middleware that add request header",
			mf: MiddlewareFunc(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r.Header.Set("X-Foo", "unko")
					next.ServeHTTP(w, r)
				})
			}),
			args: args{http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(r.Header.Get("X-Foo")))
			})},
			want: want{
				code:   200,
				body:   "unko",
				header: http.Header{},
			},
		},
		{
			name: "Middleware that set response header",
			mf: MiddlewareFunc(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
					originalHeader := w.Header().Clone()
					parseInt := func(s string) int64 {
						i, _ := strconv.ParseInt(s, 10, 64)
						return i
					}
					w.Header().Set("X-Set", "aaa")
					w.Header().Set("X-Set-Double", strconv.FormatInt(2*parseInt(originalHeader.Get("X-Set-Double")), 10))
					w.Header().Add("X-Add-Double", strconv.FormatInt(2*parseInt(originalHeader.Get("X-Add-Double")), 10))
				})
			}),
			args: args{http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Set-Double", "2")
				w.Header().Add("X-Add-Double", "3")
			})},
			want: want{
				code: 200,
				body: "",
				header: http.Header{
					"X-Set":        {"aaa"},
					"X-Set-Double": {"4"},
					"X-Add-Double": {"3", "6"},
				},
			},
		},
		{
			name: "Middleware that allways send 403 Forbidden",
			mf: MiddlewareFunc(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(403)
				})
			}),
			args: args{http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Cogito ergo sum"))
			})},
			want: want{
				code:   403,
				body:   "",
				header: http.Header{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			tt.mf.Handle(tt.args.next).ServeHTTP(rec, r)
			responseHeader := rec.Header().Clone()
			responseHeader.Del("Content-Type")
			got := want{
				code:   rec.Code,
				body:   rec.Body.String(),
				header: responseHeader,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MiddlewareFunc.Handle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMiddlewareFuncEasy_Handle(t *testing.T) {
	type args struct {
		next http.Handler
	}
	type want struct {
		code   int
		body   string
		header http.Header
	}
	tests := []struct {
		name string
		mf   Middleware
		args args
		want want
	}{
		{
			name: "Middleware that add request header",
			mf: MiddlewareFuncEasy(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
				r.Header.Set("X-Foo", "unko")
				next.ServeHTTP(w, r)
			}),
			args: args{http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(r.Header.Get("X-Foo")))
			})},
			want: want{
				code:   200,
				body:   "unko",
				header: http.Header{},
			},
		},
		{
			name: "Middleware that set response header",
			mf: MiddlewareFuncEasy(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
				next.ServeHTTP(w, r)
				originalHeader := w.Header().Clone()
				parseInt := func(s string) int64 {
					i, _ := strconv.ParseInt(s, 10, 64)
					return i
				}
				w.Header().Set("X-Set", "aaa")
				w.Header().Set("X-Set-Double", strconv.FormatInt(2*parseInt(originalHeader.Get("X-Set-Double")), 10))
				w.Header().Add("X-Add-Double", strconv.FormatInt(2*parseInt(originalHeader.Get("X-Add-Double")), 10))
			}),
			args: args{http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Set-Double", "2")
				w.Header().Add("X-Add-Double", "3")
			})},
			want: want{
				code: 200,
				body: "",
				header: http.Header{
					"X-Set":        {"aaa"},
					"X-Set-Double": {"4"},
					"X-Add-Double": {"3", "6"},
				},
			},
		},
		{
			name: "Middleware that allways send 403 Forbidden",
			mf: MiddlewareFuncEasy(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
				w.WriteHeader(403)
			}),
			args: args{http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Cogito ergo sum"))
			})},
			want: want{
				code:   403,
				body:   "",
				header: http.Header{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			tt.mf.Handle(tt.args.next).ServeHTTP(rec, r)
			responseHeader := rec.Header().Clone()
			responseHeader.Del("Content-Type")
			got := want{
				code:   rec.Code,
				body:   rec.Body.String(),
				header: responseHeader,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MiddlewareFunc.Handle() = %v, want %v", got, tt.want)
			}
		})
	}
}
