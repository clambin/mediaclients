package xxxarr

import (
	"context"
	"net/http"
	"strconv"
)

// RadarrClient calls Radarr endpoints
type RadarrClient struct {
	Client *http.Client
	URL    string
}

func NewRadarrClient(url, apiKey string, roundtripper http.RoundTripper) *RadarrClient {
	if roundtripper == nil {
		roundtripper = http.DefaultTransport
	}
	return &RadarrClient{
		URL: url,
		Client: &http.Client{
			Transport: &authentication{
				apiKey: apiKey,
				next:   roundtripper,
			},
		},
	}
}

const radarrAPIPrefix = "/api/v3"

func (c RadarrClient) GetURL() string {
	return c.URL
}

// GetSystemStatus calls Radarr's  /api/v3/system/status endpoint. It returns the system status of the Radarr instance
func (c RadarrClient) GetSystemStatus(ctx context.Context) (RadarrSystemStatus, error) {
	return call[RadarrSystemStatus](ctx, c.Client, c.URL+radarrAPIPrefix+"/system/status")
}

// GetHealth calls Radarr's /api/v3/health endpoint. It returns the health of the Radarr instance
func (c RadarrClient) GetHealth(ctx context.Context) ([]RadarrHealth, error) {
	return call[[]RadarrHealth](ctx, c.Client, c.URL+radarrAPIPrefix+"/health")
}

// GetCalendar calls Radarr's /api/v3/calendar endpoint. It returns all movies that will become available in the next 24 hours
func (c RadarrClient) GetCalendar(ctx context.Context) ([]RadarrCalendar, error) {
	return call[[]RadarrCalendar](ctx, c.Client, c.URL+radarrAPIPrefix+"/calendar")
}

type radarrQueueResponse struct {
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	SortKey       string        `json:"sortKey"`
	SortDirection string        `json:"sortDirection"`
	TotalRecords  int           `json:"totalRecords"`
	Records       []RadarrQueue `json:"records"`
}

// GetQueuePage calls Radarr's /api/v3/queue/page=:pageNr endpoint. It returns one page of movies currently queued for download
func (c RadarrClient) GetQueuePage(ctx context.Context, pageNr int) ([]RadarrQueue, int, error) {
	response, err := call[radarrQueueResponse](ctx, c.Client, c.URL+radarrAPIPrefix+"/queue?page="+strconv.Itoa(pageNr))
	return response.Records, response.Page, err
}

// GetQueue calls Radarr's /api/v3/queue endpoint. It returns all movies currently queued for download
func (c RadarrClient) GetQueue(ctx context.Context) ([]RadarrQueue, error) {
	response, err := call[radarrQueueResponse](ctx, c.Client, c.URL+radarrAPIPrefix+"/queue")
	records := response.Records
	totalRecords := response.TotalRecords
	page := response.Page

	for err == nil && len(records) < totalRecords {
		var tmp []RadarrQueue
		if tmp, page, err = c.GetQueuePage(ctx, page+1); err == nil {
			records = append(records, tmp...)
		}
	}

	return records, err
}

// GetMovies calls Radarr's /api/v3/movie endpoint. It returns all movies added to Radarr
func (c RadarrClient) GetMovies(ctx context.Context) ([]RadarrMovie, error) {
	return call[[]RadarrMovie](ctx, c.Client, c.URL+radarrAPIPrefix+"/movie")
}

// GetMovieByID calls Radar's "/api/v3/movie/:movieID endpoint. It returns details for the specified movieID
func (c RadarrClient) GetMovieByID(ctx context.Context, movieID int) (RadarrMovie, error) {
	return call[RadarrMovie](ctx, c.Client, c.URL+radarrAPIPrefix+"/movie/"+strconv.Itoa(movieID))
}
