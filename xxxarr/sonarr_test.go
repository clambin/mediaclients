package xxxarr

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

var sonarrResponses = Responses{
	`/api/v3/system/status`: SonarrSystemStatus{Version: "1.2.3.4444"},
	`/api/v3/series/11`:     SonarrSeries{Title: "Foo"},
	`/api/v3/queue`:         sonarrQueueResponse{Page: 1, PageSize: 1, TotalRecords: 2, Records: []SonarrQueue{{Title: "foo"}}},
	`/api/v3/queue?page=2`:  sonarrQueueResponse{Page: 2, PageSize: 1, TotalRecords: 2, Records: []SonarrQueue{{Title: "bar"}}},
	`/api/v3/series`:        []SonarrSeries{{Title: "Foo", Monitored: true}, {Title: "Bar", Monitored: false}},
	`/api/v3/episode/11`:    SonarrEpisode{Title: "Foo", SeasonNumber: 1, EpisodeNumber: 2, Series: SonarrEpisodeSeries{Title: "Bar"}},
	`/api/v3/calendar`: []SonarrCalendar{
		{SeriesID: 11, Title: "Foo", SeasonNumber: 2, EpisodeNumber: 9, HasFile: false, Monitored: true},
		{SeriesID: 12, Title: "Bar", SeasonNumber: 1, EpisodeNumber: 1, HasFile: true, Monitored: true},
		{SeriesID: 13, Title: "Snafu", SeasonNumber: 1, EpisodeNumber: 1, HasFile: false, Monitored: false},
	},
}

func TestNewSonarrClient_GetURL(t *testing.T) {
	c := NewSonarrClient("foo", "1234", nil)
	assert.Equal(t, "foo", c.GetURL())
}

func TestSonarrClient_GetSystemStatus(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	response, err := c.GetSystemStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "1.2.3.4444", response.Version)
}

func TestSonarrClient_GetCalendar(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	calendar, err := c.GetCalendar(context.Background())
	require.NoError(t, err)
	require.Len(t, calendar, 3)
	assert.Equal(t, "Foo", calendar[0].Title)
	assert.True(t, calendar[1].HasFile)
	assert.False(t, calendar[2].Monitored)
}

func TestSonarrClient_GetQueue(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	queue, err := c.GetQueue(context.Background())
	require.NoError(t, err)
	require.Len(t, queue, 2)
	assert.Equal(t, "foo", queue[0].Title)
}

func TestSonarrClient_GetQueuePage(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	queue, _, err := c.GetQueuePage(context.Background(), 2)
	require.NoError(t, err)
	require.Len(t, queue, 1)
	assert.Equal(t, "bar", queue[0].Title)
}

func TestSonarrClient_GetSeries(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	series, err := c.GetSeries(context.Background())
	require.NoError(t, err)
	require.Len(t, series, 2)
	assert.Equal(t, "Foo", series[0].Title)
	assert.True(t, series[0].Monitored)
	assert.False(t, series[1].Monitored)
}

func TestSonarrClient_GetSeriesByID(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	series, err := c.GetSeriesByID(context.Background(), 11)
	require.NoError(t, err)
	assert.Equal(t, "Foo", series.Title)
}

func TestSonarrClient_GetEpisodeByID(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "1234", nil)
	episode, err := c.GetEpisodeByID(context.Background(), 11)
	require.NoError(t, err)
	assert.Equal(t, "Foo", episode.Title)
	assert.Equal(t, "Bar", episode.Series.Title)
	assert.Equal(t, 1, episode.SeasonNumber)
	assert.Equal(t, 2, episode.EpisodeNumber)
}

func TestSonarrClient_BadKey(t *testing.T) {
	s := NewTestServer(sonarrResponses, "1234")
	defer s.server.Close()

	c := NewSonarrClient(s.server.URL, "4321", nil)
	_, err := c.GetHealth(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403 Forbidden")
}

func TestSonarrClient_BadOutput(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("bad output"))
	}))
	defer s.Close()

	c := NewSonarrClient(s.URL, "1234", nil)
	_, err := c.GetHealth(context.Background())
	assert.Error(t, err)
	var err2 *ErrInvalidJSON
	assert.ErrorAs(t, err, &err2)
	assert.Equal(t, "bad output", string(err2.Body))
}
