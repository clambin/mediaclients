package plextv

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenSource_WithToken(t *testing.T) {
	ts := DefaultConfig().TokenSource(WithToken("abc"))
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
	cfg, _, s := newTestServer(DefaultConfig().WithClientID("my-client-id"))
	t.Cleanup(s.Close)
	doTokenSourceTest(t, cfg.TokenSource(WithCredentials("user", "pass")), legacyToken)
}

func TestTokenSource_WithPIN(t *testing.T) {
	// auth server
	cfg, _, s := newTestServer(DefaultConfig().WithClientID("my-client-id"))
	t.Cleanup(s.Close)
	doTokenSourceTest(t, cfg.TokenSource(WithPIN(func(_ PINResponse, _ string) {}, 100*time.Millisecond)), legacyToken)
}

func TestTokenSource_WithJWT(t *testing.T) {
	// auth server
	cfg, s, tts := newTestServer(DefaultConfig().WithClientID("my-client-id"))
	t.Cleanup(tts.Close)

	s.tokens.SetToken("my-client-id", legacyToken)
	var f fakeVault
	ts := cfg.TokenSource(
		WithCredentials("user", "pass"),
		WithJWT(&f),
		WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))),
	)
	doTokenSourceTest(t, ts, "")
}

func doTokenSourceTest(t *testing.T, ts TokenSource, want string) {
	t.Helper()
	// happy path
	token, err := ts.Token(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want != "" {
		if got := token.String(); got != want {
			t.Fatalf("unexpected token: want=%v, got=%v", want, token)
		}
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

func Test_jwtTokenSource(t *testing.T) {
	// auth server
	cfg, s, tts := newTestServer(DefaultConfig().WithClientID("my-client-id"))
	t.Cleanup(tts.Close)

	s.tokens.SetToken("my-client-id", legacyToken)

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
	if !token.IsJWT() {
		t.Fatal("expected JWT token")
	}
}

func Test_jwtTokenSource_WrongClientID(t *testing.T) {
	// auth server
	cfg, s, tts := newTestServer(DefaultConfig().WithClientID("my-client-id"))
	t.Cleanup(tts.Close)

	s.tokens.SetToken("my-client-id", legacyToken)

	v := fakeVault{err: ErrInvalidClientID}
	ts := jwtTokenSource{
		registrar: fakeRegistrar{token: legacyToken},
		vault:     &v,
		logger:    slog.New(slog.DiscardHandler),
		config:    &cfg,
	}
	ctx := t.Context()

	token, err := ts.Token(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !token.IsJWT() {
		t.Fatal("expected JWT token")
	}
	if secureData := v.data.Load(); secureData.ClientID != "my-client-id" {
		t.Fatalf("unexpected client ID: %s", secureData.ClientID)
	}
}

func TestTokenSource_NoRegistrar(t *testing.T) {
	// auth server
	cfg, _, s := newTestServer(DefaultConfig().WithClientID("my-client-id"))
	s.Close()

	var f fakeVault
	ts := cfg.TokenSource(
		WithJWT(&f),
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

func BenchmarkCachingTokenSource_Token_LegacyToken(b *testing.B) {
	b.Run("legacy", func(b *testing.B) {
		ts := cachingTokenSource{tokenSource: fakeRegistrar{token: legacyToken}}
		ctx := context.Background()
		b.ReportAllocs()
		for b.Loop() {
			_, err := ts.Token(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
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

func TestJWTDataStore(t *testing.T) {
	var v fakeVault
	s := jwtDataStore{vault: &v, clientID: "my-client-id"}
	if _, err := s.Load(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}

	want := JWTSecureData{
		KeyID:      "my-key-id",
		ClientID:   "my-client-id",
		PrivateKey: []byte("my-private-key"),
	}
	if err := s.Save(want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() want: %+v, got: %+v", want, got)
	}

	// store the wrong client ID
	data, _ := v.Load()
	data.ClientID = "invalid-client-id"
	_ = v.Save(data)

	got, err = s.Load()
	if !errors.Is(err, ErrInvalidClientID) {
		t.Fatalf("expected error, got nil")
	}
}

type fakeVault struct {
	data atomic.Pointer[JWTSecureData]
	err  error
}

func (f *fakeVault) Load() (JWTSecureData, error) {
	if f.err != nil {
		return JWTSecureData{}, f.err
	}
	if data := f.data.Load(); data != nil {
		return *data, nil
	}
	return JWTSecureData{}, os.ErrNotExist
}

func (f *fakeVault) Save(data JWTSecureData) error {
	f.data.Store(&data)
	return nil
}
