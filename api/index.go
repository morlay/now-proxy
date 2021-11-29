package handler

import (
	"github.com/morlay/now-proxy/pkg/proxy"
	"net/http"
)

var p = &proxy.Proxy{}

func Handler(w http.ResponseWriter, r *http.Request) {
	p.ServeHTTP(w, r)
}
