package xxxarr

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProwlarrClient_GetIndexStats(t *testing.T) {
	s := NewTestServer(prowlarrResponses, "1234")
	defer s.server.Close()

	c := NewProwlarrClient(s.server.URL, "1234", nil)
	ctx := context.Background()
	resp, err := c.GetIndexStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, prowlarrResponses["/api/v1/indexerstats"], resp)
	assert.Equal(t, 100*time.Millisecond, time.Duration(resp.Indexers[0].AverageResponseTime))
}

var prowlarrResponses = Responses{
	"/api/v1/indexerstats": ProwlarrIndexersStats{
		Id: 1,
		Indexers: []ProwlarrIndexerStats{{
			IndexerId:                 1,
			IndexerName:               "indexer",
			AverageResponseTime:       ProwlarrResponseTime(100 * time.Millisecond),
			NumberOfQueries:           20,
			NumberOfGrabs:             10,
			NumberOfRssQueries:        5,
			NumberOfAuthQueries:       1,
			NumberOfFailedQueries:     4,
			NumberOfFailedGrabs:       3,
			NumberOfFailedRssQueries:  2,
			NumberOfFailedAuthQueries: 1,
		}},
		UserAgents: []ProwlarrUserAgentStats{{
			UserAgent:       "user-agent",
			NumberOfQueries: 20,
			NumberOfGrabs:   10,
		}},
		Hosts: []ProwlarrHostStats{{
			Host:            "host1",
			NumberOfQueries: 20,
			NumberOfGrabs:   10,
		}},
	},
}
