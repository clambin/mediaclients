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

func TestClient_GetMovies(t *testing.T) {
	c, s := makeClientAndServer(nil)
	t.Cleanup(s.Close)

	movies, err := c.GetMovies(context.Background(), "1")
	require.NoError(t, err)
	assert.Equal(t, []plex.Movie{{Guid: "1", Title: "foo"}}, movies)
}

func TestClient_GetShows(t *testing.T) {
	c, s := makeClientAndServer(nil)
	t.Cleanup(s.Close)

	shows, err := c.GetShows(context.Background(), "2")
	require.NoError(t, err)
	assert.Equal(t, []plex.Show{{Guid: "2", Title: "bar"}}, shows)
}

func TestClient_GetSeasons(t *testing.T) {
	c, s := makeClientAndServer(nil)
	t.Cleanup(s.Close)

	shows, err := c.GetSeasons(context.Background(), "200")
	require.NoError(t, err)
	assert.Equal(t, []plex.Season{{Guid: "2", Title: "Season 1"}}, shows)
}

func TestClient_GetEpisodes(t *testing.T) {
	c, s := makeClientAndServer(nil)
	t.Cleanup(s.Close)

	shows, err := c.GetEpisodes(context.Background(), "201")
	require.NoError(t, err)
	assert.Equal(t, []plex.Episode{{Guid: "2", Title: "Episode 1"}}, shows)
}
