package plex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/clambin/mediaclients/plex/plexauth"
)

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// Client calls the Plex APIs
type Client struct {
	httpClient  *http.Client
	tokenSource plexauth.AuthTokenSource
	url         string
}

func New(url string, tokenSource plexauth.AuthTokenSource, opts ...Option) *Client {
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
		// TODO: do errors match the Plex Cloud API?
		return response.MediaContainer, plexauth.ParsePlexError(resp)
	}

	var buf bytes.Buffer
	if err = json.NewDecoder(io.TeeReader(resp.Body, &buf)).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}
	_ = buf

	return response.MediaContainer, err
}
