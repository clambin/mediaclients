package plex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Client calls the Plex APIs
type Client struct {
	URL        string
	HTTPClient *http.Client
	*authenticator
}

func New(username, password, product, version, url string, roundTripper http.RoundTripper) *Client {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	auth := &authenticator{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		username:   username,
		password:   password,
		authURL:    authURL,
		product:    product,
		version:    version,
		next:       roundTripper,
	}

	return &Client{
		URL:           url,
		HTTPClient:    &http.Client{Transport: auth},
		authenticator: auth,
	}
}

func call[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	target := c.URL + endpoint
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	req.Header.Add("Accept", "application/json")

	var response struct {
		MediaContainer T `json:"MediaContainer"`
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return response.MediaContainer, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return response.MediaContainer, errors.New(resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	return response.MediaContainer, err
}
