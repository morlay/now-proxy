package proxy

import (
	"cmp"
	"fmt"
	"io"
	"log/slog"
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

	rng, err := ParseRange(r.Header.Get("Range"), MaxResponseContentLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p.replyFrom(w, req, rng)
}

func (p *Proxy) replyFrom(w http.ResponseWriter, r *http.Request, rng *Range) {
	if r.Method == http.MethodOptions {
		header := w.Header()

		header.Set("Access-Control-Allow-Method", r.Header.Get("Access-Control-Allow-Method"))
		header.Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Allow-Headers"))
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Max-Age", "3600")

		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write(nil)
		return
	}

	l := slog.Default().With(
		slog.String("http.request.method", r.Method),
	)
	for k := range r.Header {
		if strings.HasPrefix(k, "Range") || strings.HasPrefix(k, "Content-") {
			l = l.With(slog.String("http.request.header."+k, r.Header.Get(k)))
		}
	}

	if rng != nil {
		r.Header.Set("Range", rng.RangeString())
	}

	client := createClient(10 * time.Second)
	resp, err := client.Do(r)
	if err != nil {
		l.Error(err.Error())
		p.writeErr(w, http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	l = l.With(slog.Int("http.response.status", resp.StatusCode))
	for k := range resp.Header {
		if strings.HasPrefix(k, "Range") || strings.HasPrefix(k, "Content-") {
			l = l.With(slog.String("http.response.header."+k, resp.Header.Get(k)))
		}
	}

	l.Info("requested")

	for k, vv := range resp.Header {
		w.Header()[k] = vv
	}

	// force enable CORS
	w.Header().Set("Access-Control-Allow-Method", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// force no-cache
	w.Header().Set("Cache-Control", "no-cache")

	// force partial-content
	if r.Method == http.MethodGet {
		if resp.ContentLength > MaxResponseContentLength {
			if rng == nil {
				rng = &Range{Start: 0}
			}
			rng.Length = MaxResponseContentLength

			w.Header().Set("Content-Range", rng.ContentRange(cmp.Or(ContentRangeTotal(resp.Header.Get("Content-Range")), resp.ContentLength)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", rng.Length))

			w.WriteHeader(http.StatusPartialContent)

			if _, err := io.Copy(w, io.LimitReader(resp.Body, MaxResponseContentLength)); err != nil {
				l.Error(err.Error())
			}
		}
		return
	}

	w.WriteHeader(resp.StatusCode)

	if resp.StatusCode != http.StatusNoContent {
		if _, err := io.Copy(w, resp.Body); err != nil {
			l.Error(err.Error())
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
