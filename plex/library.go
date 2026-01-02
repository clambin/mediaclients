package plex

import (
	"context"
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

type Movie struct {
	LastViewedAt          Timestamp `json:"lastViewedAt,omitempty"`
	AddedAt               Timestamp `json:"addedAt"`
	UpdatedAt             Timestamp `json:"updatedAt"`
	RatingKey             string    `json:"ratingKey"`
	Key                   string    `json:"key"`
	Guid                  string    `json:"guid"`
	Studio                string    `json:"studio,omitempty"`
	Type                  string    `json:"type"`
	Title                 string    `json:"title"`
	ContentRating         string    `json:"contentRating,omitempty"`
	Summary               string    `json:"summary"`
	Tagline               string    `json:"tagline,omitempty"`
	Thumb                 string    `json:"thumb,omitempty"`
	Art                   string    `json:"art,omitempty"`
	OriginallyAvailableAt string    `json:"originallyAvailableAt,omitempty"`
	AudienceRatingImage   string    `json:"audienceRatingImage,omitempty"`
	PrimaryExtraKey       string    `json:"primaryExtraKey,omitempty"`
	RatingImage           string    `json:"ratingImage,omitempty"`
	ChapterSource         string    `json:"chapterSource,omitempty"`
	TitleSort             string    `json:"titleSort,omitempty"`
	Media                 []Media   `json:"Media"`
	Genre                 []struct {
		Tag string `json:"tag"`
	} `json:"Genre,omitempty"`
	Country []struct {
		Tag string `json:"tag"`
	} `json:"Country,omitempty"`
	Director []struct {
		Tag string `json:"tag"`
	} `json:"Director,omitempty"`
	Writer []struct {
		Tag string `json:"tag"`
	} `json:"Writer,omitempty"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role,omitempty"`
	Rating         float64 `json:"rating,omitempty"`
	AudienceRating float64 `json:"audienceRating,omitempty"`
	ViewCount      int     `json:"viewCount,omitempty"`
	Year           int     `json:"year,omitempty"`
	Duration       int     `json:"duration"`
	SkipCount      int     `json:"skipCount,omitempty"`
	UserRating     float64 `json:"userRating,omitempty"`
	LastRatedAt    int     `json:"lastRatedAt,omitempty"`
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

type Show struct {
	LastViewedAt          Timestamp `json:"lastViewedAt,omitempty"`
	AddedAt               Timestamp `json:"addedAt"`
	UpdatedAt             Timestamp `json:"updatedAt"`
	RatingKey             string    `json:"ratingKey"`
	Key                   string    `json:"key"`
	Guid                  string    `json:"guid"`
	Studio                string    `json:"studio"`
	Type                  string    `json:"type"`
	Title                 string    `json:"title"`
	ContentRating         string    `json:"contentRating"`
	Summary               string    `json:"summary"`
	Thumb                 string    `json:"thumb"`
	Art                   string    `json:"art"`
	Theme                 string    `json:"theme,omitempty"`
	OriginallyAvailableAt string    `json:"originallyAvailableAt"`
	AudienceRatingImage   string    `json:"audienceRatingImage"`
	PrimaryExtraKey       string    `json:"primaryExtraKey,omitempty"`
	Tagline               string    `json:"tagline,omitempty"`
	TitleSort             string    `json:"titleSort,omitempty"`
	Genre                 []struct {
		Tag string `json:"tag"`
	} `json:"Genre"`
	Country []struct {
		Tag string `json:"tag"`
	} `json:"Country"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
	Index           int     `json:"index"`
	AudienceRating  float64 `json:"audienceRating"`
	ViewCount       int     `json:"viewCount,omitempty"`
	Year            int     `json:"year"`
	Duration        int     `json:"duration"`
	LeafCount       int     `json:"leafCount"`
	ViewedLeafCount int     `json:"viewedLeafCount"`
	ChildCount      int     `json:"childCount"`
	SkipCount       int     `json:"skipCount,omitempty"`
}

type Season struct {
	LastViewedAt          Timestamp `json:"lastViewedAt"`
	AddedAt               Timestamp `json:"addedAt"`
	UpdatedAt             Timestamp `json:"updatedAt"`
	RatingKey             string    `json:"ratingKey"`
	Key                   string    `json:"key"`
	ParentRatingKey       string    `json:"parentRatingKey"`
	GrandparentRatingKey  string    `json:"grandparentRatingKey"`
	Guid                  string    `json:"guid"`
	ParentGuid            string    `json:"parentGuid"`
	GrandparentGuid       string    `json:"grandparentGuid"`
	Type                  string    `json:"type"`
	Title                 string    `json:"title"`
	GrandparentKey        string    `json:"grandparentKey"`
	ParentKey             string    `json:"parentKey"`
	GrandparentTitle      string    `json:"grandparentTitle"`
	ParentTitle           string    `json:"parentTitle"`
	ContentRating         string    `json:"contentRating"`
	Summary               string    `json:"summary"`
	Thumb                 string    `json:"thumb"`
	Art                   string    `json:"art"`
	ParentThumb           string    `json:"parentThumb"`
	GrandparentThumb      string    `json:"grandparentThumb"`
	GrandparentArt        string    `json:"grandparentArt"`
	GrandparentTheme      string    `json:"grandparentTheme"`
	OriginallyAvailableAt string    `json:"originallyAvailableAt"`
	AudienceRatingImage   string    `json:"audienceRatingImage"`
	Media                 []Media   `json:"Media"`
	Director              []struct {
		Tag string `json:"tag"`
	} `json:"Director"`
	Writer []struct {
		Tag string `json:"tag"`
	} `json:"Writer"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
	Index          int     `json:"index"`
	ParentIndex    int     `json:"parentIndex"`
	AudienceRating float64 `json:"audienceRating"`
	ViewCount      int     `json:"viewCount"`
	Year           int     `json:"year"`
	Duration       int     `json:"duration"`
}

type Episode struct {
	LastViewedAt          Timestamp `json:"lastViewedAt"`
	AddedAt               Timestamp `json:"addedAt"`
	UpdatedAt             Timestamp `json:"updatedAt"`
	RatingKey             string    `json:"ratingKey"`
	Key                   string    `json:"key"`
	ParentRatingKey       string    `json:"parentRatingKey"`
	GrandparentRatingKey  string    `json:"grandparentRatingKey"`
	Guid                  string    `json:"guid"`
	ParentGuid            string    `json:"parentGuid"`
	GrandparentGuid       string    `json:"grandparentGuid"`
	Type                  string    `json:"type"`
	Title                 string    `json:"title"`
	GrandparentKey        string    `json:"grandparentKey"`
	ParentKey             string    `json:"parentKey"`
	GrandparentTitle      string    `json:"grandparentTitle"`
	ParentTitle           string    `json:"parentTitle"`
	ContentRating         string    `json:"contentRating"`
	Summary               string    `json:"summary"`
	Thumb                 string    `json:"thumb"`
	Art                   string    `json:"art"`
	ParentThumb           string    `json:"parentThumb"`
	GrandparentThumb      string    `json:"grandparentThumb"`
	GrandparentArt        string    `json:"grandparentArt"`
	GrandparentTheme      string    `json:"grandparentTheme"`
	OriginallyAvailableAt string    `json:"originallyAvailableAt"`
	AudienceRatingImage   string    `json:"audienceRatingImage"`
	Media                 []Media   `json:"Media"`
	Director              []struct {
		Tag string `json:"tag"`
	} `json:"Director"`
	Writer []struct {
		Tag string `json:"tag"`
	} `json:"Writer"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
	Index          int     `json:"index"`
	ParentIndex    int     `json:"parentIndex"`
	AudienceRating float64 `json:"audienceRating"`
	ViewCount      int     `json:"viewCount"`
	Year           int     `json:"year"`
	Duration       int     `json:"duration"`
}

func (c *PMSClient) GetLibraries(ctx context.Context) ([]Library, error) {
	type response struct {
		Directory []Library `json:"Directory"`
	}
	resp, err := call[response](ctx, c, "/library/sections")
	return resp.Directory, err
}

func (c *PMSClient) GetMovies(ctx context.Context, key string) ([]Movie, error) {
	type response struct {
		Metadata []Movie `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/sections/"+key+"/all")
	return resp.Metadata, err
}

func (c *PMSClient) GetShows(ctx context.Context, key string) ([]Show, error) {
	type response struct {
		Metadata []Show `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/sections/"+key+"/all")
	return resp.Metadata, err
}

func (c *PMSClient) GetSeasons(ctx context.Context, key string) ([]Season, error) {
	type response struct {
		Metadata []Season `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/metadata/"+key+"/children")
	return resp.Metadata, err
}

func (c *PMSClient) GetEpisodes(ctx context.Context, key string) ([]Episode, error) {
	type response struct {
		Metadata []Episode `json:"Metadata"`
	}
	resp, err := call[response](ctx, c, "/library/metadata/"+key+"/children")
	return resp.Metadata, err
}
