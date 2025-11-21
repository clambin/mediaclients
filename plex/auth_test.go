package plex

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/clambin/mediaclients/plex/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_WithCredentials(t *testing.T) {
	const plexToken = "1234"
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := r.Header.Get("X-Plex-Token"); token != plexToken {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	s := httptest.NewServer(mw(&testutil.TestServer))
	t.Cleanup(s.Close)
	auth := authServer{token: plexToken}
	as := httptest.NewServer(&auth)
	t.Cleanup(as.Close)

	// create a client with credentials and redirect to the auth server
	c := New(s.URL, WithCredentials("username", "password", ClientIdentity{Identifier: "foo"}))
	c.tokenSource.(*legacyCredentialsTokenSource).authURL = as.URL

	// the first call uses the auth server
	_, err := c.GetIdentity(t.Context())
	require.NoError(t, err)

	// the second call uses the cached token
	_, err = c.GetIdentity(t.Context())
	require.NoError(t, err)
	assert.Equal(t, int32(1), auth.calls.Load())
}

type authServer struct {
	token string
	calls atomic.Int32
}

func (a *authServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.calls.Add(1)
	body, _ := io.ReadAll(r.Body)
	if string(body) != "user%5Blogin%5D=username&user%5Bpassword%5D=password" {
		http.Error(w, "invalid username/password", http.StatusUnauthorized)
		return
	}
	if r.Header.Get("X-Plex-Client-Identifier") != "foo" {
		http.Error(w, "missing client identifier", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`<user authenticationToken="` + a.token + `"></user>`))
}
