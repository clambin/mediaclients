package plex_test

import (
	"context"

	"github.com/clambin/mediaclients/plex"
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

func ExampleNewPMSClient_jwt() {
	// jwt requires persistence to store the private key for the device's client id.
	// vault provides a basic encrypted file to securely store the device's private data.
	v := vault.New[plextv.JWTSecureData](config.ClientID+".enc", "my-secret-passphrase")

	// create a token source that will use the provided credentials to authenticate with plextv the first time.
	// it then registers a public key with plextv and requests a JWT token using the private key.
	src := config.TokenSource(
		plextv.WithCredentials("plex-username", "plex-password"),
		plextv.WithJWT(v),
	)

	// create a plex.tv client with the provided token source.
	ctx := context.Background()
	plexTVClient := config.Client(ctx, src)

	// create a PMS client that will use the provided token source to authenticate itself with plex.tv
	// and determine the token to interact with the Plex Media Server.
	plexPMSClient := plex.NewPMSClient("http://plex-hostname:32400", plexTVClient)

	_, _ = plexPMSClient.GetLibraries(ctx)
}
