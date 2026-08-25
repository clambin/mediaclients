package plextv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/clambin/go-common/testutils"
)

var (
	devicesResponse = []Device{
		{Name: "my device #1", Token: "token-00000000000001", Product: "Plex Media Server", ClientIdentifier: "client-00000000000001"},
		{Name: "my device #2", Token: "token-00000000000002"},
		{Name: "my device #3", Token: "token-00000000000003"},
	}
	fakePlexTVServer = testutils.TestServer{Responses: map[string]testutils.PathResponse{
		"/api/v2/user":    {http.MethodGet: testutils.Response{Body: User{Username: "user"}, StatusCode: http.StatusOK}},
		"/api/v2/devices": {http.MethodGet: testutils.Response{Body: devicesResponse, StatusCode: http.StatusOK}},
	}}
)

func TestClient_User(t *testing.T) {
	ts := httptest.NewServer(&fakePlexTVServer)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	cfg := DefaultConfig().WithClientID("client-user")
	cfg.URL = ts.URL
	cfg.V2URL = ts.URL
	c := cfg.Client(ctx, cfg.TokenSource(WithToken(legacyToken)))

	user, err := c.User(ctx)
	if err != nil {
		t.Fatalf("User error: %v", err)
	}
	if user.Username != "user" {
		t.Fatalf("unexpected user: %+v", user)
	}

	ts.Close()
	if _, err = c.User(ctx); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestClient_Devices(t *testing.T) {
	ts := httptest.NewServer(&fakePlexTVServer)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	cfg := DefaultConfig().WithClientID("client-user")
	cfg.URL = ts.URL
	cfg.V2URL = ts.URL
	c := cfg.Client(ctx, cfg.TokenSource(WithToken(legacyToken)))
	//c.config.URL = ts.URL

	devs, err := c.Devices(ctx, nil)
	if err != nil {
		t.Fatalf("Devices error: %v", err)
	}
	if got := len(devs); got != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devs))
	}

	ts.Close()
	if _, err = c.Devices(ctx, nil); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestClient_MediaServers(t *testing.T) {
	ts := httptest.NewServer(&fakePlexTVServer)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	cfg := DefaultConfig().WithClientID("client-user")
	cfg.URL = ts.URL
	cfg.V2URL = ts.URL
	c := cfg.Client(ctx, cfg.TokenSource(WithToken(legacyToken)))

	devs, err := c.MediaServers(ctx)
	if err != nil {
		t.Fatalf("MediaServers error: %v", err)
	}
	if got := len(devs); got != 1 {
		t.Fatalf("expected 1 devices, got %d", len(devs))
	}
	want := "client-00000000000001"
	if got := devs[0].ClientIdentifier; got != want {
		t.Fatalf("unexpected client ID. want: %s, got: %s", want, got)
	}

	ts.Close()
	if _, err = c.User(ctx); err == nil {
		t.Fatalf("expected error from closed server")
	}
}
