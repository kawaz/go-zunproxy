package middleware

import (
	"io"
	"net/http"
)

type ResponseRecorder interface {
	http.ResponseWriter
	Code() int
	ContentLength() int
	AddWriteHeaderListener(func(code int, header http.Header))
}

type responseRecorder struct {
	w          http.ResponseWriter
	bw         io.Writer
	code       int
	clen       int
	listenerWH []func(code int, header http.Header)
}

var _ ResponseRecorder = (*responseRecorder)(nil)
var _ http.ResponseWriter = (*responseRecorder)(nil)

func NewResponseRecorder(w http.ResponseWriter, bodyWriter io.Writer) ResponseRecorder {
	if bodyWriter == nil {
		return &responseRecorder{
			w:  w,
			bw: w,
		}
	}
	return &responseRecorder{
		w:  w,
		bw: io.MultiWriter(w, bodyWriter),
	}
}

func (rec *responseRecorder) Write(p []byte) (n int, err error) {
	n, err = rec.bw.Write(p)
	rec.clen += n
	return n, err
}

func (rec *responseRecorder) WriteHeader(code int) {
	if len(rec.listenerWH) != 0 {
		for _, listener := range rec.listenerWH {
			listener(code, rec.Header().Clone())
		}
	}
	rec.code = code
	rec.w.WriteHeader(code)
}

func (rec *responseRecorder) Header() http.Header {
	return rec.w.Header()
}

func (rec *responseRecorder) Code() int {
	return rec.code
}

func (rec *responseRecorder) ContentLength() int {
	return rec.clen
}

func (rec *responseRecorder) AddWriteHeaderListener(listener func(code int, header http.Header)) {
	rec.listenerWH = append(rec.listenerWH, listener)
}
