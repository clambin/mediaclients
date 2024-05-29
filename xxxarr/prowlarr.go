package xxxarr

import (
	"context"
	"net/http"
)

type ProwlarrClient struct {
	URL        string
	HTTPClient *http.Client
}

func NewProwlarrClient(url, apikey string, roundtripper http.RoundTripper) *ProwlarrClient {
	if roundtripper == nil {
		roundtripper = http.DefaultTransport
	}
	return &ProwlarrClient{
		URL: url,
		HTTPClient: &http.Client{Transport: &authentication{
			apiKey: apikey,
			next:   roundtripper,
		}},
	}
}

func (c ProwlarrClient) GetIndexStats(ctx context.Context) (ProwlarrIndexersStats, error) {
	return call[ProwlarrIndexersStats](ctx, c.HTTPClient, c.URL+"/api/v1/indexerstats")
}
