package plextv

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// Client returns a [Client] that uses the provided [TokenSource] to authenticate itself with plex.tv.
//
// The returned client will use the http.Client associated with the provided context to make requests,
// or a default one if none is set.
func (c Config) Client(ctx context.Context, src TokenSource) Client {
	// create a new httpClient to interact with plex.tv.
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
// This call also updates the DeviceInformation information in plex.tv.
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

// Resources returns all resources (mainly Plex Media Servers) visible for the current token.
//
// Use values to filter the results. According to the [Plex API documentation], the following values are supported:
// - includeHttps=1: include only HTTPS resources
// - includeRelay=1: include only relay resources
// - includeIPv6=1: include only IPv6 resources
//
// [Plex API documentation]: https://developer.plex.tv/pms/#section/API-Info/Authenticating-with-Plex
func (c Client) Resources(ctx context.Context, values url.Values) ([]Resource, error) {
	target := c.config.V2URL + "/api/v2/resources"
	if len(values) > 0 {
		target += "?" + values.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
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
//
// Use values to filter the results. According to the [Plex API documentation], the following values are supported:
// - includeHttps=1: include only HTTPS resources
// - includeRelay=1: include only relay resources
// - includeIPv6=1: include only IPv6 resources
//
// This call can be used to list the Plex Media Server (PMS) instances available to the token.
// Use the Token to interact with the PMS instance and the list of connection URLs to locate it.
// Connections labeled as local should be preferred over those that are not,
// and relay should only be used as a last resort as bandwidth on relay connections is limited.
//
// [Plex API documentation]: https://developer.plex.tv/pms/#section/API-Info/Authenticating-with-Plex
func (c Client) Devices(ctx context.Context, values url.Values) ([]Device, error) {
	target := c.config.V2URL + "/api/v2/devices"
	if len(values) > 0 {
		target += "?" + values.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var buf bytes.Buffer
	r := io.TeeReader(resp.Body, &buf)

	var devices []Device
	if err = json.NewDecoder(r).Decode(&devices); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return devices, nil
}

// MediaServers is a convenience function that returns all Plex Media Servers registered under the provided token
func (c Client) MediaServers(ctx context.Context) ([]Device, error) {
	// get all devices
	devices, err := c.Devices(ctx, nil)
	if err == nil {
		// remove any non-Plex Media Server devices
		devices = slices.DeleteFunc(devices, func(device Device) bool {
			return device.Product != "Plex Media Server"
		})
	}
	return devices, err
}

// RegisteredDevices is similar to Devices, but calls the legacy /devices.xml endpoint.
//
// Currently, the endpoint returns all devices, but the Token is not provided.
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

	// for troubleshooting:
	// var cp bytes.Buffer
	// r := io.TeeReader(resp.Body, &cp)
	r := resp.Body

	if err = xml.NewDecoder(r).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return response.Devices, nil
}

var _ http.RoundTripper = (*authMiddleware)(nil)

// authMiddleware adds the X-Plex-Token, X-Plex-Client-Identifier and Plex device headers to outgoing requests.
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
	a.config.Device.addDeviceHeaders(r)
	return a.next.RoundTrip(r)
}
