package plex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clambin/mediaclients/plex"
	"github.com/clambin/mediaclients/plex/internal/testutil"
	"github.com/clambin/mediaclients/plex/plextv"
	"github.com/stretchr/testify/require"
)

func TestClient_Failures(t *testing.T) {
	c, s := makeClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server's having a hard day", http.StatusInternalServerError)
	}))

	ctx := t.Context()
	_, err := c.GetIdentity(ctx)
	require.Error(t, err)

	s.Close()
	_, err = c.GetIdentity(ctx)
	require.Error(t, err)
	//assert.ErrorIs(t, err, unix.ECONNREFUSED)
}

func TestClient_Decode_Failure(t *testing.T) {
	c, s := makeClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("this is definitely not json"))
	}))
	t.Cleanup(s.Close)

	_, err := c.GetIdentity(context.Background())
	require.Error(t, err)
}

func makeClientAndServer(h http.Handler) (*plex.Client, *httptest.Server) {
	if h == nil {
		h = &testutil.TestServer
	}
	server := httptest.NewServer(h)

	ptvc := fakePlexTVClient{devices: []plextv.RegisteredDevice{
		{ClientID: "pms-client-id-srv1", Token: "legacy-token-srv1"},
		{ClientID: "pms-client-id-srv2", Token: "legacy-token-srv2"},
	}}

	client := plex.New(server.URL, &ptvc)
	return client, server
}

var _ plex.PlexTVClient = &fakePlexTVClient{}

type fakePlexTVClient struct {
	devices []plextv.RegisteredDevice
	err     error
}

func (f fakePlexTVClient) User(_ context.Context) (plextv.User, error) {
	return plextv.User{}, nil
}

func (f fakePlexTVClient) MediaServers(_ context.Context) ([]plextv.RegisteredDevice, error) {
	return f.devices, f.err
}
