package middleware

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	var numRequests = 30

	var hits uint32
	var expectedStatus int = 200
	var expectedBody = []byte("ほげ")

	app := func(w http.ResponseWriter, r *http.Request) {
		log.Println("app handler..")

		atomic.AddUint32(&hits, 1)

		hitsNow := atomic.LoadUint32(&hits)
		if hitsNow > 1 {
			panic("uh oh")
		}

		time.Sleep(100 * time.Millisecond) // slow handler
		w.Header().Set("X-Httpjoin", "test")
		w.WriteHeader(expectedStatus)
		w.Write(expectedBody)
	}

	var count uint32
	counter := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint32(&count, 1)
			next.ServeHTTP(w, r)
			atomic.AddUint32(&count, ^uint32(0))
			// log.Println("COUNT:", atomic.LoadUint32(&count))
		})
	}

	recoverer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("recovered panicing request:", r)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
	// dpm := NewDuplicatePreventer(nil)
	dpm := NewRequestBundler(nil)
	ts := httptest.NewServer(counter(recoverer(dpm.Handle(http.HandlerFunc(app)))))
	defer ts.Close()

	var wg sync.WaitGroup

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, _ := http.Get(ts.URL)
			body, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()

			// log.Println("got resp:", resp, "len:", len(body), "body:", string(body))

			if string(body) != string(expectedBody) {
				t.Error("expecting response body:", string(expectedBody))
			}

			if resp.StatusCode != expectedStatus {
				t.Error("expecting response status:", expectedStatus)
			}

			if resp.Header.Get("X-Httpjoin") != "test" {
				t.Error("expecting x-httpjoin test header")
			}

		}()
	}

	wg.Wait()

	totalHits := atomic.LoadUint32(&hits)
	if totalHits > 1 {
		t.Error("handler was hit more than once. hits:", totalHits)
	}
	// log.Println("total hits:", totalHits)

	finalCount := atomic.LoadUint32(&count)
	if finalCount > 0 {
		t.Error("queue count was expected to be empty, but count:", finalCount)
	}
	// log.Println("final count:", finalCount)
}
