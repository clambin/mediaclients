package plexauth

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/clambin/mediaclients/plex/internal/vault"
)

func TestJWTDataStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.enc")

	s1 := newJWTDataStore(path, "my-secure-passphrase", "my-client-id")
	want := jwtSecureData{
		KeyID:      "my-key-id",
		ClientID:   "my-client-id",
		PrivateKey: []byte("my-private-key"),
	}
	err := s1.Save(want)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := s1.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() want: %+v, got: %+v", want, got)
	}

	s2 := newJWTDataStore(path, "invalid-passphrase", "my-client-id")
	_, err = s2.Load()
	var errInvalidKey *vault.ErrDecryptionFailed
	if !errors.As(err, &errInvalidKey) {
		t.Fatalf("expected error, got %#+v", err)
	}

	s3 := newJWTDataStore(path, "my-secure-passphrase", "invalid-client-id")
	got, err = s3.Load()
	if !errors.Is(err, ErrInvalidClientID) {
		t.Fatalf("expected error, got nil")
	}
	if got.ClientID != "my-client-id" {
		t.Fatalf("unexpected client ID: %s", got.ClientID)
	}
}
