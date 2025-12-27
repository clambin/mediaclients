package plextv

import (
	"errors"
	"testing"
)

func TestConfig_GenerateAndUploadPublicKey(t *testing.T) {
	cfg, s, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	token := s.tokens.CreateLegacyToken(cfg.ClientID)

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, Token(token))
	if err != nil {
		t.Fatalf("GenerateAndUploadPublicKey error: %v", err)
	}
	if len(privateKey) != 64 {
		t.Fatalf("unexpected key length: %d", len(privateKey))
	}
	if keyID == "" {
		t.Fatalf("expected non-empty key id")
	}

	// bad token
	if _, _, err = cfg.GenerateAndUploadPublicKey(ctx, "bad-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token error")
	}

	// invalid token
	if _, _, err := cfg.GenerateAndUploadPublicKey(ctx, ""); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken from empty token, got: %v", err)
	}

	// errors
	ts.Close()
	if _, _, err = cfg.GenerateAndUploadPublicKey(ctx, Token(token)); err == nil {
		t.Fatalf("expected error from closed server")
	}
}

func TestConfig_JWTToken(t *testing.T) {
	cfg, s, ts := newTestServer(baseConfig)
	t.Cleanup(ts.Close)
	ctx := t.Context()

	token := s.tokens.CreateLegacyToken(cfg.ClientID)

	privateKey, keyID, err := cfg.GenerateAndUploadPublicKey(ctx, Token(token))
	if err != nil {
		t.Fatalf("GenerateAndUploadPublicKey error: %v", err)
	}

	tok, err := cfg.JWTToken(ctx, privateKey, keyID)
	if err != nil {
		t.Fatalf("JWTToken error: %v", err)
	}
	if !tok.IsJWT() {
		t.Fatal("expected JWT token")
	}

	// errors
	ts.Close()
	if _, err = cfg.JWTToken(ctx, privateKey, keyID); err == nil {
		t.Fatalf("expected error from closed server")
	}
}
