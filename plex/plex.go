package plex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

func WithToken(token string) Option {
	return func(client *Client) {
		client.tokenSource = &fixedTokenSource{token: token}
	}
}

func WithCredentials(username, password string, identity ClientIdentity) Option {
	return func(client *Client) {
		client.tokenSource = &legacyCredentialsTokenSource{
			httpClient: &http.Client{},
			username:   username,
			password:   password,
			identity:   identity,
		}
	}
}

// Client calls the Plex APIs
type Client struct {
	httpClient  *http.Client
	tokenSource tokenSource
	url         string
}

func New(url string, opts ...Option) *Client {
	client := Client{
		httpClient: &http.Client{},
		url:        url,
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
		return response.MediaContainer, errors.New(resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		err = fmt.Errorf("decode: %w", err)
	}

	return response.MediaContainer, err
}
