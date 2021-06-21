package middleware

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/itchyny/timefmt-go"
	"github.com/oklog/ulid"
)

var textTypes = NewWildCardsOr("text/*", "application/json*", "application/javascript*")

type dumpHandler struct {
	DumpDir      string
	TruncateSize int
	rng          io.Reader
}

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

func NewDumpHandler(dumpDir string) Middleware {
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
		dumpDir := timefmt.Format(tsStart, dh.DumpDir)
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
		// Body保存用のファイルのCloser
		var bodyC io.Closer
		defer Close(bodyC)
		rec := NewResponseRecorder(w)
		rec.AddWriteHeaderListener(func(code int, header http.Header) {
			// text のみ保存する
			t := header.Get("Content-Type")
			if !textTypes.Match(t) {
				return
			}
			f, err := CreateFile(dumpDir, dump.ID+".body")
			if err != nil {
				log.Printf("could not create dump body: %v", err)
			}
			bodyC = f
			rec.AddWriter(f)
		})
		// 次の Handler を実行
		next.ServeHTTP(rec, r)
		if bodyC == nil {
			// ダンプ対象じゃ無いので何もしない
			return
		}
		// メタ情報を保存
		dump.Duration = time.Since(dump.Ts.UTC())
		dump.Response = &dumpRes{
			Code:          rec.Code(),
			ContentLength: rec.ContentLength(),
			Header:        rec.Header().Clone(),
		}
		jsonFile, err := CreateFile(dumpDir, dump.ID+".json")
		if err != nil {
			log.Printf("could not create dump json: %v", err)
			return
		}
		defer jsonFile.Close()
		jsonBytes, err := json.Marshal(dump)
		if err != nil {
			log.Printf("could not marshal dump json: %v", err)
			return
		}
		_, err = jsonFile.Write(jsonBytes)
		if err != nil {
			log.Printf("could not write dump json: %v", err)
		}
	})
}
