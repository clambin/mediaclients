package plex

type Library struct {
	AllowSync        bool      `json:"allowSync"`
	Art              string    `json:"art"`
	Composite        string    `json:"composite"`
	Filters          bool      `json:"filters"`
	Refreshing       bool      `json:"refreshing"`
	Thumb            string    `json:"thumb"`
	Key              string    `json:"key"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	Agent            string    `json:"agent"`
	Scanner          string    `json:"scanner"`
	Language         string    `json:"language"`
	Uuid             string    `json:"uuid"`
	UpdatedAt        Timestamp `json:"updatedAt"`
	CreatedAt        Timestamp `json:"createdAt"`
	ScannedAt        Timestamp `json:"scannedAt"`
	Content          bool      `json:"content"`
	Directory        bool      `json:"directory"`
	ContentChangedAt int       `json:"contentChangedAt"`
	Hidden           int       `json:"hidden"`
	Location         []struct {
		Id   int    `json:"id"`
		Path string `json:"path"`
	} `json:"Location"`
}

type Movie struct {
	RatingKey             string  `json:"ratingKey"`
	Key                   string  `json:"key"`
	Guid                  string  `json:"guid"`
	Studio                string  `json:"studio,omitempty"`
	Type                  string  `json:"type"`
	Title                 string  `json:"title"`
	ContentRating         string  `json:"contentRating,omitempty"`
	Summary               string  `json:"summary"`
	Rating                float64 `json:"rating,omitempty"`
	AudienceRating        float64 `json:"audienceRating,omitempty"`
	ViewCount             int     `json:"viewCount,omitempty"`
	LastViewedAt          int     `json:"lastViewedAt,omitempty"`
	Year                  int     `json:"year,omitempty"`
	Tagline               string  `json:"tagline,omitempty"`
	Thumb                 string  `json:"thumb,omitempty"`
	Art                   string  `json:"art,omitempty"`
	Duration              int     `json:"duration"`
	OriginallyAvailableAt string  `json:"originallyAvailableAt,omitempty"`
	AddedAt               int     `json:"addedAt"`
	UpdatedAt             int     `json:"updatedAt"`
	AudienceRatingImage   string  `json:"audienceRatingImage,omitempty"`
	PrimaryExtraKey       string  `json:"primaryExtraKey,omitempty"`
	RatingImage           string  `json:"ratingImage,omitempty"`
	Media                 []Media `json:"Media"`
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
	ChapterSource string  `json:"chapterSource,omitempty"`
	TitleSort     string  `json:"titleSort,omitempty"`
	SkipCount     int     `json:"skipCount,omitempty"`
	UserRating    float64 `json:"userRating,omitempty"`
	LastRatedAt   int     `json:"lastRatedAt,omitempty"`
}

type Media struct {
	Id                    int         `json:"id"`
	Duration              int         `json:"duration"`
	Bitrate               int         `json:"bitrate"`
	Width                 int         `json:"width"`
	Height                int         `json:"height"`
	AspectRatio           float64     `json:"aspectRatio"`
	AudioChannels         int         `json:"audioChannels"`
	AudioCodec            string      `json:"audioCodec"`
	VideoCodec            string      `json:"videoCodec"`
	VideoResolution       string      `json:"videoResolution"`
	Container             string      `json:"container"`
	VideoFrameRate        string      `json:"videoFrameRate"`
	OptimizedForStreaming int         `json:"optimizedForStreaming,omitempty"`
	AudioProfile          string      `json:"audioProfile,omitempty"`
	Has64BitOffsets       bool        `json:"has64bitOffsets,omitempty"`
	VideoProfile          string      `json:"videoProfile"`
	Part                  []MediaPart `json:"Part"`
}

type MediaPart struct {
	Id                    int    `json:"id"`
	Key                   string `json:"key"`
	Duration              int    `json:"duration"`
	File                  string `json:"file"`
	Size                  int64  `json:"size"`
	AudioProfile          string `json:"audioProfile,omitempty"`
	Container             string `json:"container"`
	Has64BitOffsets       bool   `json:"has64bitOffsets,omitempty"`
	OptimizedForStreaming bool   `json:"optimizedForStreaming,omitempty"`
	VideoProfile          string `json:"videoProfile"`
	HasThumbnail          string `json:"hasThumbnail,omitempty"`
}

type Show struct {
	RatingKey             string  `json:"ratingKey"`
	Key                   string  `json:"key"`
	Guid                  string  `json:"guid"`
	Studio                string  `json:"studio"`
	Type                  string  `json:"type"`
	Title                 string  `json:"title"`
	ContentRating         string  `json:"contentRating"`
	Summary               string  `json:"summary"`
	Index                 int     `json:"index"`
	AudienceRating        float64 `json:"audienceRating"`
	ViewCount             int     `json:"viewCount,omitempty"`
	LastViewedAt          int     `json:"lastViewedAt,omitempty"`
	Year                  int     `json:"year"`
	Thumb                 string  `json:"thumb"`
	Art                   string  `json:"art"`
	Theme                 string  `json:"theme,omitempty"`
	Duration              int     `json:"duration"`
	OriginallyAvailableAt string  `json:"originallyAvailableAt"`
	LeafCount             int     `json:"leafCount"`
	ViewedLeafCount       int     `json:"viewedLeafCount"`
	ChildCount            int     `json:"childCount"`
	AddedAt               int     `json:"addedAt"`
	UpdatedAt             int     `json:"updatedAt"`
	AudienceRatingImage   string  `json:"audienceRatingImage"`
	PrimaryExtraKey       string  `json:"primaryExtraKey,omitempty"`
	Genre                 []struct {
		Tag string `json:"tag"`
	} `json:"Genre"`
	Country []struct {
		Tag string `json:"tag"`
	} `json:"Country"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
	SkipCount int    `json:"skipCount,omitempty"`
	Tagline   string `json:"tagline,omitempty"`
	TitleSort string `json:"titleSort,omitempty"`
}

