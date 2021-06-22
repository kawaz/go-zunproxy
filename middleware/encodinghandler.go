package middleware

import (
	"io"
	"net/http"
	"strings"
)

const (
	// uncompressed, raw binary
	EncIdentity = "identity"
	// 1st priority
	EncBrotli = "br"
	// 2nd priority
	EncGzip = "gzip"
	// not support (tiny user, deprecated)
	EncXGzip = "x-gzip"
	// not support (tiny user)
	EncDeflate = "deflate"
	// not support (deprecated)
	EncCompress = "compress"
	// not support (deprecated)
	EncSdch = "sdch"
	// streaming etc...
	EncChunked = "chunked"
)

// br, gzip だけで十分でしょ
var defaultAcceptOrder = []string{EncBrotli, EncGzip}

type BodyEncorder interface {
	Middleware

	// ここで取得したWriterをnextハンドラーに渡す
	// Content-Encoding を見て identity を取り出しつつ3種類のエンコーディングも作成する
	ResponseWriter(w http.ResponseWriter)

	// リクエストに応じた Content-Encoding で返す
	// br, gzip, identity に対応 // deflate, compress は対応しない
	// Content-Encoding, Vary: Content-Encoding, Content-Length
	WriteTo(w http.ResponseWriter, r *http.Request)

	// 非圧縮のボディを取得する
	// If BodyEncoder is not applied to request, return nil.
	// If w.WriteHeader does not proceed, return nil.
	BodyImplicit() EncordedBody

	// Return optimal Content-Type for Accept-Encording
	// If there is no particular match, return "implicit".
	Encoding() string

	// 全Encodingのボディーを取得する
	BodyAll() map[string]EncordedBody
}

type EncordedBody interface {
	Length() int
	Reader() io.Reader
	Encording() string
}

type bodyEncorder struct {
	bodies     []bodyRW
	acceptEncs []string
}

var _ EncordedBody = &bodyEncorder{}

func (be *bodyEncorder) BodyAll() map[string]EncordedBody {
	var m = make(map[string]EncordedBody, len(be.bodies))
	for _, br := range be.bodies {
		m[br.enc] = &br
	}
	return m
}

func (be *bodyEncorder) Encording() string {
	return ""
}
func (be *bodyEncorder) Length() int {
	return 0
}
func (be *bodyEncorder) Reader() io.Reader {
	return nil
}

type bodyRW struct {
	enc string
	len int
	w   io.Writer
	r   io.Reader
}

var _ EncordedBody = &bodyRW{}

func (body *bodyRW) Length() int {
	return body.len
}
func (body *bodyRW) Encording() string {
	return body.enc
}
func (body *bodyRW) Reader() io.Reader {
	return nil
}

func NewBodyEncorder() Middleware {
	return MiddlewareFuncEasy(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		// encoder := bodyEncorder{}
		// ce := GetContentEncoding(r.Header.Get("Accept-Encording"))
		// pw, pw := io.Pipe()
		// // select ce {
		// // case "implicit":

		// // }
		// buf := bytes.NewBuffer(nil)
		// encoder.bodies = append(encoder.bodies, bodyRW{
		// 	w:   gzip.NewWriter(buf),
		// 	r:   buf,
		// 	enc: "gzip",
		// })

	})
}

func GetContentEncodingDefault(acceptEnc string) string {
	return GetContentEncording(acceptEnc, defaultAcceptOrder)
}

func GetContentEncording(acceptEnc string, acceptOrder []string) string {
	for _, ce := range acceptOrder {
		ae := acceptEnc[:]
		for {
			i := strings.Index(ae, ce)
			if i == -1 {
				break
			}
			if 0 < i && 0 <= strings.IndexByte(" ,;", ae[i-1]) {
				ae = ae[i:]
			}
			if len(ae) == len(ce) {
				return ce
			}
			if 0 <= strings.IndexByte(" ,;", ae[len(ce)]) {
				return ce
			}
			ae = ae[len(ce)+1:]
		}
	}
	return EncIdentity
}
