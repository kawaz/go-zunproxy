package extension

import "net/http"

// Extension do nothing
type NopExtension struct{}

var _ Extention = &NopExtension{}
var _ HttpExtention = &NopExtension{}

func (ex *NopExtension) Config() interface{} { return nil }
func (ex *NopExtension) Init() error         { return nil }
func (ex *NopExtension) Dispose() error      { return nil }
func (ex *NopExtension) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
