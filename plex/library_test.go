package plex_test

import (
	"context"
	"github.com/clambin/mediaclients/plex"
	"github.com/clambin/mediaclients/plex/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetLibraries(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(testutil.Handler))
	defer testServer.Close()

	c := plex.New("user@example.com", "somepassword", "", "", testServer.URL, nil)
	c.HTTPClient.Transport = http.DefaultTransport

	libraries, err := c.GetLibraries(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []plex.Library{
		{Key: "1", Type: "movie", Title: "Movies"},
		{Key: "2", Type: "show", Title: "Shows"},
	}, libraries)
}

func TestClient_GetMovies(t *testing.T) {
	c, s := makeClientAndServer(nil)
	defer s.Close()

	movies, err := c.GetMovies(context.Background(), "1")
	require.NoError(t, err)
	assert.Equal(t, []plex.Movie{{Guid: "1", Title: "foo"}}, movies)
}

func TestClient_GetShows(t *testing.T) {
	c, s := makeClientAndServer(nil)
	defer s.Close()

	shows, err := c.GetShows(context.Background(), "2")
	require.NoError(t, err)
	assert.Equal(t, []plex.Show{{Guid: "2", Title: "bar"}}, shows)
}

func TestClient_GetSeasons(t *testing.T) {
	c, s := makeClientAndServer(nil)
	defer s.Close()

	shows, err := c.GetSeasons(context.Background(), "200")
	require.NoError(t, err)
	assert.Equal(t, []plex.Season{{Guid: "2", Title: "Season 1"}}, shows)
}

func TestClient_GetEpisodes(t *testing.T) {
	c, s := makeClientAndServer(nil)
	defer s.Close()

	shows, err := c.GetEpisodes(context.Background(), "201")
	require.NoError(t, err)
	assert.Equal(t, []plex.Episode{{Guid: "2", Title: "Episode 1"}}, shows)
}
