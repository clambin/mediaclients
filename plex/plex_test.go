package plex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clambin/mediaclients/plex"
	"github.com/clambin/mediaclients/plex/internal/testutil"
	"github.com/clambin/mediaclients/plex/plexauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Failures(t *testing.T) {
	c, s := makeClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server's having a hard day", http.StatusInternalServerError)
	}))

	ctx := t.Context()
	_, err := c.GetIdentity(ctx)
	require.Error(t, err)
	assert.Equal(t, "plex: 500 Internal Server Error", err.Error())

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
	assert.Equal(t, "decode: invalid character 'h' in literal true (expecting 'r')", err.Error())
}

func makeClientAndServer(h http.Handler) (*plex.Client, *httptest.Server) {
	if h == nil {
		h = &testutil.TestServer
	}
	server := httptest.NewServer(h)
	return plex.New(server.URL, plexauth.DefaultConfig.TokenSource().FixedToken("some-token"), plex.WithHTTPClient(&http.Client{})), server
}
