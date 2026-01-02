package plextv

import (
	"net/http"
	"testing"
	"time"
)

func TestConfig_RegisterWithCredentials(t *testing.T) {
	cfg, _, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := ContextWithHTTPClient(t.Context(), &http.Client{Timeout: 10 * time.Second})

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
