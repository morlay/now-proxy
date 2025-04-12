package proxy

import (
	"encoding/json"
	"log/slog"
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

	if err := json.NewEncoder(w).Encode(&Err{Msg: err.Error()}); err != nil {
		slog.Default().Error(err.Error())
	}
}

type Err struct {
	Msg string `json:"msg"`
}
