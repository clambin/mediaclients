package plexauth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestFixedTokenSource(t *testing.T) {
	ts := NewFixedTokenSource("abc")
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "abc" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestLegacyTokenSourceWithCredentials(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	ts := NewLegacyTokenSource(CredentialsRegistrar{
		Config:   &cfg,
		Username: "user",
		Password: "pass",
	})
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "tok123" {
		t.Fatalf("unexpected token: %s", token)
	}

	// clear the cached token
	ts.(*cachingTokenSource).authToken = ""
	// a failed registrar will fail the token source
	ts.(*cachingTokenSource).AuthTokenSource.(*legacyTokenSource).Registrar = fakeRegistrar{err: errors.New("test error")}
	_, err = ts.Token(t.Context())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLegacyTokenSourceWithPIN(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	ts := NewLegacyTokenSource(PINRegistrar{
		Config:   &cfg,
		Callback: func(_ PINResponse, _ string) {},
	})
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "tok-abc" {
		t.Fatalf("unexpected token: %s", token)
	}

	// clear the cached token
	ts.(*cachingTokenSource).authToken = ""
	// a failed registrar will fail the token source
	ts.(*cachingTokenSource).AuthTokenSource.(*legacyTokenSource).Registrar = fakeRegistrar{err: errors.New("test error")}
	_, err = ts.Token(t.Context())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestPMSTokenStore(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	r := fakeRegistrar{authToken: AuthToken("tok-abc"), err: nil}
	ts := NewPMSTokenStore(cfg, &r, "srv1")
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "tok-xyz" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestPMSTokenSourceWithJWT(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	r := fakeRegistrar{authToken: AuthToken("tok-abc"), err: nil}
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "token-data.enc")
	ts := NewPMSTokenSourceWithJWT(cfg, &r, "srv1", storePath, "my-secret-passphrase", nil)
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "tok-xyz" {
		t.Fatalf("unexpected token: %s", token)
	}

	// a failed load of secure data will fail the token source
	ts = NewPMSTokenSourceWithJWT(cfg, &r, "srv1", storePath, "invalid-secret-passphrase", nil)
	if _, err = ts.Token(t.Context()); err == nil {
		t.Fatalf("expected error, got nil")
	}

	// a failed registrar will fail the token source
	storePath = filepath.Join(tempDir, "token-data-2.enc")
	ts = NewPMSTokenSourceWithJWT(cfg, &r, "srv1", storePath, "my-secret-passphrase", nil)
	r.err = errors.New("test error")
	if _, err = ts.Token(t.Context()); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

var _ Registrar = fakeRegistrar{}

type fakeRegistrar struct {
	authToken AuthToken
	err       error
}

func (f fakeRegistrar) Register(_ context.Context) (AuthToken, error) {
	return f.authToken, f.err
}
