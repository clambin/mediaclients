package plextv

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConfig_WithClientIDAndDevice(t *testing.T) {
	cfg := DefaultConfig().WithClientID("abc").WithDevice(Device{Product: "X"})
	if cfg.ClientID != "abc" {
		t.Fatalf("expected client id to be set")
	}
	if cfg.Device.Product != "X" {
		t.Fatalf("expected device to be set")
	}
}

func TestConfig_BadData(t *testing.T) {
	cfg, _, ts := newTestServer(baseConfig)
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
}
