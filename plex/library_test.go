package plex_test

import (
	"context"
	"testing"

	"github.com/clambin/mediaclients/plex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetLibraries(t *testing.T) {
	c, testServer := makeClientAndServer(nil)
	t.Cleanup(testServer.Close)

	libraries, err := c.GetLibraries(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []plex.Library{
		{Key: "1", Type: "movie", Title: "Movies"},
		{Key: "2", Type: "show", Title: "Shows"},
	}, libraries)
}

func TestClient_GetAllLibraryMedia(t *testing.T) {
	c, testServer := makeClientAndServer(nil)
	t.Cleanup(testServer.Close)

	var media []plex.MediaMetadata
	for entry, err := range c.GetAllLibraryMedia(context.Background(), "1") {
		require.NoError(t, err)
		media = append(media, entry)
	}
	assert.Equal(t, []plex.MediaMetadata{
		{Title: "baz", Guid: "1"},
	}, media)
}
