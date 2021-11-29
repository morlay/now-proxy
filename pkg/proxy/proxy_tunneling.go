package proxy

import (
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

func (p *Proxy) HandleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	if err != nil {
		p.writeErr(w, http.StatusServiceUnavailable, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		p.writeErr(w, http.StatusInternalServerError, errors.New("Hijacking not supported"))
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
