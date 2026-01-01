package plex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clambin/mediaclients/plex"
	"github.com/clambin/mediaclients/plex/internal/testutil"
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

func makeClientAndServer(h http.Handler) (*plex.PMSClient, *httptest.Server) {
	if h == nil {
		h = &testutil.TestServer
	}
	server := httptest.NewServer(h)
	client := plex.NewPMSClientWithToken(server.URL, "my-pms-token")
	return client, server
}