type Season struct {
	RatingKey             string  `json:"ratingKey"`
	Key                   string  `json:"key"`
	ParentRatingKey       string  `json:"parentRatingKey"`
	GrandparentRatingKey  string  `json:"grandparentRatingKey"`
	Guid                  string  `json:"guid"`
	ParentGuid            string  `json:"parentGuid"`
	GrandparentGuid       string  `json:"grandparentGuid"`
	Type                  string  `json:"type"`
	Title                 string  `json:"title"`
	GrandparentKey        string  `json:"grandparentKey"`
	ParentKey             string  `json:"parentKey"`
	GrandparentTitle      string  `json:"grandparentTitle"`
	ParentTitle           string  `json:"parentTitle"`
	ContentRating         string  `json:"contentRating"`
	Summary               string  `json:"summary"`
	Index                 int     `json:"index"`
	ParentIndex           int     `json:"parentIndex"`
	AudienceRating        float64 `json:"audienceRating"`
	ViewCount             int     `json:"viewCount"`
	LastViewedAt          int     `json:"lastViewedAt"`
	Year                  int     `json:"year"`
	Thumb                 string  `json:"thumb"`
	Art                   string  `json:"art"`
	ParentThumb           string  `json:"parentThumb"`
	GrandparentThumb      string  `json:"grandparentThumb"`
	GrandparentArt        string  `json:"grandparentArt"`
	GrandparentTheme      string  `json:"grandparentTheme"`
	Duration              int     `json:"duration"`
	OriginallyAvailableAt string  `json:"originallyAvailableAt"`
	AddedAt               int     `json:"addedAt"`
	UpdatedAt             int     `json:"updatedAt"`
	AudienceRatingImage   string  `json:"audienceRatingImage"`
	Media                 []Media `json:"Media"`
	Director              []struct {
		Tag string `json:"tag"`
	} `json:"Director"`
	Writer []struct {
		Tag string `json:"tag"`
	} `json:"Writer"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
}

type Episode struct {
	RatingKey             string  `json:"ratingKey"`
	Key                   string  `json:"key"`
	ParentRatingKey       string  `json:"parentRatingKey"`
	GrandparentRatingKey  string  `json:"grandparentRatingKey"`
	Guid                  string  `json:"guid"`
	ParentGuid            string  `json:"parentGuid"`
	GrandparentGuid       string  `json:"grandparentGuid"`
	Type                  string  `json:"type"`
	Title                 string  `json:"title"`
	GrandparentKey        string  `json:"grandparentKey"`
	ParentKey             string  `json:"parentKey"`
	GrandparentTitle      string  `json:"grandparentTitle"`
	ParentTitle           string  `json:"parentTitle"`
	ContentRating         string  `json:"contentRating"`
	Summary               string  `json:"summary"`
	Index                 int     `json:"index"`
	ParentIndex           int     `json:"parentIndex"`
	AudienceRating        float64 `json:"audienceRating"`
	ViewCount             int     `json:"viewCount"`
	LastViewedAt          int     `json:"lastViewedAt"`
	Year                  int     `json:"year"`
	Thumb                 string  `json:"thumb"`
	Art                   string  `json:"art"`
	ParentThumb           string  `json:"parentThumb"`
	GrandparentThumb      string  `json:"grandparentThumb"`
	GrandparentArt        string  `json:"grandparentArt"`
	GrandparentTheme      string  `json:"grandparentTheme"`
	Duration              int     `json:"duration"`
	OriginallyAvailableAt string  `json:"originallyAvailableAt"`
	AddedAt               int     `json:"addedAt"`
	UpdatedAt             int     `json:"updatedAt"`
	AudienceRatingImage   string  `json:"audienceRatingImage"`
	Media                 []Media `json:"Media"`
	Director              []struct {
		Tag string `json:"tag"`
	} `json:"Director"`
	Writer []struct {
		Tag string `json:"tag"`
	} `json:"Writer"`
	Role []struct {
		Tag string `json:"tag"`
	} `json:"Role"`
}
