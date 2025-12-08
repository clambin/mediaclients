package plexauth

import (
	"context"
	"net/http"
	"time"
)

var defaultHTTPClient = &http.Client{
	Timeout:   15 * time.Second,
	Transport: http.DefaultTransport,
}

type httpClientType struct{}

// WithHTTPClient returns a new context with an added HTTP client.
func WithHTTPClient(ctx context.Context, httpClient *http.Client) context.Context {
	return context.WithValue(ctx, httpClientType{}, httpClient)
}

// httpClient returns the HTTP set in the context. If none is set, it returns a default client.
func httpClient(ctx context.Context) *http.Client {
	if c, ok := ctx.Value(httpClientType{}).(*http.Client); ok {
		return c
	}
	return defaultHTTPClient
}
