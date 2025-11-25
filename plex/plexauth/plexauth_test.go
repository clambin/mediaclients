package plexauth_test

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"

	"github.com/clambin/mediaclients/plex/plexauth"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthConfig_AuthorizeDevice(t *testing.T) {
	ts := httptest.NewServer(authV1Server{
		username:        "username",
		password:        "password",
		token:           "token",
		expectedHeaders: []string{"X-Plex-Client-Identifier", "X-Plex-Platform", "X-Plex-Platform-Version"},
	})
	cfg := plexauth.DefaultAuthConfig.WithClientID("123").WithDevice(plexauth.Device{
		Platform:        runtime.GOOS,
		PlatformVersion: runtime.Version(),
	})
	cfg.AuthURL = ts.URL

	ctx := plexauth.WithHTTPClient(t.Context(), &http.Client{})
	token, err := cfg.AuthorizeDevice(ctx, "username", "password")
	require.NoError(t, err)
	assert.Equal(t, "token", token)
	_, err = cfg.AuthorizeDevice(ctx, "username", "wrongpassword")
	require.Error(t, err)
	ts.Close()
	_, err = cfg.AuthorizeDevice(ctx, "username", "password")
	require.Error(t, err)
}

func TestAuthConfig_GetAuthToken(t *testing.T) {
	ts1 := httptest.NewServer(authV1Server{
		username: "username",
		password: "password",
		token:    "token",
	})
	t.Cleanup(ts1.Close)
	var authServer authV2Server
	ts2 := httptest.NewServer(&authServer)
	t.Cleanup(ts2.Close)
	cfg := plexauth.DefaultAuthConfig.WithClientID("123")
	cfg.AuthURL = ts1.URL
	cfg.AuthV2URL = ts2.URL
	ctx := t.Context()

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, "token") // test server doesn't validate token
	require.NoError(t, err)

	newToken, err := cfg.GetAuthToken(ctx, privateKey, keyID)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, "token", newToken)
}

type authV1Server struct {
	username        string
	password        string
	token           string
	expectedHeaders []string
}

func (a authV1Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if string(body) != `user%5Blogin%5D=`+a.username+`&user%5Bpassword%5D=`+a.password {
		http.Error(w, "invalid username/password", http.StatusUnauthorized)
		return
	}
	for _, header := range a.expectedHeaders {
		if r.Header.Get(header) == "" {
			http.Error(w, "missing header: "+header, http.StatusBadRequest)
		}
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`<user authenticationToken="` + a.token + `"></user>`))
}

var _ http.Handler = (*authV2Server)(nil)

type authV2Server struct {
	keys map[string]jwk.Key
	lock sync.Mutex
}

func (a *authV2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v2/auth/nonce":
		a.handleNonce(w, r)
	case "/api/v2/auth/jwk":
		a.handleJWK(w, r)
	case "/api/v2/auth/token":
		a.handleToken(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (a *authV2Server) handleNonce(w http.ResponseWriter, _ *http.Request) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{ "nonce": "` + hex.EncodeToString(nonce) + `" }`))
}

func (a *authV2Server) handleJWK(w http.ResponseWriter, r *http.Request) {
	var parsed struct {
		JWK json.RawMessage `json:"jwk"`
	}
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &parsed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := jwk.ParseKey(parsed.JWK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pub, err := key.PublicKey()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.lock.Lock()
	defer a.lock.Unlock()
	if a.keys == nil {
		a.keys = make(map[string]jwk.Key)
	}
	a.keys[r.Header.Get("X-Plex-Client-Identifier")] = pub
	w.WriteHeader(http.StatusCreated)
}

func (a *authV2Server) handleToken(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	signedToken, ok := req["jwt"]
	if !ok {
		http.Error(w, "missing jwt", http.StatusBadRequest)
		return
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	pub, ok := a.keys[r.Header.Get("X-Plex-Client-Identifier")]
	if !ok {
		http.Error(w, "no key found for clientID", http.StatusUnauthorized)
		return
	}
	// Parse and verify JWT
	parsed, err := jwt.Parse([]byte(signedToken), jwt.WithKey(jwa.EdDSA(), pub), jwt.WithVerify(true))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// we ignore the payload for now.  mainly token expiration is of interest.
	_ = parsed

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{ "auth_token": "` + uuid.New().String() + `" }`))
}
