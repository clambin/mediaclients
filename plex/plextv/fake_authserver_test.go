package plextv

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const legacyToken = "12345678901234567890"

var baseConfig = DefaultConfig().
	WithClientID("abc").
	WithDevice(Device{
		Product:         "TestProduct",
		Version:         "1.0",
		Platform:        "unit",
		PlatformVersion: "test",
		Device:          "dev",
		DeviceVendor:    "vendor",
		DeviceName:      "devname",
		Model:           "model",
	})

func newTestServer(cfg Config) (Config, *fakeAuthServer, *httptest.Server) {
	s := makeFakeServer(&cfg)
	ts := httptest.NewServer(&s)
	cfg.URL = ts.URL
	cfg.V2URL = ts.URL
	return cfg, &s, ts
}

var _ http.Handler = &fakeAuthServer{}

type fakeAuthServer struct {
	http.Handler
	tokens     *tokens
	config     *Config
	jwtHandler *jwtHandler
}

func makeFakeServer(cfg *Config) fakeAuthServer {
	t := tokens{tokens: make(map[string]string)}
	f := fakeAuthServer{
		config:     cfg,
		tokens:     &t,
		jwtHandler: &jwtHandler{keySets: make(map[string]jwk.Set), tokens: &t},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/sign_in.xml", f.handleRegisterWithCredentials)
	mux.HandleFunc("POST /api/v2/pins", f.handlePIN)
	mux.HandleFunc("GET /api/v2/pins/", f.handleValidatePIN)
	mux.HandleFunc("POST /api/v2/auth/jwk", f.jwtHandler.handleJWK)
	mux.HandleFunc("GET /api/v2/auth/nonce", f.jwtHandler.handleNonce)
	mux.HandleFunc("POST /api/v2/auth/token", f.jwtHandler.handleJWToken)
	mux.HandleFunc("GET /devices.xml", f.handleDevices)
	f.Handler = mux
	return f
}

func (f fakeAuthServer) handleRegisterWithCredentials(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Content-Type":             "application/x-www-form-urlencoded",
		"Accept":                   "application/xml",
		"X-Plex-Client-Identifier": f.config.ClientID,
		"X-Plex-Product":           f.config.Device.Product,
		"X-Plex-Version":           f.config.Device.Version,
		"X-Plex-Platform":          f.config.Device.Platform,
		"X-Plex-Platform-Version":  f.config.Device.PlatformVersion,
		"X-Plex-Device":            f.config.Device.Device,
		"X-Plex-Device-Vendor":     f.config.Device.DeviceVendor,
		"X-Plex-Device-Name":       f.config.Device.DeviceName,
		"X-Plex-Model":             f.config.Device.Model,
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, _ := io.ReadAll(r.Body)
	vals, _ := url.ParseQuery(string(body))
	if vals.Get("user[login]") != "user" || vals.Get("user[password]") != "pass" {
		http.Error(w, "invalid login/password", http.StatusBadRequest)
		return
	}
	// Return XML
	w.WriteHeader(http.StatusCreated)
	_ = xml.NewEncoder(w).Encode(struct {
		XMLName             xml.Name `xml:"user"`
		AuthenticationToken string   `xml:"authenticationToken,attr"`
	}{AuthenticationToken: legacyToken})
}

func (f fakeAuthServer) handlePIN(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": f.config.ClientID,
		"X-Plex-Product":           f.config.Device.Product,
		"X-Plex-Version":           f.config.Device.Version,
		"X-Plex-Platform":          f.config.Device.Platform,
		"X-Plex-Platform-Version":  f.config.Device.PlatformVersion,
		"X-Plex-Device":            f.config.Device.Device,
		"X-Plex-Device-Vendor":     f.config.Device.DeviceVendor,
		"X-Plex-Device-Name":       f.config.Device.DeviceName,
		"X-Plex-Model":             f.config.Device.Model,
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	code, id := "1234", 42
	if f.config.ClientID == "pin-timeout-test" {
		code, id = "5678", 43
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": code,
		"id":   id,
	})
}

