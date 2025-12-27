package plextv

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestConfig_RegisterWithPIN(t *testing.T) {
	cfg, _, ts := newTestServer(baseConfig)
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
	cfg, _, ts := newTestServer(baseConfig.WithClientID("pin-timeout-test"))
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
	cfg, _, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	// PINRequest
	pr, urlStr, err := cfg.PINRequest(ctx)
	if err != nil {
		t.Fatalf("PINRequest error: %v", err)
	}
	if pr.Code != "1234" || !strings.Contains(urlStr, "plex.tv/pin?pin=1234") || pr.Id != 42 {
		t.Fatalf("unexpected pin response: code:%+v id:%d url:%s", pr.Code, pr.Id, urlStr)
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
