package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// Client interacts with a Plex Media Server.
type Client struct {
	httpClient *http.Client
	url        string
	token      string
}

// New returns a new Client that interacts with a Plex Media Server at url, using token for authentication.
func New(url string, token string, opts ...Option) *Client {
	client := Client{
		httpClient: http.DefaultClient,
		url:        url,
		token:      token,
	}
	for _, opt := range opts {
		opt(&client)
	}
	return &client
}

type mediaContainer[T any] struct {
	MediaContainer T `json:"MediaContainer"`
}

func call[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	var response mediaContainer[T]
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.url+endpoint, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", c.token)

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
