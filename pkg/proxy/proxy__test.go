package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:8701"), http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			data, _ := io.ReadAll(request.Body)
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"method":  request.Method,
				"url":     request.URL.String(),
				"headers": request.Header,
				"body":    string(data),
			})
		}))
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:8700"), &Proxy{})
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	t.Run("PROXY GET", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "http://0.0.0.0:8700/http://0.0.0.0:8701/a/b", nil)
		r.Header.Add("X-Proxy-Forward-Origin", "github.com")
		resp, _ := http.DefaultClient.Do(r)
		data, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(data))
	})

	t.Run("PROXY POST", func(t *testing.T) {
		r, _ := http.NewRequest("POST", "http://0.0.0.0:8700/http://0.0.0.0:8701/a/b", bytes.NewBufferString("{}"))
		r.Header.Add("X-Proxy-Forward-Origin", "github.com")
		resp, _ := http.DefaultClient.Do(r)
		data, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(data))
	})
}
