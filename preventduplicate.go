package zunproxy

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/kawaz/go-requestid"
)

// NewDuplicatePreventer 重複する同時リクエストは一つだけバックエンドに流してレスポンスをシェアすることで高速化を図るミドルウェア
// リクエストの一意性はデフォルトの RequestIDGenerator を利用する
func NewDuplicatePreventerDefault() *DuplicatePreventer {
	return NewDuplicatePreventer(nil)
}

// NewDuplicatePreventer 重複する同時リクエストは一つだけバックエンドに流してレスポンスをシェアすることで高速化を図るミドルウェア
// idgen でリクエストの一意性を調整する
func NewDuplicatePreventer(idgen *requestid.RequestIDGenerator) *DuplicatePreventer {
	dp := &DuplicatePreventer{
		idgen: idgen,
		dw:    map[requestid.RequestID]*DuplicateWriter{},
		dwM:   sync.Mutex{},
	}
	return dp
}

func (dp *DuplicatePreventer) Handler(next http.Handler) http.Handler {
	if dp.idgen == nil {
		dp.idgen = requestid.NewDefaultRequestIDGeneratorConfig().NewGenerator()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID, ok := dp.idgen.GenerateID(r)
		if !ok {
			// reqID が取得できなかったリクエストは対象外なので何もせず次のハンドラに投げて終わる
			next.ServeHTTP(w, r)
			return
		}
		dw, done, first := dp.Register(w, reqID)
		if first {
			// 同時リクエストの最初のリクエストが代表して dw を次のハンドラに投げる
			go func() {
				defer dw.Done()
				next.ServeHTTP(dw, r)
			}()
		}
		// 代表リクエストのレスポンスが返ってくるのを待つ
		<-done
	})
}

// DuplicatePreventer is middleware
type DuplicatePreventer struct {
	idgen  *requestid.RequestIDGenerator
	dwSync sync.Map
	dw     map[requestid.RequestID]*DuplicateWriter
	dwM    sync.Mutex
}

// Register http.ResponseWriter, get *DuplicateWriter
func (dp *DuplicatePreventer) Register(w http.ResponseWriter, reqID requestid.RequestID) (dw *DuplicateWriter, done <-chan struct{}, first bool) {
	dw, first = dp.getDuplicateWriter(reqID)
	doneRW := make(chan struct{})
	done = doneRW
	dw.writersM.Lock()
	defer dw.writersM.Unlock()
	dw.writers = append(dw.writers, &waitingWriter{w, doneRW})
	return
}

func (dp *DuplicatePreventer) getDuplicateWriter(reqID requestid.RequestID) (dw *DuplicateWriter, first bool) {
	dp.dwM.Lock()
	defer dp.dwM.Unlock()
	dw, found := dp.dw[reqID]
	first = !found
	if first {
		dw = &DuplicateWriter{
			dp:              dp,
			reqID:           reqID,
			container:       &ResponseContainer{statusCode: 0, header: http.Header{}, body: bytes.Buffer{}},
			writers:         []*waitingWriter{},
			writersM:        sync.Mutex{},
			responseWritten: make(chan struct{}),
		}
		dp.dw[reqID] = dw
	}
	return
}

// DuplicateWriter は複数の *http.ResponseWriter へ一つのレスポンスの内容を複製して書き込みます
type DuplicateWriter struct {
	dp              *DuplicatePreventer
	reqID           requestid.RequestID
	container       *ResponseContainer
	writers         []*waitingWriter
	writersM        sync.Mutex
	responseWritten chan struct{}
}

var _ http.ResponseWriter = &DuplicateWriter{}      // Verify that T implements I.
var _ http.ResponseWriter = (*DuplicateWriter)(nil) // Verify that *T implements I.

// Header implements http.ResponseWriter
func (dw *DuplicateWriter) Header() http.Header {
	return dw.container.header
}

// WriteHeader implements http.ResponseWriter
func (dw *DuplicateWriter) WriteHeader(statusCode int) {
	dw.container.statusCode = statusCode
}

// Write implements http.ResponseWriter
func (dw *DuplicateWriter) Write(chunk []byte) (int, error) {
	return dw.container.body.Write(chunk)
}

// Done dw へのレスポンスが完了したら呼んでもらう
func (dw *DuplicateWriter) Done() {
	// dw へのリクエスト登録の受付を終了させる
	func() {
		dw.dp.dwM.Lock()
		defer dw.dp.dwM.Unlock()
		delete(dw.dp.dw, dw.reqID)
	}()
	// dw で受けたレスポンスを本来の ResponseWriter 達にコピーする
	for _, ww := range dw.writers {
		// ResponseWriter の相手毎に回線速度差とかあるので並列にコピって個別に開放する
		go func(ww *waitingWriter) {
			dw.container.WriteTo(ww.w)
			// HundlerFunc で待たせてる処理を再開させる
			close(ww.done)
		}(ww)
	}
}

// ResponseContainer レスポンスの一時保管場所
type ResponseContainer struct {
	statusCode int
	header     http.Header
	body       bytes.Buffer
}

// WriteTo http.ResponseWriter
func (rc *ResponseContainer) WriteTo(w http.ResponseWriter) (int, error) {
	header := w.Header()
	for k, values := range rc.header {
		for _, v := range values {
			header.Add(k, v)
		}
	}
	w.WriteHeader(rc.statusCode)
	return w.Write(rc.body.Bytes())
}

// waitingWriter make set http.ResponseWriter と終了用チャンネルをセットにする入れ物
type waitingWriter struct {
	w    http.ResponseWriter
	done chan struct{}
}
