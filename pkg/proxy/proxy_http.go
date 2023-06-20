package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/vincent-petithory/dataurl"
)

const MaxResponseContentLength = 4 * MiB

func (p *Proxy) HandleHttp(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.RequestURI, "/data:") {
		dataURL, err := dataurl.DecodeString(r.RequestURI[1:])
		if err != nil {
			p.writeErr(w, http.StatusBadRequest, err)
			return
		}

		w.Header().Set("Content-Type", dataURL.MediaType.String())
		w.WriteHeader(200)
		_, _ = w.Write(dataURL.Data)

		return
	}

	req, err := processRequest(r)
	if err != nil {
		p.writeErr(w, http.StatusBadRequest, err)
		return
	}

	rr, err := ParseRange(r.Header.Get("Range"), MaxResponseContentLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rr != nil {
		req.Header.Set("Range", rr.RangeString())
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

	client := createClient(10 * time.Second)
	resp, err := client.Do(r)
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

	// force no-cache
	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == http.MethodGet && resp.StatusCode != http.StatusPartialContent && resp.ContentLength > MaxResponseContentLength {
		w.Header().Set("Content-Range", (&Range{
			Start:  0,
			Length: MaxResponseContentLength,
		}).ContentRange(resp.ContentLength))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", MaxResponseContentLength))
		w.WriteHeader(http.StatusPartialContent)

		if _, err := io.Copy(w, io.LimitReader(resp.Body, MaxResponseContentLength)); err != nil {
			fmt.Println(err)
		}
		return
	}

	w.WriteHeader(resp.StatusCode)
	if resp.StatusCode != http.StatusNoContent {
		if _, err := io.Copy(w, resp.Body); err != nil {
			fmt.Println(err)
		}
	}
}

func createClient(timeout time.Duration) *http.Client {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 0,
		}).DialContext,
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: t,
	}
}

const ForwardPrefix = "X-Proxy-Forward-"
const HeaderInQuery = "X-Param-Header-"

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

	deferHeaders := http.Header{}

	for k, vv := range r.Header {
		if strings.HasPrefix(k, ForwardPrefix) {
			deferHeaders[k[len(ForwardPrefix):]] = vv
		} else {
			req.Header[k] = vv
		}
	}

	query := req.URL.Query()

	for k := range query {
		if strings.HasPrefix(textproto.CanonicalMIMEHeaderKey(k), HeaderInQuery) {
			req.Header[textproto.CanonicalMIMEHeaderKey(k[len(HeaderInQuery):])] = query[k]
			query.Del(HeaderInQuery)
		}
	}

	r.URL.RawQuery = query.Encode()

	for k, vv := range deferHeaders {
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
