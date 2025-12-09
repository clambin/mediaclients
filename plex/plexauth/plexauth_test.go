package plexauth

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

var baseConfig = DefaultConfig.
	WithClientID("abc").
	WithClientDevice(ClientDevice{
		Product:         "TestProduct",
		Version:         "1.0",
		Platform:        "unit",
		PlatformVersion: "test",
		Device:          "dev",
		DeviceVendor:    "vendor",
		DeviceName:      "devname",
		Model:           "model",
	})

func newTestServer(cfg Config) (Config, *httptest.Server) {
	ts := httptest.NewServer(makeFakeServer(&cfg))
	cfg.AuthURL = ts.URL
	cfg.AuthV2URL = ts.URL
	return cfg, ts
}

func TestConfig_WithClientIDAndDevice(t *testing.T) {
	cfg := DefaultConfig.WithClientID("abc").WithClientDevice(ClientDevice{Product: "X"})
	if cfg.ClientID != "abc" {
		t.Fatalf("expected client id to be set")
	}
	if cfg.Device.Product != "X" {
		t.Fatalf("expected device to be set")
	}
}

func TestConfig_RegisterWithCredentials(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	tok, err := cfg.RegisterWithCredentials(ctx, "user", "pass")
	if err != nil {
		t.Fatalf("RegisterWithCredentials error: %v", err)
	}
	if tok.String() != "tok-abc" {
		t.Fatalf("unexpected token: %s", tok)
	}
}

func TestConfig_RegisterWithPIN(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)

	// RegisterWithPIN should poll until token available
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	t.Cleanup(cancel)
	tok2, err := cfg.RegisterWithPIN(ctx, func(resp PINResponse, url string) {}, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("RegisterWithPIN error: %v", err)
	}
	if tok2.String() != "tok-abc" {
		t.Fatalf("unexpected token: %s", tok2)
	}
}

func TestConfig_RegisterWithPIN_Timeout(t *testing.T) {
	cfg, ts := newTestServer(baseConfig.WithClientID("pin-timeout-test"))
	t.Cleanup(ts.Close)

	// RegisterWithPIN should poll until token available
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	t.Cleanup(cancel)
	_, err := cfg.RegisterWithPIN(ctx, func(resp PINResponse, url string) {}, 10*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestConfig_PINRequest_And_ValidatePIN(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	// PINRequest
	pr, urlStr, err := cfg.PINRequest(ctx)
	if err != nil {
		t.Fatalf("PINRequest error: %v", err)
	}
	if pr.Code != "1234" || !strings.Contains(urlStr, "plex.tv/pin?pin=1234") {
		t.Fatalf("unexpected pin response/url: %+v %s", pr, urlStr)
	}
	if pr.Id != 42 {
		t.Fatalf("unexpected pin id: %d", pr.Id)
	}

	// ValidatePIN first without token then with token
	tok, resp, err := cfg.ValidatePIN(ctx, pr.Id)
	if err != nil {
		t.Fatalf("ValidatePIN error: %v", err)
	}
	if resp.Code != "1234" {
		t.Fatalf("unexpected code: %s", resp.Code)
	}
	if tok.String() != "tok-abc" {
		t.Fatalf("unexpected token: %s", tok)
	}

}
func TestConfig_GenerateAndUploadPublicKey(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, "tok-abc")
	if err != nil {
		t.Fatalf("GenerateAndUploadPublicKey error: %v", err)
	}
	if len(privateKey) != 64 {
		t.Fatalf("unexpected key length: %d", len(privateKey))
	}
	if keyID == "" {
		t.Fatalf("expected non-empty key id")
	}

	_, _, err = cfg.GenerateAndUploadPublicKey(ctx, "bad-token")
	if err == nil {
		t.Fatalf("expected invalid token error")
	}
}

func TestConfig_JWTToken(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, "tok-abc")
	if err != nil {
		t.Fatalf("GenerateAndUploadPublicKey error: %v", err)
	}

	tok, err := cfg.JWTToken(ctx, privateKey, keyID)
	if err != nil {
		t.Fatalf("JWTToken error: %v", err)
	}
	if got := tok.String(); got != "tok-abc" {
		t.Fatalf("unexpected token: %s, want: %s", got, "tok-abc")
	}
}

func TestConfig_RegisteredDevices_And_MediaServers(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	devs, err := cfg.RegisteredDevices(ctx, AuthToken("tok-abc"))
	if err != nil {
		t.Fatalf("RegisteredDevices error: %v", err)
	}
	if len(devs) != 3 {
		t.Fatalf("expected 2 devices, got %d", len(devs))
	}
	servers, err := cfg.MediaServers(context.Background(), AuthToken("tok-abc"))
	if err != nil {
		t.Fatalf("MediaServers error: %v", err)
	}
	if len(servers) != 2 || servers[0].Name != "srv1" || servers[1].Name != "srv2" {
		t.Fatalf("unexpected servers: %+v", servers)
	}
}

