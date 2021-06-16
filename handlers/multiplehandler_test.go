package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMultipleHandler(t *testing.T) {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("h")) })
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
			args: args{
				h:  h,
				ms: []Middleware{},
			},
			want: "h",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			mh := MultipleHandler(tt.args.h, tt.args.ms...)
			mh.ServeHTTP(rec, req)
			got := string(rec.Body.Bytes())
			if got != tt.want {
				t.Errorf("MultipleHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
