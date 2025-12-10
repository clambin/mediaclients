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

func TestVault_Types(t *testing.T) {
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

func TestVault_Errors(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "vault.enc")
	c := New[string](filename, "my-passphrase")

	// loading an empty should return os.ErrNotExist
	if _, err := c.Load(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}

	if err := c.Save("hello world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c.content = nil
	if value, err := c.Load(); err != nil || value != "hello world" {
		t.Fatalf("unexpected value: %v, %v", value, err)
	}

	// load should fail on invalid passphrase
	c.content = nil
	c.passphrase = "invalid-passphrase"
	var errDecryptionFailed *ErrDecryptionFailed
	if _, err := c.Load(); !errors.As(err, &errDecryptionFailed) {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}

	// load should fail on invalid file content
	if err := os.WriteFile(filename, []byte("invalid-content"), 0600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := c.Load(); err == nil {
		t.Fatalf("expected error")
	}

	// version is mandatory in file
	if err := os.WriteFile(filename, []byte("{ }"), 0600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := c.Load(); err == nil {
		t.Fatalf("expected error")
	}
}

func doTest[T any](t *testing.T, v T) {
	filename := filepath.Join(t.TempDir(), "vault.enc")
	c := New[T](filename, "my-passphrase")

	// loading an empty should return
	if _, err := c.Load(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}

	if err := c.Save(v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// clear the cache
	c.content = nil

	for range 2 {
		got, err := c.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("Load() want: %+v, got: %+v", v, got)
		}
	}
}
