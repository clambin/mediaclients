package vault

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/afero"
)

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
	t.Helper()

	f := afero.NewMemMapFs()
	c := newWithFS[T](f, "vault.enc", "my-passphrase")

	if err := c.Save(v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := afero.ReadFile(f, "vault.enc")
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
