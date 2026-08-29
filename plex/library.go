package plex

import (
	"context"
	"errors"
	"iter"
)

type Library struct {
	UpdatedAt        Timestamp `json:"updatedAt"`
	CreatedAt        Timestamp `json:"createdAt"`
	ScannedAt        Timestamp `json:"scannedAt"`
	ContentChangedAt Timestamp `json:"contentChangedAt"`
	Art              string    `json:"art"`
	Composite        string    `json:"composite"`
	Thumb            string    `json:"thumb"`
	Key              string    `json:"key"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	Agent            string    `json:"agent"`
	Scanner          string    `json:"scanner"`
	Language         string    `json:"language"`
	Uuid             string    `json:"uuid"`
	Location         []struct {
		Path string `json:"path"`
		Id   int    `json:"id"`
	} `json:"Location"`
	Hidden     int  `json:"hidden"`
	AllowSync  bool `json:"allowSync"`
	Filters    bool `json:"filters"`
	Refreshing bool `json:"refreshing"`
	Content    bool `json:"content"`
	Directory  bool `json:"directory"`
}

type MediaMetadata struct {
	UltraBlurColors struct {
		TopLeft     string `json:"topLeft"`
		TopRight    string `json:"topRight"`
		BottomRight string `json:"bottomRight"`
		BottomLeft  string `json:"bottomLeft"`
	} `json:"UltraBlurColors"`
	GrandparentRatingKey  string `json:"grandparentRatingKey"`
	ParentThumb           string `json:"parentThumb"`
	AudienceRatingImage   string `json:"audienceRatingImage"`
	ParentRatingKey       string `json:"parentRatingKey"`
	GrandparentKey        string `json:"grandparentKey"`
	ParentKey             string `json:"parentKey"`
	OriginallyAvailableAt string `json:"originallyAvailableAt"`
	RatingKey             string `json:"ratingKey"`
	ParentTitle           string `json:"parentTitle"`
	Guid                  string `json:"guid"`
	Type                  string `json:"type"`
	Title                 string `json:"title"`
	GrandparentTitle      string `json:"grandparentTitle"`
	GrandparentArt        string `json:"grandparentArt"`
	Key                   string `json:"key"`
	GrandparentGuid       string `json:"grandparentGuid"`
	GrandparentSlug       string `json:"grandparentSlug"`
	Summary               string `json:"summary"`
	ParentGuid            string `json:"parentGuid"`
	Thumb                 string `json:"thumb"`
	Art                   string `json:"art"`
	ContentRating         string `json:"contentRating"`
	GrandparentThumb      string `json:"grandparentThumb"`
	GrandparentTheme      string `json:"grandparentTheme"`
	Writer                []struct {
		Tag string `json:"tag"`
	} `json:"Writer"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
	Image []struct {
		Alt  string `json:"alt"`
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"Image"`
	Media    []Media `json:"Media"`
	Director []struct {
		Tag string `json:"tag"`
	} `json:"Director"`
	AddedAt        int     `json:"addedAt"`
	AudienceRating float64 `json:"audienceRating"`
	LastViewedAt   int     `json:"lastViewedAt"`
	UpdatedAt      int     `json:"updatedAt"`
	Year           int     `json:"year"`
	Duration       int     `json:"duration"`
	Index          int     `json:"index"`
	ParentIndex    int     `json:"parentIndex"`
	ViewCount      int     `json:"viewCount"`
}

type Media struct {
	AudioCodec            string      `json:"audioCodec"`
	VideoCodec            string      `json:"videoCodec"`
	VideoResolution       string      `json:"videoResolution"`
	Container             string      `json:"container"`
	VideoFrameRate        string      `json:"videoFrameRate"`
	AudioProfile          string      `json:"audioProfile,omitempty"`
	VideoProfile          string      `json:"videoProfile"`
	Part                  []MediaPart `json:"Part"`
	Id                    int         `json:"id"`
	Duration              int         `json:"duration"`
	Bitrate               int         `json:"bitrate"`
	Width                 int         `json:"width"`
	Height                int         `json:"height"`
	AspectRatio           float64     `json:"aspectRatio"`
	AudioChannels         int         `json:"audioChannels"`
	OptimizedForStreaming int         `json:"optimizedForStreaming,omitempty"`
	Has64BitOffsets       bool        `json:"has64bitOffsets,omitempty"`
	HasVoiceActivity      bool        `json:"hasVoiceActivity"`
}

type MediaPart struct {
	Key                   string `json:"key"`
	File                  string `json:"file"`
	AudioProfile          string `json:"audioProfile,omitempty"`
	Container             string `json:"container"`
	VideoProfile          string `json:"videoProfile"`
	HasThumbnail          string `json:"hasThumbnail,omitempty"`
	Id                    int    `json:"id"`
	Duration              int    `json:"duration"`
	Size                  int64  `json:"size"`
	Has64BitOffsets       bool   `json:"has64bitOffsets,omitempty"`
	OptimizedForStreaming bool   `json:"optimizedForStreaming,omitempty"`
}

func (c *Client) GetLibraries(ctx context.Context) ([]Library, error) {
	type response struct {
		MediaContainer struct {
			Directory []Library `json:"directory"`
		} `json:"MediaContainer"`
	}
	resp, err := call[response](ctx, c, "/library/sections")
	return resp.MediaContainer.Directory, err
}

func (c *Client) GetAllLibraryMedia(ctx context.Context, key string) iter.Seq2[MediaMetadata, error] {
	type response struct {
		MediaContainer struct {
			TotalSize int             `json:"totalSize"`
			Metadata  []MediaMetadata `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	return func(yield func(metadata MediaMetadata, err error) bool) {
		const pageSize = 500
		var currentPage int
		var currentRecords int
		var totalRecords int
		for {
			// get a new page
			resp, err := call[response](ctx, c, "/library/sections/"+key+"/allLeaves", withPagination(currentPage, pageSize))
			if err != nil {
				yield(MediaMetadata{}, err)
				return
			}

			// update total number of records
			totalRecords = resp.MediaContainer.TotalSize

			// yield all records, increasing currentRecords
			for _, metadata := range resp.MediaContainer.Metadata {
				if !yield(metadata, nil) {
					return
				}
			}

			// increase record counter.
			// if we got an empty page, we may fall into an infinite loop, so abort.
			pageRecords := len(resp.MediaContainer.Metadata)
			if pageRecords == 0 {
				yield(MediaMetadata{}, errors.New("api call returned empty page"))
				return
			}
			currentRecords += pageRecords

			// check for more data
			if currentRecords >= totalRecords {
				return
			}

			// set up next page
			currentPage++
		}
	}
}
