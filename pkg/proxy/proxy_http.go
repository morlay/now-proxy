package proxy

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func (p *Proxy) HandleHttp(w http.ResponseWriter, r *http.Request) {
	req, err := processRequest(r)
	if err != nil {
		p.writeErr(w, http.StatusBadRequest, err)
		return
	}
	p.replyFrom(w, req)
}

func (p *Proxy) replyFrom(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		header := w.Header()

		header.Set("access-control-allow-method", r.Header.Get("access-control-request-method"))
		header.Set("access-control-allow-headers", r.Header.Get("access-control-request-headers"))
		header.Set("access-control-allow-credentials", "true")
		header.Set("access-control-allow-origin", "*")
		header.Set("access-control-max-age", "3600")

		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write(nil)
		return
	}

	c := getShortConnClient(10 * time.Second)

	resp, err := c.Do(r)

	if err != nil {
		p.writeErr(w, http.StatusInternalServerError, err)
		return
	}

	defer resp.Body.Close()

	for k, vv := range resp.Header {
		w.Header()[k] = vv
	}

	// force enable CORS
	w.Header().Set("Access-Control-Allow-Method", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		p.writeErr(w, http.StatusInternalServerError, err)
	}
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

const ForwardPrefix = "X-Proxy-Forward-"

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
		if strings.HasPrefix(k, ForwardPrefix) {
			req.Header[k[len(ForwardPrefix):]] = vv
		} else {
			req.Header[k] = vv
		}
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
