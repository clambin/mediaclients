package testutil

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ClientLogger(w io.Writer, next http.RoundTripper) http.RoundTripper {
	return loggingRoundTripper{w: cmp.Or(w, io.Writer(os.Stdout)), next: cmp.Or(next, http.DefaultTransport)}
}

type loggingRoundTripper struct {
	w    io.Writer
	next http.RoundTripper
}

func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	_, _ = fmt.Fprintf(l.w, "%s %s\n", req.Method, req.URL.String())
	l.dumpHeader(req.Header)
	req.Body = l.dumpBody(req.Body)
	_, _ = fmt.Fprintln(l.w, "--------")
	defer func() { _, _ = fmt.Fprintln(l.w, "\n========") }()

	resp, err := l.next.RoundTrip(req)

	if err != nil {
		_, _ = fmt.Fprintf(l.w, "%v\n", err)
		return nil, err
	}
	_, _ = fmt.Fprintf(l.w, "%s\n", cmp.Or(resp.Status, http.StatusText(resp.StatusCode)))
	l.dumpHeader(resp.Header)
	resp.Body = l.dumpBody(resp.Body)

	return resp, nil
}

func (l loggingRoundTripper) dumpHeader(h http.Header) {
	for k, v := range h {
		_, _ = fmt.Fprintf(l.w, "%s: %v\n", k, v)
	}
}

func (l loggingRoundTripper) dumpBody(r io.ReadCloser) io.ReadCloser {
	_, _ = fmt.Fprintln(l.w)
	if r == nil {
		return nil
	}
	var origBody bytes.Buffer
	r2 := io.TeeReader(r, &origBody)
	_, _ = io.Copy(l.w, r2)
	/*
		body, _ := io.ReadAll(r2)
		for line := range bytes.Lines(body) {
			if bytes.HasSuffix(line, []byte("\n")) {
				line = line[:len(line)-1]
			}
			fmt.Println("", string(line))
		}
	*/
	_ = r.Close()
	return io.NopCloser(&origBody)
}
