package proxy

import (
	"encoding/json"
	"net/http"
)

type Proxy struct {
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.HandleTunneling(w, r)
		return
	}
	p.HandleHttp(w, r)
}

func (Proxy) writeErr(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&Err{Msg: err.Error()})
}

type Err struct {
	Msg string
}
