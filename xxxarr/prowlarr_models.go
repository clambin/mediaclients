package xxxarr

import (
	"strconv"
	"time"
)

// ProwlarrIndexersStats is the response of the /api/v1/indexerstatus API
type ProwlarrIndexersStats struct {
	Id         int                      `json:"id"`
	Indexers   []ProwlarrIndexerStats   `json:"indexers"`
	UserAgents []ProwlarrUserAgentStats `json:"userAgents"`
	Hosts      []ProwlarrHostStats      `json:"hosts"`
}

type ProwlarrIndexerStats struct {
	IndexerId                 int                  `json:"indexerId"`
	IndexerName               string               `json:"indexerName"`
	AverageResponseTime       ProwlarrResponseTime `json:"averageResponseTime"`
	NumberOfQueries           int                  `json:"numberOfQueries"`
	NumberOfGrabs             int                  `json:"numberOfGrabs"`
	NumberOfRssQueries        int                  `json:"numberOfRssQueries"`
	NumberOfAuthQueries       int                  `json:"numberOfAuthQueries"`
	NumberOfFailedQueries     int                  `json:"numberOfFailedQueries"`
	NumberOfFailedGrabs       int                  `json:"numberOfFailedGrabs"`
	NumberOfFailedRssQueries  int                  `json:"numberOfFailedRssQueries"`
	NumberOfFailedAuthQueries int                  `json:"numberOfFailedAuthQueries"`
}

type ProwlarrResponseTime time.Duration

func (p *ProwlarrResponseTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Duration(*p).Milliseconds(), 10)), nil
}

func (p *ProwlarrResponseTime) UnmarshalJSON(bytes []byte) error {
	val, err := strconv.Atoi(string(bytes))
	if err == nil {
		*p = ProwlarrResponseTime(time.Duration(val) * time.Millisecond)
	}
	return err
}

type ProwlarrUserAgentStats struct {
	UserAgent       string `json:"userAgent"`
	NumberOfQueries int    `json:"numberOfQueries"`
	NumberOfGrabs   int    `json:"numberOfGrabs"`
}

type ProwlarrHostStats struct {
	Host            string `json:"host"`
	NumberOfQueries int    `json:"numberOfQueries"`
	NumberOfGrabs   int    `json:"numberOfGrabs"`
}
