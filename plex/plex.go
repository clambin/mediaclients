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

func call[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	var response T

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.url+endpoint, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("http: %d - %s", resp.StatusCode, resp.Status)
	}

	//var buf bytes.Buffer
	//r := io.TeeReader(resp.Body, &buf)
	r := resp.Body

	if err = json.NewDecoder(r).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	//r2 := buf.String()[:min(500, buf.Len())]
	//_ = r2
	return response, err
}
