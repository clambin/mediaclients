package plex

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/clambin/mediaclients/plex/plexhttp"
)

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

func WithRefreshingToken(httpClient *http.Client, initialToken string) Option {
	return func(client *Client) {
		// TODO: handle persistence
		client.tokenSource = &refreshingTokenSource{
			httpClient: cmp.Or(httpClient, http.DefaultClient),
			clientID:   client.clientID,
			token:      initialToken,
		}
	}
}

// Client calls the Plex APIs
type Client struct {
	httpClient  *http.Client
	tokenSource tokenSource
	url         string
	clientID    string
}

func New(url, token, clientID string, opts ...Option) *Client {
	client := Client{
		httpClient:  &http.Client{},
		url:         url,
		clientID:    clientID,
		tokenSource: &fixedTokenSource{token: token},
	}
	for _, o := range opts {
		o(&client)
	}
	return &client
}

func call[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	token, err := c.tokenSource.Token(ctx)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("token: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.url+endpoint, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Token", token)
	req.Header.Add("X-Plex-Client-Identifier", c.clientID)

	var response struct {
		MediaContainer T `json:"MediaContainer"`
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response.MediaContainer, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return response.MediaContainer, plexhttp.Parse(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	return response.MediaContainer, err
}
