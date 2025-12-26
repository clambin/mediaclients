package plex

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/clambin/mediaclients/plex/plexauth"
	"golang.org/x/sync/singleflight"
)

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// Client calls the Plex API
type Client struct {
	httpClient *http.Client
	url        string
}

type PlexTVClient interface {
	MediaServers(ctx context.Context) ([]plexauth.RegisteredDevice, error)
}

// New creates a new Plex client, located at the given URL.
func New(url string, plexTVClient PlexTVClient, opts ...Option) *Client {
	client := Client{
		httpClient: &http.Client{},
		url:        url,
	}
	for _, opt := range opts {
		opt(&client)
	}
	client.httpClient.Transport = &authMiddleware{
		httpClient:   &http.Client{},
		next:         cmp.Or(client.httpClient.Transport, http.DefaultTransport),
		url:          url,
		plexTVClient: plexTVClient,
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

type authMiddleware struct {
	next         http.RoundTripper
	g            singleflight.Group
	httpClient   *http.Client
	plexTVClient PlexTVClient
	token        atomic.Pointer[plexauth.Token]
	url          string
}

func (a *authMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	token := a.token.Load()
	if token == nil {
		// get token
		tok, err := a.getToken(req.Context())
		if err != nil {
			return nil, fmt.Errorf("token: %w", err)
		}
		token = &tok
		a.token.Store(token)

	}
	req.Header.Set("X-Plex-Token", token.String())
	return a.next.RoundTrip(req)
}

func (a *authMiddleware) getToken(ctx context.Context) (plexauth.Token, error) {
	// get the Plex Media Server's ClientID (machineID)
	pmsClientID, err := a.getPMSClientID(ctx)
	if err != nil {
		return "", fmt.Errorf("get pms client id: %w", err)
	}
	// get all registered PMS devices
	devices, err := a.plexTVClient.MediaServers(ctx)
	if err != nil {
		return "", fmt.Errorf("media servers: %w", err)
	}

	// find the correct device and return its token
	for _, device := range devices {
		if device.ClientID == pmsClientID {
			return plexauth.Token(device.Token), nil
		}
	}
	// if no PMS server is found, return an error
	return "", fmt.Errorf("pms server not registered")
}

func (a *authMiddleware) getPMSClientID(ctx context.Context) (string, error) {
	// we can't call Identity() here as it causes an infinite loop, trying to get a token, so we roll our own.
	type identity struct {
		MachineIdentifier string `json:"machineIdentifier"`
		Version           string `json:"version"`
		Size              int    `json:"size"`
		Claimed           bool   `json:"claimed"`
	}
	var response mediaContainer[identity]
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.url+"/identity", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("identity: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return response.MediaContainer.MachineIdentifier, nil
}
