package plexauth

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/clambin/mediaclients/plex/internal/vault"
)

func TestFixedTokenSource(t *testing.T) {
	ts := DefaultConfig.TokenSource().FixedToken("abc")
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
	ts := cfg.TokenSource().LegacyToken(CredentialsRegistrar{
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
	ts := cfg.TokenSource().LegacyToken(PINRegistrar{
		Config:       &cfg,
		Callback:     func(_ PINResponse, _ string) {},
		PollInterval: 100 * time.Millisecond,
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
	ts := cfg.TokenSource().PMSToken(&r, "srv1")
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
	var f fakeVault
	ts := cfg.TokenSource().PMSTokenWithJWT(&r, "srv1", "ignored.enc", "my-secret-passphrase", nil)
	ts.(*cachingTokenSource).AuthTokenSource.(*pmsTokenSource).tokenSource.(*jwtTokenSource).Vault = &f
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.String() != "tok-xyz" {
		t.Fatalf("unexpected token: %s", token)
	}

	// a failed load of secure data will fail the token source
	ts = cfg.TokenSource().PMSTokenWithJWT(&r, "srv1", "ignored.enc", "invalid-secret-passphrase", nil)
	ts.(*cachingTokenSource).AuthTokenSource.(*pmsTokenSource).tokenSource.(*jwtTokenSource).Vault = &f
	f.err = errors.New("test error")
	if _, err = ts.Token(t.Context()); err == nil {
		t.Fatalf("expected error, got nil")
	}

	// a failed registrar will fail the token source
	ts = cfg.TokenSource().PMSTokenWithJWT(&r, "srv1", "ignored.enc", "my-secret-passphrase", nil)
	ts.(*cachingTokenSource).AuthTokenSource.(*pmsTokenSource).tokenSource.(*jwtTokenSource).Vault = &f
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
	return jwtSecureData{}, vault.ErrNotFound
}

func (f *fakeVault) Save(data jwtSecureData) error {
	f.data.Store(&data)
	return nil
}
