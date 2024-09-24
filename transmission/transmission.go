package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Client calls the Transmission APIs
type Client struct {
	HTTPClient *http.Client
	URL        string
}

// NewClient returns a new Transmission Client
//
// deprecated
func NewClient(url string, roundTripper http.RoundTripper) *Client {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &Client{
		HTTPClient: &http.Client{Transport: &authenticator{next: roundTripper}},
		URL:        url,
	}
}

// GetSessionParameters calls Transmission's "session-get" method. It returns the Transmission instance's configuration parameters
func (c *Client) GetSessionParameters(ctx context.Context) (SessionParameters, error) {
	params, err := post[SessionParameters](ctx, c, "session-get")
	if err == nil && params.Result != "success" {
		err = fmt.Errorf("session-get failed: %s", params.Result)
	}
	return params, err
}

// GetSessionStatistics calls Transmission's "session-stats" method. It returns the Transmission instance's session statistics
func (c *Client) GetSessionStatistics(ctx context.Context) (SessionStats, error) {
	stats, err := post[SessionStats](ctx, c, "session-stats")
	if err == nil && stats.Result != "success" {
		err = fmt.Errorf("session-stats failed: %s", stats.Result)
	}
	return stats, err
}

func post[T any](ctx context.Context, client *Client, method string) (T, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, client.URL, bytes.NewBufferString(`{ "method": "`+method+`" }`))
	req.Header.Set("Content-Type", "application/json")

	var response T
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return response, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return response, errors.New(resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}
	return response, err
}
