package plextv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/clambin/go-common/testutils"
)

const (
	devicesResponse = `<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer publicAddress="81.164.72.79">
  <Device name="device-1" product="product-1" clientIdentifier="client-id-1" token="token-00000000000001"></Device>
  <Device name="device-2" product="product-2" clientIdentifier="client-id-2" token="token-00000000000002"></Device>
  <Device name="device-3" product="Plex Media Server" clientIdentifier="client-id-3" token="token-00000000000003"></Device>
</MediaContainer>`
)

var fakePlexTVServer = testutils.TestServer{Responses: map[string]testutils.PathResponse{
	"/api/v2/user": {http.MethodGet: testutils.Response{Body: User{Username: "user"}, StatusCode: http.StatusOK}},
	"/devices.xml": {http.MethodGet: testutils.Response{Body: devicesResponse, StatusCode: http.StatusOK}},
}}

func TestClient_User(t *testing.T) {
	ts := httptest.NewServer(&fakePlexTVServer)
	t.Cleanup(ts.Close)

	cfg := DefaultConfig().WithClientID("client-user")
	c := cfg.PlexTVClient(WithToken(legacyToken))
	c.config.URL = ts.URL

	user, err := c.User(t.Context())
	if err != nil {
		t.Fatalf("User error: %v", err)
	}
	if user.Username != "user" {
		t.Fatalf("unexpected user: %+v", user)
	}

	ts.Close()
	if _, err = c.User(t.Context()); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestClient_RegisteredDevices(t *testing.T) {
	ts := httptest.NewServer(&fakePlexTVServer)
	t.Cleanup(ts.Close)

	cfg := DefaultConfig().WithClientID("client-user")
	c := cfg.PlexTVClient(WithToken(legacyToken))
	c.config.URL = ts.URL

	devs, err := c.RegisteredDevices(t.Context())
	if err != nil {
		t.Fatalf("RegisteredDevices error: %v", err)
	}
	if got := len(devs); got != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devs))
	}

	ts.Close()
	if _, err = c.RegisteredDevices(t.Context()); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestClient_MediaServers(t *testing.T) {
	ts := httptest.NewServer(&fakePlexTVServer)
	t.Cleanup(ts.Close)

	cfg := DefaultConfig().WithClientID("client-user")
	c := cfg.PlexTVClient(WithToken(legacyToken))
	c.config.URL = ts.URL

	devs, err := c.MediaServers(t.Context())
	if err != nil {
		t.Fatalf("MediaServers error: %v", err)
	}
	if got := len(devs); got != 1 {
		t.Fatalf("expected 1 devices, got %d", len(devs))
	}
	if got := devs[0].ClientID; got != "client-id-3" {
		t.Fatalf("unexpected client ID: %s", got)
	}

	ts.Close()
	if _, err = c.User(t.Context()); err == nil {
		t.Fatalf("expected error from closed server")
	}
}
