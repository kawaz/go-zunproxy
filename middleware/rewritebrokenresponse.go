package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"log"

	"github.com/andybalholm/brotli"
)

func NewBrokenRewriteGuard() Middleware {
	return &BrokenRewriteGuardHandler{}
}

type BrokenRewriteGuardHandler struct {
}

func (rewrite *BrokenRewriteGuardHandler) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewResponseSteeler()
		buf := bytes.NewBuffer([]byte{})
		rec.AddWriter(buf)
		next.ServeHTTP(rec, r)
		code := rec.Code()
		TE := rec.Header().Get("Transfer-Encoding")
		broken := false
		if code == http.StatusOK {
			if strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
				var reader io.Reader
				var err error
				switch TE {
				case "gzip":
					reader, err = gzip.NewReader(bytes.NewBuffer(buf.Bytes()))
				case "br":
					reader = brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
				default:
					reader = bytes.NewBuffer(buf.Bytes())
				}
				if err != nil {
					log.Printf("BrokenRewriteGuardHandler: %v", err)
					broken = true
				}
				plain, err := ioutil.ReadAll(reader)
				if err != nil {
					log.Printf("BrokenRewriteGuardHandler: %v", err)
					broken = true
				}
				html := string(plain)
				if strings.Contains(html, "<html") && !strings.Contains(html, "</html>") {
					broken = true
					log.Printf("ERROR BrokenRewriteGuardHandler: no </html>: %v", r.URL)
				}
				// if strings.Contains(html, "\uFFFD") {
				// 	broken = true
				// 	log.Printf("ERROR BrokenRewriteGuardHandler: found \\uFFFD: %v", r.URL)
				// }
				// log.Printf("TE=%v plain=%v, buflen=%v, plainlen=%v, broken=%v", TE, string(plain), buf.Len(), len(plain), broken)
			}
			if broken {
				buf.Reset()
				TE = ""
				reloadHTML := `<!DOCTYPE html><html><head><meta charset="utf-8"><script>setTimeout(function(){location.reload()}, 5000)</script></head><body>Server Error. Reload after 5 seconds...</body></html>\n`
				buf.WriteString(reloadHTML)
				code = http.StatusInternalServerError
			}
		}
		// 取得したレスポンスを書き出す
		if TE == "" {
			rec.Header().Del("Content-Encoding")
		}
		rec.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		for k, values := range rec.Header() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(code)
		w.Write(buf.Bytes())
	})
}
