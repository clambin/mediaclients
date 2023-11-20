package xxxarr

import (
	"context"
	"net/http"
	"strconv"
)

// SonarrClient calls Sonarr endpoints
type SonarrClient struct {
	Client *http.Client
	URL    string
}

func NewSonarrClient(url, apiKey string, roundtripper http.RoundTripper) *SonarrClient {
	if roundtripper == nil {
		roundtripper = http.DefaultTransport
	}
	return &SonarrClient{
		URL: url,
		Client: &http.Client{
			Transport: &authentication{
				apiKey: apiKey,
				next:   roundtripper,
			},
		},
	}
}

const sonarrAPIPrefix = "/api/v3"

func (c SonarrClient) GetURL() string {
	return c.URL
}

// GetSystemStatus calls Sonarr's /api/v3/system/status endpoint. It returns the system status of the Sonarr instance
func (c SonarrClient) GetSystemStatus(ctx context.Context) (response SonarrSystemStatus, err error) {
	return call[SonarrSystemStatus](ctx, c.Client, c.URL+sonarrAPIPrefix+"/system/status")
}

// GetHealth calls Sonarr's /api/v3/health endpoint. It returns the health of the Radarr instance
func (c SonarrClient) GetHealth(ctx context.Context) (response []SonarrHealth, err error) {
	return call[[]SonarrHealth](ctx, c.Client, c.URL+sonarrAPIPrefix+"/health")
}

// GetCalendar calls Sonarr's /api/v3/calendar endpoint. It returns all episodes that will become available in the next 24 hours
func (c SonarrClient) GetCalendar(ctx context.Context) (response []SonarrCalendar, err error) {
	return call[[]SonarrCalendar](ctx, c.Client, c.URL+sonarrAPIPrefix+"/calendar")
}

type sonarrQueueResponse struct {
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	SortKey       string        `json:"sortKey"`
	SortDirection string        `json:"sortDirection"`
	TotalRecords  int           `json:"totalRecords"`
	Records       []SonarrQueue `json:"records"`
}

// GetQueuePage calls Sonarr's /api/v3/queue/page=:pageNr endpoint. It returns one page of episodes currently queued for download
func (c SonarrClient) GetQueuePage(ctx context.Context, pageNr int) ([]SonarrQueue, int, error) {
	response, err := call[sonarrQueueResponse](ctx, c.Client, c.URL+sonarrAPIPrefix+"/queue?page="+strconv.Itoa(pageNr))
	return response.Records, response.Page, err
}

// GetQueue calls Sonarr's /api/v3/queue endpoint. It returns all episodes currently queued for download
func (c SonarrClient) GetQueue(ctx context.Context) ([]SonarrQueue, error) {
	response, err := call[sonarrQueueResponse](ctx, c.Client, c.URL+sonarrAPIPrefix+"/queue")
	records := response.Records
	totalRecords := response.TotalRecords
	page := response.Page

	for err == nil && len(records) < totalRecords {
		var tmp []SonarrQueue
		if tmp, page, err = c.GetQueuePage(ctx, page+1); err == nil {
			records = append(records, tmp...)
		}
	}

	return records, err
}

// GetSeries calls Sonarr's /api/v3/series endpoint. It returns all series added to Sonarr
func (c SonarrClient) GetSeries(ctx context.Context) (response []SonarrSeries, err error) {
	return call[[]SonarrSeries](ctx, c.Client, c.URL+sonarrAPIPrefix+"/series")
}

// GetSeriesByID calls Sonarr's /api/v3/series/:seriesID endpoint. It returns details for the specified seriesID
func (c SonarrClient) GetSeriesByID(ctx context.Context, seriesID int) (response SonarrSeries, err error) {
	return call[SonarrSeries](ctx, c.Client, c.URL+sonarrAPIPrefix+"/series/"+strconv.Itoa(seriesID))
}

// GetEpisodeByID calls Sonarr's /api/v3/episode/:episodeID endpoint. It returns details for the specified responseID
func (c SonarrClient) GetEpisodeByID(ctx context.Context, episodeID int) (response SonarrEpisode, err error) {
	return call[SonarrEpisode](ctx, c.Client, c.URL+sonarrAPIPrefix+"/episode/"+strconv.Itoa(episodeID))
}
