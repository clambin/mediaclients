package plextv

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Device identifies the client when using Plex username/password credentials.
// Although this package provides a default, it is recommended to set this yourself.
//
// Limitation: currently the Device attributes are only registered when the device is registered
// (during [Config.RegisterWithCredentials] or [Config.RegisterWithPIN]).  If the attributes change
// after registration (e.g., the Version is updated), this is not (yet) reflected in the registered
// device on plex.tv.
type Device struct {
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
	// In Authorized Devices, it is shown on line 80.
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
	// Provides describes the type of device.
	// Passed as X-Plex-Provides header.
	Provides string
}

// populateRequest populates the request headers with the device information.
func (id Device) populateRequest(req *http.Request) {
	headers := map[string]string{
		"X-Plex-Product":          id.Product,
		"X-Plex-Version":          id.Version,
		"X-Plex-Platform":         id.Platform,
		"X-Plex-Platform-Version": id.PlatformVersion,
		"X-Plex-Device":           id.Device,
		"X-Plex-Device-Vendor":    id.DeviceVendor,
		"X-Plex-Device-Name":      id.DeviceName,
		"X-Plex-Model":            id.Model,
		"X-Plex-Provides":         id.Provides,
	}
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
}

// Config contains the configuration required to authenticate with Plex.
type Config struct {
	// Device information used during username/password authentication.
	Device Device
	// URL is the base URL of the legacy Plex authentication endpoint.
	// Defaults to https://plex.tv and should not need to be changed.
	URL string
	// V2URL is the base URL of the new Plex authentication endpoint.
	// Defaults to https://clients.plex.tv and should not need to be changed.
	V2URL string
	// ClientID is the unique identifier of the client application.
	ClientID string
	aud      string
	// Scopes is a list of scopes to request.
	// This may become non-exported in the future.
	Scopes []string
	// TokenTTL is the duration of the authentication token.
	// Defaults to 7 days, in line with Plex specifications.
	// Normally, this should not need to be changed.
	tokenTTL time.Duration
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		URL:      "https://plex.tv",
		V2URL:    "https://clients.plex.tv", // TODO: do any endpoints mandate clients.plex.tv?
		Scopes:   []string{"username", "email", "friendly_name", "restricted", "anonymous"},
		ClientID: uuid.New().String(),
		aud:      "plex.tv",
		tokenTTL: 7 * 24 * time.Hour,
	}
}

// WithClientID sets the Client ID.
func (c Config) WithClientID(clientID string) Config {
	c.ClientID = clientID
	return c
}

// WithDevice sets the device information used during username/password and pin authentication.
//
// See the [Device] type for details on what each field means.
//
// Limitation: currently the Device attributes are only registered when the device is registered
// // (during [Config.RegisterWithCredentials] or [Config.RegisterWithPIN]).  If the attributes change
// // after registration (e.g., the Version is updated), this is not (yet) reflected in the registered
// // device on plex.tv.
func (c Config) WithDevice(device Device) Config {
	c.Device = device
	return c
}
