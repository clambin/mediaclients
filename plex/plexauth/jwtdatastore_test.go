package plexauth

import (
	"errors"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
)

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
