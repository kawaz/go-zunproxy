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
	AddWriter(io.Writer)
}

type responseRecorder struct {
	// original writer
	w http.ResponseWriter
	// additional writer
	ws []io.Writer
	// multiple writer
	mw         io.Writer
	header     http.Header
	code       int
	clen       int
	listenerWH []func(code int, header http.Header)
}

var _ ResponseRecorder = (*responseRecorder)(nil)
var _ http.ResponseWriter = (*responseRecorder)(nil)

func NewResponseRecorder(w http.ResponseWriter) ResponseRecorder {
	return &responseRecorder{
		w:  w,
		ws: []io.Writer{w},
	}
}
func NewResponseSteeler() ResponseRecorder {
	return &responseRecorder{
		ws: []io.Writer{},
	}
}

func (rec *responseRecorder) Header() http.Header {
	if rec.w != nil {
		return rec.w.Header()
	}
	if rec.header == nil {
		rec.header = http.Header{}
	}
	return rec.header
}

func (rec *responseRecorder) WriteHeader(code int) {
	if len(rec.listenerWH) != 0 {
		for _, listener := range rec.listenerWH {
			listener(code, rec.Header().Clone())
		}
	}
	rec.code = code
	if rec.w != nil {
		rec.w.WriteHeader(code)
	}
	rec.mw = io.MultiWriter(rec.ws...)
}

func (rec *responseRecorder) Write(p []byte) (n int, err error) {
	n, err = rec.mw.Write(p)
	rec.clen += n
	return n, err
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

func (rec *responseRecorder) AddWriter(w io.Writer) {
	if w != nil {
		rec.ws = append(rec.ws, w)
	}
}

type responseSteeler struct {
	responseRecorder
}
