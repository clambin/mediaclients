package plexauth

import (
	"testing"
)

func TestPlexTVClient_RegisteredDevices_And_MediaServers(t *testing.T) {
	// TODO: probably don't need a real test server anymore for PlexTV tests. Mocking the handler is probably enough.
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()
	c := cfg.PlexTVClient(WithToken(legacyToken))

	devs, err := c.RegisteredDevices(ctx)
	if err != nil {
		t.Fatalf("RegisteredDevices error: %v", err)
	}
	if len(devs) != 3 {
		t.Fatalf("expected 2 devices, got %d", len(devs))
	}
	servers, err := c.MediaServers(ctx)
	if err != nil {
		t.Fatalf("MediaServers error: %v", err)
	}
	if len(servers) != 2 || servers[0].Name != "srv1" || servers[1].Name != "srv2" {
		t.Fatalf("unexpected servers: %+v", servers)
	}
	// errors
	ts.Close()
	if _, err = c.RegisteredDevices(ctx); err == nil {
		t.Fatalf("expected error from closed server")
	}
	if _, err = c.MediaServers(ctx); err == nil {
		t.Fatalf("expected error from closed server")
	}
}
