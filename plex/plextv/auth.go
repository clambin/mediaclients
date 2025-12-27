package plextv

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	// DefaultConfig contains the default configuration required to authenticate with Plex.
	defaultHTTPClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: http.DefaultTransport,
	}
)

type httpClientType struct{}

// ContextWithHTTPClient returns a new context with an added HTTP client. When passed to [Config]'s methods,
// they use that HTTP client to perform their authentication calls.
// If no HTTP client is set, a default HTTP client is used.
func ContextWithHTTPClient(ctx context.Context, httpClient *http.Client) context.Context {
	return context.WithValue(ctx, httpClientType{}, httpClient)
}

// httpClient returns the HTTP set in the context. If none is set, it returns a default client.
func httpClient(ctx context.Context) *http.Client {
	if c, ok := ctx.Value(httpClientType{}).(*http.Client); ok {
		return c
	}
	return defaultHTTPClient
}

// requestFormatter modifies a request before [Config.do] sends to its destination.
type requestFormatter func(*http.Request)

// do builds a new HTTP request and sends it to the destination URL.
// note: url cannot include query parameters
func (c Config) do(ctx context.Context, method string, url string, body io.Reader, wantStatusCode int, formatters ...requestFormatter) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("X-Plex-Client-Identifier", c.ClientID)
	c.Device.populateRequest(req)
	for _, formatter := range formatters {
		formatter(req)
	}
	resp, err := httpClient(ctx).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != wantStatusCode {
		defer func() { _ = resp.Body.Close() }()
		return nil, ParsePlexError(resp)
	}
	return resp, nil
}