var _ http.Handler = &fakeServer{}

type fakeServer struct {
	http.Handler
	config *Config
}

func makeFakeServer(cfg *Config) fakeServer {
	f := fakeServer{config: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/sign_in.xml", f.handleRegisterWithCredentials)
	mux.HandleFunc("POST /api/v2/pins", f.handlePIN)
	mux.HandleFunc("GET /api/v2/pins/", f.handleValidatePIN)
	mux.HandleFunc("POST /api/v2/auth/jwk", f.handleJWK)
	mux.HandleFunc("GET /api/v2/auth/nonce", f.handleNonce)
	mux.HandleFunc("POST /api/v2/auth/token", f.handleJWToken)
	mux.HandleFunc("GET /devices.xml", f.handleDevices)
	f.Handler = mux
	return f
}

func (f fakeServer) handleRegisterWithCredentials(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Content-Type":               "application/x-www-form-urlencoded",
		"Accept":                     "application/xml",
		"X-Plex-Client-Identifier":   f.config.ClientID,
		"X-Plex-Product":             f.config.Device.Product,
		"X-Plex-Version":             f.config.Device.Version,
		"X-Plex-Platform":            f.config.Device.Platform,
		"X-Plex-Platform-Version":    f.config.Device.PlatformVersion,
		"X-Plex-ClientDevice":        f.config.Device.Device,
		"X-Plex-ClientDevice-Vendor": f.config.Device.DeviceVendor,
		"X-Plex-ClientDevice-Name":   f.config.Device.DeviceName,
		"X-Plex-Model":               f.config.Device.Model,
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
	}{AuthenticationToken: "tok-abc"})
}

func (f fakeServer) handlePIN(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                     "application/json",
		"X-Plex-Client-Identifier":   f.config.ClientID,
		"X-Plex-Product":             f.config.Device.Product,
		"X-Plex-Version":             f.config.Device.Version,
		"X-Plex-Platform":            f.config.Device.Platform,
		"X-Plex-Platform-Version":    f.config.Device.PlatformVersion,
		"X-Plex-ClientDevice":        f.config.Device.Device,
		"X-Plex-ClientDevice-Vendor": f.config.Device.DeviceVendor,
		"X-Plex-ClientDevice-Name":   f.config.Device.DeviceName,
		"X-Plex-Model":               f.config.Device.Model,
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

func (f fakeServer) handleValidatePIN(w http.ResponseWriter, r *http.Request) {
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
		"authToken": "tok-abc",
		"code":      code,
	})
}

func (f fakeServer) handleJWK(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": f.config.ClientID,
		"X-Plex-Token":             "tok-abc",
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	req := make(map[string]any)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jwk, ok := req["jwk"].(map[string]any)
	if !ok {
		http.Error(w, "missing jwk", http.StatusBadRequest)
		return
	}
	for _, attrib := range []string{"alg", "crv", "kid", "kty", "use"} {
		if value, ok := jwk[attrib].(string); !ok || value == "" {
			http.Error(w, "missing jwt attribute: "+attrib, http.StatusBadRequest)
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (f fakeServer) handleNonce(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": f.config.ClientID,
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"nonce": "01234567890"})
}

func (f fakeServer) handleJWToken(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/json",
		"X-Plex-Client-Identifier": f.config.ClientID,
	}
	if err := validateRequest(r, wantHeaders); err != nil {
		plexError(w, http.StatusBadRequest, err.Error())
		return
	}
	var request map[string]string
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	signedToken, ok := request["jwt"]
	if !ok {
		http.Error(w, "missing jwt", http.StatusBadRequest)
		return
	}
	tok, err := jwt.Parse([]byte(signedToken), jwt.WithVerify(false))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if aud, ok := tok.Audience(); !ok || len(aud) == 0 || aud[0] != "plex.tv" {
		http.Error(w, "audience missing/invalid", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"auth_token": "tok-abc"})
}

func (f fakeServer) handleDevices(w http.ResponseWriter, r *http.Request) {
	wantHeaders := map[string]string{
		"Accept":                   "application/xml",
		"X-Plex-Client-Identifier": f.config.ClientID,
		"X-Plex-Token":             "tok-abc",
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
		if r.Header.Get(k) != v {
			return fmt.Errorf("invalid/missing header: %s=%s", k, r.Header.Get(k))
		}
	}
	return nil
}

func plexError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`{ "error": "` + msg + `" }"`))
}
