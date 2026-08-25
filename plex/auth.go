package plex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/clambin/mediaclients/plex/plextv"
)

type PlexTVClient interface {
	User(ctx context.Context) (plextv.User, error)
	MediaServers(ctx context.Context) ([]plextv.Device, error)
}

// Token returns a (legacy) Plex Media Server token for the server at the provided URL.
// client is the plextv.Client used to retrieve the device information from the PlexTV API.
func Token(ctx context.Context, target string, client PlexTVClient) (string, error) {
	id, err := pmsIdentifier(ctx, target)
	if err != nil {
		return "", fmt.Errorf("pmsIdentifier: %w", err)
	}

	token, err := deviceToken(ctx, id, client)
	if err != nil {
		return "", fmt.Errorf("deviceToken: %w", err)
	}

	// call User() to update the client device information stored in the PlexTV configuration
	// if it fails, we don't care
	_, _ = client.User(ctx)

	return token, nil
}

func pmsIdentifier(ctx context.Context, target string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target+"/identity", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http: %d - %s", resp.StatusCode, resp.Status)
	}

	var response struct {
		MediaContainer Identity
	}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return response.MediaContainer.MachineIdentifier, nil
}

func deviceToken(ctx context.Context, id string, client PlexTVClient) (string, error) {
	devices, err := client.MediaServers(ctx)
	if err != nil {
		return "", fmt.Errorf("devices: %w", err)
	}

	for _, device := range devices {
		if device.ClientIdentifier == id {
			// We don't need the user information here.
			// This just updates the device information with the PlexTV services.
			return device.Token, nil
		}
	}

	return "", errors.New("device not found")
}
