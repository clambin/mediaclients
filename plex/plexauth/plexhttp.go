package plexauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// HTTPClient returns the HTTP set in the context. If none is set, it returns a default client.
func HTTPClient(ctx context.Context) *http.Client {
	if c, ok := ctx.Value(httpClientType{}).(*http.Client); ok {
		return c
	}
	return defaultHTTPClient
}

func ParsePlexError(r *http.Response) error {
	var (
		errorBody struct {
			Error  string `json:"error"`
			Errors []struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
				Status  int    `json:"status"`
			} `json:"errors"`
		}
		buf bytes.Buffer
	)

	var reason string
	if r.Body != nil {
		_ = json.NewDecoder(io.TeeReader(r.Body, &buf)).Decode(&errorBody)
	}

	switch {
	case errorBody.Error != "":
		// single error message
		reason = errorBody.Error
	case len(errorBody.Errors) > 0:
		// multi-message error
		messages := make([]string, len(errorBody.Errors))
		for i, entry := range errorBody.Errors {
			messages[i] = fmt.Sprintf("%d - %s", entry.Code, entry.Message)
		}
		reason = strings.Join(messages, ", ")
	default:
		reason = r.Status
	}

	return &PlexError{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Reason:     reason,
		Body:       buf.Bytes()}
}
