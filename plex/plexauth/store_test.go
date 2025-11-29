package plexauth

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func TestStore(t *testing.T) {
	store := NewJWTDataStore("ignored", "my-passphrase")
	store.store = afero.NewMemMapFs()

	if _, err := store.Get(); !os.IsNotExist(err) {
		t.Fatalf("expected file not found error, got: %v", err)
	}

	orig := JWTSecureData{
		KeyID:      "my-key-id",
		ClientID:   "my-client-id",
		PrivateKey: []byte("my-private-key"),
	}
	if err := store.Set(orig); err != nil {
		t.Fatalf("error storing data")
	}
	want, err := store.Get()
	if err != nil {
		t.Fatalf("error retrieving data")
	}
	if !reflect.DeepEqual(orig, want) {
		t.Fatalf("stored data differs from original")
	}

	store.encryptionKey = []byte("my-passphrase")
	store.cachedData = nil
	if _, err = store.Get(); err == nil {
		t.Fatalf("data retrieved without passphrase should fail")
	}
}
