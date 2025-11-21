package plex

import (
	"bytes"
	"cmp"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/google/uuid"
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		defaultClientIdentity.Version = info.Main.Version
	}
	defaultClientIdentity.DeviceName, _ = os.Hostname()
}

const legacyAuthURL = "https://plex.tv/users/sign_in.xml"

// ClientIdentity identifies the client when using Plex username/password credentials.
// Although this package provides a default, it is recommended to set this yourself.
type ClientIdentity struct {
	// Product is the name of the client product.
	// Passed as X-Plex-Product header.
	// In Authorized Devices, it is shown on line 3.
	Product string
	// Version is the version of the client application.
	// Passed as X-Plex-Version header.
	// In Authorized Devices, it is shown on line 2.
	Version string
	// Platform is the operating system or compiler of the client application.
	// Passed as X-Plex-Platform header.
	Platform string
	// PlatformVersion is the version of the platform.
	// Passed as X-Plex-Platform-Version header.
	PlatformVersion string
	// Device is a relatively friendly name for the client device.
	// Passed as X-Plex-Device header.
	// In Authorized Devices, it is shown on line 4.
	Device string
	// Model is a potentially less friendly identifier for the device model.
	// Passed as X-Plex-Model header.
	Model string
	// DeviceVendor is the name of the device vendor.
	// Passed as X-Plex-Device-Vendor header.
	DeviceVendor string
	// DeviceName is a friendly name for the client.
	// Passed as X-Plex-Device-Name header.
	// In Authorized Devices, it is shown on line 1.
	DeviceName string
	// Identifier is a unique identifier for the client.
	// Passed as X-Plex-Client-Identifier header.
	Identifier string
}

func (id ClientIdentity) populateRequest(req *http.Request) {
	headers := map[string]string{
		"X-Plex-Product":           id.Product,
		"X-Plex-Version":           id.Version,
		"X-Plex-Platform":          id.Platform,
		"X-Plex-Platform-Version":  id.PlatformVersion,
		"X-Plex-Device":            id.Device,
		"X-Plex-Device-Vendor":     id.DeviceVendor,
		"X-Plex-Device-Name":       id.DeviceName,
		"X-Plex-Model":             id.Model,
		"X-Plex-Client-Identifier": id.Identifier,
	}
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
}

var defaultClientIdentity = ClientIdentity{
	Product:         "github.com/clambin/mediaclients/plex",
	Version:         "(devel)",
	Device:          "plex",
	Platform:        runtime.GOOS,
	PlatformVersion: runtime.Version(),
	Identifier:      uuid.New().String(),
}

type tokenSource interface {
	Token(context.Context) (string, error)
}

// fixedTokenSource returns a tokenSource that always returns the same token.
type fixedTokenSource struct {
	token string
}

func (s *fixedTokenSource) Token(_ context.Context) (string, error) {
	return s.token, nil
}

// legacyCredentialsTokenSource returns a tokenSource that authenticates with the legacy Plex API auth endpoint,
// i.e., https://plex.tv/users/sign_in.xml
type legacyCredentialsTokenSource struct {
	httpClient *http.Client
	identity   ClientIdentity
	username   string
	password   string
	token      string
	authURL    string
	lock       sync.Mutex
}

func (s *legacyCredentialsTokenSource) Token(ctx context.Context) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// return cached token if available
	if s.token != "" {
		return s.token, nil
	}

	// credentials are passed in the request body, as a url-encoded form
	v := make(url.Values)
	v.Set("user[login]", s.username)
	v.Set("user[password]", s.password)

	// call the auth endpoint
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cmp.Or(s.authURL, legacyAuthURL), bytes.NewBufferString(v.Encode()))
	s.identity.populateRequest(req)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// a successful response contains an XML document with an authentication token
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("http: %s", resp.Status)
	}
	var authResponse struct {
		XMLName             xml.Name `xml:"user"`
		AuthenticationToken string   `xml:"authenticationToken,attr"`
	}
	if err = xml.NewDecoder(resp.Body).Decode(&authResponse); err == nil {
		s.token = authResponse.AuthenticationToken
	}

	return s.token, err
}
