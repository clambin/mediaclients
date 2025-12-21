package plexauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConfig_WithClientIDAndDevice(t *testing.T) {
	cfg := DefaultConfig.WithClientID("abc").WithDevice(Device{Product: "X"})
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
	ctx := WithHTTPClient(t.Context(), &http.Client{Timeout: 10 * time.Second})

	tok, err := cfg.RegisterWithCredentials(ctx, "user", "pass")
	if err != nil {
		t.Fatalf("RegisterWithCredentials error: %v", err)
	}
	if tok.String() != legacyToken {
		t.Fatalf("unexpected token: %s", tok)
	}

	// errors
	ts.Close()
	if _, err = cfg.RegisterWithCredentials(ctx, "user", "pass"); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestConfig_RegisterWithPIN(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)

	// RegisterWithPIN polls until token available
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	t.Cleanup(cancel)
	tok2, err := cfg.RegisterWithPIN(ctx, func(resp PINResponse, url string) {}, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("RegisterWithPIN error: %v", err)
	}
	if tok2.String() != legacyToken {
		t.Fatalf("unexpected token: %s", tok2)
	}

	// errors
	ts.Close()
	if _, err = cfg.RegisterWithPIN(t.Context(), func(resp PINResponse, url string) {}, 10*time.Millisecond); err == nil {
		t.Fatalf("expected error from closed server")
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

	// errors
	ts.Close()
	if _, err = cfg.RegisterWithPIN(t.Context(), func(resp PINResponse, url string) {}, time.Minute); err == nil {
		t.Fatalf("expected error from closed server")
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
	if pr.Code != "1234" || !strings.Contains(urlStr, "plex.tv/pin?pin=1234") || pr.Id != 42 {
		t.Fatalf("unexpected pin response/url: %+v %s", pr, urlStr)
	}

	// ValidatePIN
	tok, resp, err := cfg.ValidatePIN(ctx, pr.Id)
	if err != nil {
		t.Fatalf("ValidatePIN error: %v", err)
	}
	if resp.Code != "1234" {
		t.Fatalf("unexpected code: %s", resp.Code)
	}
	if tok.String() != legacyToken {
		t.Fatalf("unexpected token: %s", tok)
	}

	// Invalid Id
	if _, _, err = cfg.ValidatePIN(ctx, 43); err == nil {
		t.Fatalf("expected error from invalid pin id")
	}

	// errors
	ts.Close()
	if _, _, err = cfg.PINRequest(ctx); err == nil {
		t.Fatalf("expected error from closed server")
	}
	if _, _, err = cfg.ValidatePIN(ctx, pr.Id); err == nil {
		t.Fatalf("expected error from closed server")
	}
}
func TestConfig_GenerateAndUploadPublicKey(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, legacyToken)
	if err != nil {
		t.Fatalf("GenerateAndUploadPublicKey error: %v", err)
	}
	if len(privateKey) != 64 {
		t.Fatalf("unexpected key length: %d", len(privateKey))
	}
	if keyID == "" {
		t.Fatalf("expected non-empty key id")
	}

	if _, _, err = cfg.GenerateAndUploadPublicKey(ctx, "bad-token"); err == nil {
		t.Fatalf("expected invalid token error")
	}

	// errors
	ts.Close()
	if _, _, err = cfg.GenerateAndUploadPublicKey(ctx, legacyToken); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestConfig_JWTToken(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, legacyToken)
	if err != nil {
		t.Fatalf("GenerateAndUploadPublicKey error: %v", err)
	}

	tok, err := cfg.JWTToken(ctx, privateKey, keyID)
	if err != nil {
		t.Fatalf("JWTToken error: %v", err)
	}
	if got := tok.String(); got != legacyToken {
		t.Fatalf("unexpected token: %s, want: %s", got, legacyToken)
	}

	// errors
	ts.Close()
	if _, err = cfg.JWTToken(ctx, privateKey, keyID); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestConfig_RegisteredDevices_And_MediaServers(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	devs, err := cfg.RegisteredDevices(ctx, legacyToken)
	if err != nil {
		t.Fatalf("RegisteredDevices error: %v", err)
	}
	if len(devs) != 3 {
		t.Fatalf("expected 2 devices, got %d", len(devs))
	}
	servers, err := cfg.MediaServers(context.Background(), legacyToken)
	if err != nil {
		t.Fatalf("MediaServers error: %v", err)
	}
	if len(servers) != 2 || servers[0].Name != "srv1" || servers[1].Name != "srv2" {
		t.Fatalf("unexpected servers: %+v", servers)
	}

	// errors
	ts.Close()
	if _, err = cfg.RegisteredDevices(ctx, legacyToken); err == nil {
		t.Fatalf("expected error from closed server")
	}
	if _, err = cfg.MediaServers(context.Background(), legacyToken); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestConfig_BadData(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	ts.Close()
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("bad data"))
	}))
	cfg.URL = ts.URL
	cfg.V2URL = ts.URL
	ctx := t.Context()
	if _, err := cfg.RegisterWithCredentials(ctx, "user", "pass"); err == nil {
		t.Fatalf("expected error from bad data")
	}
	if _, err := cfg.RegisterWithPIN(ctx, func(PINResponse, string) {}, 10*time.Millisecond); err == nil {
		t.Fatalf("expected error from bad data")
	}
	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, legacyToken)
	if err == nil {
		t.Fatalf("expected error from bad data")
	}
	if _, err = cfg.JWTToken(ctx, privateKey, keyID); err == nil {
		t.Fatalf("expected error from bad data")
	}
	if _, err = cfg.RegisteredDevices(ctx, legacyToken); err == nil {
		t.Fatalf("expected error from bad data")
	}
}

func TestConfig_BadToken(t *testing.T) {
	cfg, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	if _, _, err := cfg.GenerateAndUploadPublicKey(ctx, ""); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken from empty token, got: %v", err)
	}
	if _, err := cfg.RegisteredDevices(ctx, ""); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken from empty token, got: %v", err)
	}
}
