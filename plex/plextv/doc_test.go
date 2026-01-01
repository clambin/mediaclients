package plextv_test

import (
	"context"
	"fmt"
	"time"

	"github.com/clambin/mediaclients/plex/plextv"
	"github.com/clambin/mediaclients/plex/vault"
)

// config contains the attributes that will be registered for the device with the provided client id.
var config = plextv.DefaultConfig().
	WithClientID("my-unique-client-id").
	WithDevice(plextv.Device{
		Product:         "my product",
		Version:         "v0.0.4",
		Platform:        "my platform",
		PlatformVersion: "my platform version",
		Device:          "my device",
		Model:           "my device model",
		DeviceVendor:    "my device vendor name",
		DeviceName:      "my device name",
		Provides:        "controller",
	})

func ExampleConfig_Client_credentials() {
	// create a token source that uses the PIN flow to authenticate the device with plex.tv.
	src := config.TokenSource(plextv.WithCredentials("plex-username", "plex-password"))

	// create a plex.tv client with the provided token source.
	ctx := context.Background()
	_ = config.Client(ctx, src)
}

func ExampleConfig_Client_pin() {
	// create a token source that uses the PIN flow to authenticate the device with plex.tv.
	src := config.TokenSource(
		plextv.WithPIN(
			// the callback to ask the user to log in
			func(_ plextv.PINResponse, url string) {
				fmt.Println("Confirm login request:", url)
			},
			// the interval to poll for the PIN Response
			10*time.Second,
		),
	)

	// create a plex.tv client with the provided token source.
	ctx := context.Background()
	_ = config.Client(ctx, src)
}

func ExampleConfig_Client_jwt() {
	// jwt requires persistence to store the private key for the device's client id.
	// vault provides a basic encrypted file to securely store the device's private data.
	v := vault.New[plextv.JWTSecureData](config.ClientID+".enc", "my-secret-passphrase")

	// create a token source that will use the provided credentials to authenticate with plextv the first time.
	// it then registers a public key with plextv and requests a JWT token using the private key.
	//
	// Note: the JWT flow requires an initial valid token to publish its public key. This can be either through
	// credentials or PIN flow. Once JWT authentication is enabled, you can't use credentials or PIN anymore for
	// the device's ClientIdentifier.
	src := config.TokenSource(
		plextv.WithCredentials("plex-username", "plex-password"),
		plextv.WithJWT(v),
	)

	// create a plex.tv client with the provided token source.
	ctx := context.Background()
	_ = config.Client(ctx, src)
}
