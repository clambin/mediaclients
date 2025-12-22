package plexauth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenSource_WithToken(t *testing.T) {
	ts := DefaultConfig.TokenSource(WithToken("abc"))
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "abc" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestTokenSource_WithCredentials(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	ts := cfg.TokenSource(WithCredentials("user", "pass"))
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != legacyToken {
		t.Fatalf("unexpected token: %s", token)
	}

	// clear the cached token
	ts.(*cachingTokenSource).token = nil
	// a failed registrar will fail the token source
	ts.(*cachingTokenSource).tokenSource = fakeRegistrar{err: errors.New("test error")}
	_, err = ts.Token(t.Context())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestTokenSource_WithPIN(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	ts := cfg.TokenSource(WithPIN(func(_ PINResponse, _ string) {}, 100*time.Millisecond))
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != legacyToken {
		t.Fatalf("unexpected token: %s", token)
	}

	// clear the cached token
	ts.(*cachingTokenSource).token = nil
	// a failed registrar will fail the token source
	ts.(*cachingTokenSource).tokenSource = fakeRegistrar{err: errors.New("test error")}
	_, err = ts.Token(t.Context())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestTokenSource_WithJWT(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	// happy path
	ts := cfg.TokenSource(
		WithCredentials("user", "pass"),
		WithJWT(filepath.Join(t.TempDir(), "vault.enc"), "my-passphrase"),
		WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))),
	)
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TODO: fakeServer should return a JWT here
	if got := token.String(); got != legacyToken {
		t.Fatalf("unexpected token: %s", got)
	}
}

func Test_jwtTokenSource(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	t.Cleanup(s.Close)

	logger := slog.New(slog.DiscardHandler) //slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	v := fakeVault{}
	ts := jwtTokenSource{
		registrar: fakeRegistrar{token: legacyToken},
		vault:     &v,
		logger:    logger,
		config:    &cfg,
	}
	ctx := t.Context()

	// happy path: no secure data, the device is registered.
	token, err := ts.Token(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := token.String(); got != legacyToken {
		t.Fatalf("unexpected token: %s", got)
	}

	// secure data exists, but it contains an invalid Client ID.
	v.err = ErrInvalidClientID
	ts = jwtTokenSource{
		registrar: fakeRegistrar{token: legacyToken},
		vault:     &v,
		logger:    logger,
		config:    &cfg,
	}

	// secure data is invalid, the device is re-registered, and a new token is returned.
	token, err = ts.Token(ctx)
	if err != nil {
		t.Fatalf("expected error, got nil")
	}
	if token.String() != legacyToken {
		t.Fatalf("unexpected token: %s", token)
	}
	if secureData := v.data.Load(); secureData.ClientID != "my-client-id" {
		t.Fatalf("unexpected client ID: %s", secureData.ClientID)
	}
}

func TestTokenSource_NoRegistrar(t *testing.T) {
	// auth server
	cfg, s := newTestServer(DefaultConfig.WithClientID("my-client-id"))
	s.Close()

	ts := cfg.TokenSource(
		WithJWT(filepath.Join(t.TempDir(), "vault.enc"), "my-passphrase"),
	)
	ctx := t.Context()
	_, err := ts.Token(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrNoTokenSource) {
		t.Fatalf("unexpected error: %v(%T)", err, err)
	}

	ts = cfg.TokenSource()
	_, err = ts.Token(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrNoTokenSource) {
		t.Fatalf("unexpected error: %v(%T)", err, err)
	}
}

var _ TokenSource = fakeRegistrar{}

type fakeRegistrar struct {
	token Token
	err   error
}

func (f fakeRegistrar) Token(_ context.Context) (Token, error) {
	return f.token, f.err
}

var _ secureDataVault = (*fakeVault)(nil)

type fakeVault struct {
	data atomic.Pointer[jwtSecureData]
	err  error
}

func (f *fakeVault) Load() (jwtSecureData, error) {
	if f.err != nil {
		return jwtSecureData{}, f.err
	}
	if data := f.data.Load(); data != nil {
		return *data, nil
	}
	return jwtSecureData{}, os.ErrNotExist
}

func (f *fakeVault) Save(data jwtSecureData) error {
	f.data.Store(&data)
	return nil
}
