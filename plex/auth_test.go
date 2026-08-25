package plex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clambin/mediaclients/plex/plextv"
)

func TestPMSToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity" {
			http.Error(w, "invalid path", http.StatusNotFound)
			return
		}
		var response struct {
			MediaContainer Identity
		}

		response.MediaContainer = Identity{MachineIdentifier: "my-pms-server-id"}
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	ptvc := fakePlexTVClient{
		devices: []plextv.Device{{ClientIdentifier: "my-pms-server-id", Token: "my-pms-server-token"}},
	}

	token, err := Token(context.Background(), ts.URL, ptvc)
	if err != nil {
		t.Errorf("PMSToken: %v", err)
	}
	if token != "my-pms-server-token" {
		t.Errorf("PMSToken: expected token %s, got %s", "my-pms-server-token", token)
	}
}

var _ PlexTVClient = fakePlexTVClient{}

type fakePlexTVClient struct {
	devices []plextv.Device
	err     error
}

func (f fakePlexTVClient) User(_ context.Context) (plextv.User, error) {
	return plextv.User{}, f.err
}

func (f fakePlexTVClient) MediaServers(_ context.Context) ([]plextv.Device, error) {
	return f.devices, f.err
}
