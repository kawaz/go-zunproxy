package handlers

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/goccy/go-json"
	"github.com/oklog/ulid"
)

type DumpHandler interface {
	Handle(next http.Handler) http.Handler
}

type dumpHandler struct {
	DumpDir      string
	TruncateSize int
	rng          io.Reader
}

var _ DumpHandler = (*dumpHandler)(nil)

type dumpContent struct {
	ID       string
	Ts       time.Time
	Duration time.Duration
	Request  dumpReq
	Response *dumpRes
}

type dumpReq struct {
	Method   string
	Path     string
	Query    url.Values
	RawPath  string
	RawQuery string
	Header   http.Header
}
type dumpRes struct {
	Code          int
	ContentLength int
	Header        http.Header
	Truncated     bool
}

func NewDumpHandler(dumpDir string) DumpHandler {
	return &dumpHandler{
		DumpDir: dumpDir,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
func (dh *dumpHandler) Handle(next http.Handler) http.Handler {
	if dh.DumpDir == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストの記録開始
		tsStart := time.Now()
		dump := &dumpContent{
			ID: ulid.MustNew(ulid.Timestamp(tsStart), dh.rng).String(),
			Ts: tsStart,
			Request: dumpReq{
				Method:   r.Method,
				Path:     r.URL.Path,
				Query:    r.URL.Query(),
				RawPath:  r.URL.RawPath,
				RawQuery: r.URL.RawQuery,
				Header:   r.Header.Clone(),
			},
		}
		// Body保存用のWriterを用意
		var bodyWriter io.Writer
		bodyFile, err := os.Create(path.Join(dh.DumpDir, dump.ID) + ".body")
		if err != nil {
			log.Printf("could not create dump body: %v", err)
			bodyWriter = io.Discard
		} else {
			bodyWriter = bodyFile
			defer bodyFile.Close()
		}
		// 次の Handler を実行
		rec := NewResponseRecorder(w, bodyWriter)
		next.ServeHTTP(rec, r)
		// メタ情報を保存
		dump.Duration = time.Since(dump.Ts.UTC())
		dump.Response = &dumpRes{
			Code:          rec.Code(),
			ContentLength: rec.ContentLength(),
			Header:        rec.Header().Clone(),
		}
		var jsonWriter io.Writer
		jsonFile, err := os.Create(path.Join(dh.DumpDir, dump.ID) + ".json")
		if err != nil {
			log.Printf("could not create dump json: %v", err)
			jsonWriter = io.Discard
		} else {
			defer jsonFile.Close()
			jsonWriter = jsonFile
		}
		jsonBytes, err := json.Marshal(dump)
		if err != nil {
			log.Printf("could not marshal dump json: %v", err)
			return
		}
		_, err = jsonWriter.Write(jsonBytes)
		if err != nil {
			log.Printf("could not write dump json: %v", err)
		}
	})
}
