package plex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/clambin/mediaclients/plex/plextv"
	"golang.org/x/sync/singleflight"
)

type Option func(*PMSClient)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *PMSClient) {
		client.httpClient = httpClient
	}
}

type tokenSource interface {
	Token(ctx context.Context) (plextv.Token, error)
}

var (
	_ tokenSource = fixedTokenSource{}
	_ tokenSource = (*plexTVTokenSource)(nil)
)

// fixedTokenSource returns a fixed token.
type fixedTokenSource struct {
	token plextv.Token
}

func (f fixedTokenSource) Token(_ context.Context) (plextv.Token, error) {
	return f.token, nil
}

// plexTVTokenSource retrieves a Plex Media Server's token via plex.tv.
//
// It retrieves the PMS' ClientID using the /identity endpoint (which doesn't need a plex token).
// Then it lists all registered PMS devices on plex.tv and returns the token of the first one that matches the ClientID.
type plexTVTokenSource struct {
	httpClient   *http.Client
	plexTVClient PlexTVClient
	token        atomic.Pointer[plextv.Token]
	g            singleflight.Group
	url          string
}

func (t *plexTVTokenSource) Token(ctx context.Context) (plextv.Token, error) {
	// pms tokens don't expire, so we can cache them
	if token := t.token.Load(); token != nil {
		return *token, nil
	}

	// use singleflight to avoid multiple authentications
	tok, err, _ := t.g.Do("pms-token", func() (any, error) {
		// get the PMS token
		pmsToken, err := t.pmsToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("get pms token: %w", err)
		}

		// call user endpoint to update the device parameters
		if _, err = t.plexTVClient.User(ctx); err != nil {
			return nil, fmt.Errorf("user: %w", err)
		}

		return pmsToken, nil
	})
	if err != nil {
		return "", err
	}

	token := tok.(plextv.Token)
	t.token.Store(&token)

	return token, nil
}

func (t *plexTVTokenSource) pmsToken(ctx context.Context) (plextv.Token, error) {
	// get the PMS ClientID
	pmsClientID, err := t.pmsClientID(ctx)
	if err != nil {
		return "", fmt.Errorf("get pms client id: %w", err)
	}
	// get all registered PMS devices
	devices, err := t.plexTVClient.MediaServers(ctx)
	if err != nil {
		return "", fmt.Errorf("media servers: %w", err)
	}

	// find the correct device and return its token
	for _, device := range devices {
		if device.ClientID == pmsClientID {
			return plextv.Token(device.Token), nil
		}
	}
	// if no PMS server is found, return an error
	return "", errors.New("pms server not registered")

}

func (t *plexTVTokenSource) pmsClientID(ctx context.Context) (string, error) {
	// we're in the process of getting a PMS token. So we can't call Identity() here as it would cause an infinite loop.
	// instead, so we roll our own.
	type identity struct {
		MachineIdentifier string `json:"machineIdentifier"`
		Version           string `json:"version"`
		Size              int    `json:"size"`
		Claimed           bool   `json:"claimed"`
	}
	var response mediaContainer[identity]
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, t.url+"/identity", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("identity: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return response.MediaContainer.MachineIdentifier, nil
}

// PMSClient interacts with a Plex Media Server.
type PMSClient struct {
	httpClient  *http.Client
	tokenSource tokenSource
	url         string
}

type PlexTVClient interface {
	User(ctx context.Context) (plextv.User, error)
	MediaServers(ctx context.Context) ([]plextv.RegisteredDevice, error)
}

func NewPMSClientWithToken(url string, token string, opts ...Option) *PMSClient {
	client := PMSClient{
		url:         url,
		httpClient:  http.DefaultClient,
		tokenSource: &fixedTokenSource{token: plextv.Token(token)},
	}
	for _, opt := range opts {
		opt(&client)
	}
	return &client
}

// NewPMSClient creates a new Plex client, located at the given URL.
func NewPMSClient(url string, plexTVClient PlexTVClient, opts ...Option) *PMSClient {
	client := PMSClient{
		url:        url,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(&client)
	}
	client.tokenSource = &plexTVTokenSource{
		plexTVClient: plexTVClient,
		url:          url,
		httpClient:   client.httpClient,
	}
	return &client
}

type mediaContainer[T any] struct {
	MediaContainer T `json:"MediaContainer"`
}

func call[T any](ctx context.Context, c *PMSClient, endpoint string) (T, error) {
	token, err := c.tokenSource.Token(ctx)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("token: %w", err)
	}

	var response mediaContainer[T]
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.url+endpoint, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", token.String())

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
