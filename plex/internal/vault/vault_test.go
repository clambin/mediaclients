package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestErrDecryptionFailed(t *testing.T) {
	err := &ErrDecryptionFailed{}
	if got := err.Error(); got != "invalid key" {
		t.Fatalf("unexpected error string: want %v, got %v", "invalid key", got)
	}
	err = &ErrDecryptionFailed{Err: os.ErrNotExist}
	if got := err.Error(); got != "invalid key: file does not exist" {
		t.Fatalf("unexpected error string: want %v, got %v", "invalid key: file does not exist", got)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: want %v, got %v", os.ErrNotExist, err)
	}

	err2 := fmt.Errorf("test fail: %w", &ErrDecryptionFailed{Err: errors.New("test fail")})
	var err3 *ErrDecryptionFailed
	if !errors.As(err2, &err3) {
		t.Fatalf("errors.As failed")
	}
}

func TestVault(t *testing.T) {
	doTest[int](t, 123)
	doTest[string](t, "hello world")
	doTest[float64](t, 123.456)
	type dataRecord struct {
		Int      int           `json:"int"`
		String   string        `json:"string"`
		Float    float64       `json:"float"`
		Duration time.Duration `json:"duration"`
	}
	doTest[dataRecord](t, dataRecord{Int: 123, String: "hello world", Float: 123.456, Duration: time.Second * 123})
}

func doTest[T any](t *testing.T, v T) {
	//t.Helper()

	filename := filepath.Join(t.TempDir(), "vault.enc")
	c := New[T](filename, "my-passphrase")

	if err := c.Save(v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// clear the cache
	c.content = nil

	got, err := c.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, v) {
		t.Fatalf("Load() want: %+v, got: %+v", v, got)
	}
}
