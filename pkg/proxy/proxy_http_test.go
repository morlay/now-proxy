package proxy

import (
	"net/http"
	"testing"
)

func Test_patchURI(t *testing.T) {
	t.Log(patchURI("http:/google.com"))
	t.Log(patchURI("https:/google.com"))
	t.Log(patchURI("http://google.com"))
	t.Log(patchURI("https://google.com"))
}

func Test_processRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Proxy-Forward-User-Agent", "A")
	req.Header.Add("X-Proxy-Forward-Origin", "a.com")
	req.Header.Add("Cookie", "aaa")

	r, _ := processRequest(req)
	t.Log(r.Header)
}
