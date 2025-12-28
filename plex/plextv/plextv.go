package plextv

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"slices"
	"time"
)

// Client interacts with the plex.tv API.
//
// Currently, only supports /api/v2/user and /devices.xml endpoints.
type Client struct {
	httpClient *http.Client
	config     *Config
}

// Client returns a [Client].
func (c Config) Client(ctx context.Context, src TokenSource) Client {
	// create a new httpClient to interact with plex.tv, using the same transport.
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: httpClient(ctx).Transport,
	}
	// add middleware to request a token and fill in the Plex headers.
	client.Transport = &authMiddleware{
		config:      &c,
		tokenSource: src,
		next:        client.Transport,
	}
	return Client{
		config:     &c,
		httpClient: client,
	}
}

// User returns the information of the user associated with the Client's TokenSource.
// This call also updates the Device information in plex.tv.
func (c Client) User(ctx context.Context) (User, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+"/api/v2/user", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	var user User
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return user, fmt.Errorf("decode: %w", err)
	}
	return user, nil
}

// RegisteredDevices returns all devices registered under the provided token
func (c Client) RegisteredDevices(ctx context.Context) ([]RegisteredDevice, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+"/devices.xml", nil)
	req.Header.Set("Accept", "application/xml")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("devices: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var response struct {
		XMLName       xml.Name           `xml:"MediaContainer"`
		PublicAddress string             `xml:"publicAddress,attr"`
		Devices       []RegisteredDevice `xml:"Device"`
	}
	if err = xml.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return response.Devices, nil
}

// MediaServers returns all Plex Media Servers registered under the provided token
func (c Client) MediaServers(ctx context.Context) ([]RegisteredDevice, error) {
	// get all devices
	devices, err := c.RegisteredDevices(ctx)
	if err == nil {
		// remove any non-Plex Media Server devices
		devices = slices.DeleteFunc(devices, func(device RegisteredDevice) bool {
			return device.Product != "Plex Media Server"
		})
	}
	return devices, err
}

var _ http.RoundTripper = (*authMiddleware)(nil)

// authMiddleware adds the X-Plex-Token and X-Plex-Client-Identifier and Plex device headers to outgoing requests.
type authMiddleware struct {
	config      *Config
	tokenSource TokenSource
	next        http.RoundTripper
}

func (a *authMiddleware) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	token, err := a.tokenSource.Token(r.Context())
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	r.Header.Set("X-Plex-Token", token.String())
	r.Header.Set("X-Plex-Client-Identifier", a.config.ClientID)
	a.config.Device.populateRequest(r)
	return a.next.RoundTrip(r)
}

/*
// Resources returns all resources (mainly Plex Media Servers) visible for the current token.
//
// Use values to filter the results. According to the [Plex API documentation], the following values are supported:
// - includeHttps=1: include only HTTPS resources
// - includeRelay=1: include only relay resources
// - includeIPv6=1: include only IPv6 resources
//
// This call can be used to list the Plex Media Server (PMS) instances available to the token.
// Use the AccessToken to interact with the PMS instance and the list of connection URLs to locate it.
// Connections labeled as local should be preferred over those that are not,
// and relay should only be used as a last resort as bandwidth on relay connections is limited.
//
// [Plex API documentation]: https://developer.plex.tv/pms/#section/API-Info/Authenticating-with-Plex
func (c Client) Resources(ctx context.Context, values url.Values) ([]Resource, error) {
	target := c.config.V2URL + "/api/v2/resources"
	if len(values) > 0 {
		target += "?" + values.Encode()
	}
	resp, err := c.doWithToken(ctx, http.MethodGet, target, nil, http.StatusOK, func(req *http.Request) {
		c.config.Device.populateRequest(req)
	})
	if err != nil {
		return nil, err
	}
	var resources []Resource
	if err = json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return resources, nil
}

// Devices return all devices visible for the current token. It's the response to /api/v2/devices endpoint.
func (c Client) Devices(ctx context.Context, values url.Values) ([]PlexTVDevice, error) {
	target := c.config.V2URL + "/api/v2/devices"
	if len(values) > 0 {
		target += "?" + values.Encode()
	}
	resp, err := c.doWithToken(ctx, http.MethodGet, target, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var devices []PlexTVDevice
	if err = json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return devices, nil
}
*/
