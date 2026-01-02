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

func Test_plexTVTokenSource(t *testing.T) {
	tests := []struct {
		name              string
		machineIdentifier string
		plexTVClient      fakePlexTVClient
		pass              bool
		wantIDCalls       int
	}{
		{
			name:              "valid client ID",
			machineIdentifier: "pms-1-client-id",
			plexTVClient:      fakePlexTVClient{devices: []plextv.RegisteredDevice{{ClientID: "pms-1-client-id", Token: "valid-token"}}},
			pass:              true,
			wantIDCalls:       1,
		},
		{
			name:              "invalid client ID",
			machineIdentifier: "pms-2-client-id",
			plexTVClient:      fakePlexTVClient{devices: []plextv.RegisteredDevice{{ClientID: "pms-1-client-id", Token: "valid-token"}}},
			pass:              false,
			wantIDCalls:       2,
		},
		{
			name:              "pms error",
			machineIdentifier: "pms-1-client-id",
			plexTVClient:      fakePlexTVClient{err: errors.New("test failure")},
			pass:              false,
			wantIDCalls:       2,
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
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(ts.Close)

			client := NewPMSClient(ts.URL, tt.plexTVClient)

			for range 2 {
				token, err := client.tokenSource.Token(t.Context())
				if tt.pass != (err == nil) {
					t.Fatalf("unexpected err: want pass: %v, got err: %v", tt.pass, err)
				}
				if err == nil {
					if got := token.String(); got == "" {
						t.Fatal("unexpected empty token")
					}
				}
			}

			if got := int(identityCalls.Load()); got != tt.wantIDCalls {
				t.Fatalf("unexpected number of identity calls: want %d, got %d", tt.wantIDCalls, got)
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
