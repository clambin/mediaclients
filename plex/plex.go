package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/clambin/mediaclients/plex/plexauth"
)

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// Client calls the Plex API
type Client struct {
	httpClient  *http.Client
	tokenSource plexauth.TokenSource
	url         string
}

// New creates a new Plex client, located at the given URL. The tokenSource is used to obtain an authentication token.
// See [plexauth.Config.TokenSource] for more information.
func New(url string, tokenSource plexauth.TokenSource, opts ...Option) *Client {
	client := Client{
		httpClient:  http.DefaultClient,
		tokenSource: tokenSource,
		url:         url,
	}
	for _, opt := range opts {
		opt(&client)
	}

	return &client
}

func call[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	var response struct {
		MediaContainer T `json:"MediaContainer"`
	}

	token, err := c.tokenSource.Token(ctx)
	if err != nil {
		return response.MediaContainer, fmt.Errorf("plex auth: %w", err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.url+endpoint, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Token", token.String())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response.MediaContainer, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return response.MediaContainer, fmt.Errorf("http: %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	return response.MediaContainer, err
}
