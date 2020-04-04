package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleTunneling(w, r)
		return
	}

	handleHttp(w, r)
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	req, err := processRequest(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	replyFrom(w, req)
}

func replyFrom(w http.ResponseWriter, r *http.Request) {
	c := getShortConnClient(10 * time.Second)

	resp, err := c.Do(r)

	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	defer resp.Body.Close()

	for k, vv := range resp.Header {
		w.Header()[k] = vv
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
	}
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		writeErr(w, http.StatusInternalServerError, errors.New("Hijacking not supported"))
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

type Err struct {
	Msg string
}

func writeErr(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&Err{Msg: err.Error()})
}

func getShortConnClient(timeout time.Duration) *http.Client {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 0,
		}).DialContext,
		DisableKeepAlives: true,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: t,
	}
}

func processRequest(r *http.Request) (*http.Request, error) {
	if r.Header.Get("Proxy-Connection") != "" {
		r.Header.Del("Proxy-Connection")
	} else {
		path := r.URL.Path

		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		if u2, err := url.Parse(patchURI(path)); err != nil {
			return nil, err
		} else {
			r.URL.Scheme = u2.Scheme
			r.URL.Host = u2.Host
			r.URL.Path = u2.Path
		}
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	for k, vv := range r.Header {
		req.Header[k] = vv
	}

	req.Header.Set("Host", r.Host)

	return req, nil
}

var re = regexp.MustCompile(`^(https?:)\/{1,2}`)

func patchURI(p string) string {
	return re.ReplaceAllStringFunc(p, func(s string) string {
		return re.FindStringSubmatch(s)[1] + "//"
	})
}