func (f fakeAuthServer) handleValidatePIN(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v2/pins/")
	codes := map[string]string{"42": "1234"}
	code, ok := codes[id]
	if !ok {
		http.Error(w, "invalid pin id", http.StatusNotFound)
		return
	}
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": f.config.ClientID,
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"authToken": legacyToken,
		"code":      code,
	})
}

func (f fakeAuthServer) handleDevices(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/xml",
		"X-Plex-Client-Identifier": f.config.ClientID,
		"X-Plex-Token":             legacyToken,
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/xml")
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer>
  <Device product="Plex Media Server" name="srv1" token="tok-abc"></Device>
  <Device product="Plex Media Server" name="srv2" token="tok-def"></Device>
  <Device product="Other" name="client"></Device>
</MediaContainer>`)
}

func validateRequest(r *http.Request, wantHeaders map[string]string) error {
	for k, v := range wantHeaders {
		got := r.Header.Get(k)
		if v == "*" {
			if got == "" {
				return fmt.Errorf("missing header: %s", k)
			}
		} else {
			if got != v {
				return fmt.Errorf("invalid header: %s=%s", k, got)
			}
		}
	}
	return nil
}

func plexError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`{ "error": "` + msg + `" }"`))
}

type tokens struct {
	tokens map[string]string
	lock   sync.RWMutex
}

func (t *tokens) Validate(k, v string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.tokens[k] == v
}

func (t *tokens) SetToken(key, value string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.tokens[key] = value
}

func (t *tokens) CreateLegacyToken(key string) string {
	clearToken := make([]byte, 10)
	_, _ = rand.Read(clearToken)
	encodedToken := hex.EncodeToString(clearToken)
	t.SetToken(key, encodedToken)
	return encodedToken
}

// jwtHandler is a fake server that handles the JWT flows for plex.tv.
type jwtHandler struct {
	tokens  *tokens
	keySets map[string]jwk.Set
	lock    sync.Mutex
}

func (h *jwtHandler) handleJWK(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": "*",
		"X-Plex-Token":             "*",
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	clientID := r.Header.Get("X-Plex-Client-Identifier")
	token := r.Header.Get("X-Plex-Token")
	if !h.tokens.Validate(clientID, token) {
		http.Error(w, "invalid X-Plex-Token header", http.StatusUnauthorized)
		return
	}
	if _, ok := h.keySets[clientID]; ok {
		http.Error(w, "key already generated", http.StatusConflict) // not the official error code, but it's fine for our tests
		return
	}

	var resp map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	keySet, err := jwk.Parse(resp["jwk"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(keySet.Keys()) == 0 {
		http.Error(w, "invalid jwk: no key found", http.StatusBadRequest)
		return
	}
	var kid string
	if err = keySet.Get("kid", &kid); err != nil || kid == "" {
		http.Error(w, "invalid jwk: no key id found", http.StatusBadRequest)
		return
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.keySets == nil {
		h.keySets = make(map[string]jwk.Set)
	}
	h.keySets[clientID] = keySet

	w.WriteHeader(http.StatusCreated)
}

func (h *jwtHandler) handleNonce(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": "*",
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	nonce := make([]byte, 18)
	_, _ = rand.Read(nonce)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"nonce": hex.EncodeToString(nonce)})
}

func (h *jwtHandler) handleJWToken(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": "*",
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	clientID := r.Header.Get("X-Plex-Client-Identifier")

	h.lock.Lock()
	defer h.lock.Unlock()
	keySet, ok := h.keySets[clientID]
	if !ok {
		http.Error(w, "no key set found for client", http.StatusUnauthorized)
		return
	}
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t, err := jwt.Parse([]byte(req["jwt"]), jwt.WithKeySet(keySet))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: this isn't correct. auth sends back its own (signed) token.
	// this is just a quick hack to get the tests working.
	_ = t.Set(jwt.IssuedAtKey, time.Now())
	_ = t.Set(jwt.ExpirationKey, time.Now().Add(7*24*time.Hour))
	body, err := jwt.NewSerializer().Serialize(t)
	if err != nil {
		panic(err)
	}

	h.tokens.SetToken(clientID, req["jwt"])
	response := struct {
		AuthToken string `json:"auth_token"`
	}{AuthToken: string(body)}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
