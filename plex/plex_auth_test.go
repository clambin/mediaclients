package plex

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/clambin/mediaclients/plex/plextv"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		machineIdentifier string
		plexTVClient      fakePlexTVClient
		pass              bool
	}{
		{
			name:              "valid client ID",
			machineIdentifier: "pms-1-client-id",
			plexTVClient:      fakePlexTVClient{devices: []plextv.RegisteredDevice{{ClientID: "pms-1-client-id", Token: "valid-token"}}},
			pass:              true,
		},
		{
			name:              "invalid client ID",
			machineIdentifier: "pms-2-client-id",
			plexTVClient:      fakePlexTVClient{devices: []plextv.RegisteredDevice{{ClientID: "pms-1-client-id", Token: "valid-token"}}},
			pass:              false,
		},
		{
			name:              "pms error",
			machineIdentifier: "pms-1-client-id",
			plexTVClient:      fakePlexTVClient{err: errors.New("test failure")},
			pass:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var identityCalls atomic.Int32
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/identity":
					identityCalls.Add(1)
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{ "MediaContainer": { "machineIdentifier": "` + tt.machineIdentifier + `" } }"`))
				default:
					if token := r.Header.Get("X-Plex-Token"); token != "valid-token" {
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					}
				}

			}))

			client := &http.Client{
				Transport: &authMiddleware{
					next:         http.DefaultTransport,
					httpClient:   &http.Client{},
					plexTVClient: tt.plexTVClient,
					url:          ts.URL,
				},
			}

			for range 3 {
				req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/", nil)
				resp, err := client.Do(req)
				if err != nil {
					if tt.pass {
						t.Fatalf("unexpected error: %v", err)
					}
					return
				}
				_ = resp.Body.Close()
				if (resp.StatusCode != http.StatusOK && tt.pass) || (resp.StatusCode == http.StatusOK && !tt.pass) {
					t.Fatalf("unexpected status code: %d", resp.StatusCode)
				}
			}

			if got := identityCalls.Load(); got != 1 {
				t.Fatalf("expected only one call, got %v", got)
			}
		})
	}
}

var _ PlexTVClient = fakePlexTVClient{}

type fakePlexTVClient struct {
	devices []plextv.RegisteredDevice
	err     error
}

func (f fakePlexTVClient) User(_ context.Context) (plextv.User, error) {
	return plextv.User{}, f.err
}

func (f fakePlexTVClient) MediaServers(_ context.Context) ([]plextv.RegisteredDevice, error) {
	return f.devices, f.err
}
