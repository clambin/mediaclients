package plex

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/clambin/go-common/set"
)

// GetSessions retrieves session information from the server.
func (c *PMSClient) GetSessions(ctx context.Context) ([]Session, error) {
	type response struct {
		Metadata []Session `json:"Metadata"`
		Size     int       `json:"size"`
	}
	resp, err := call[response](ctx, c, "/status/sessions")
	return resp.Metadata, err
}

// Session contains one record in a Sessions
type Session struct {
	LastViewedAt          Timestamp      `json:"lastViewedAt"`
	UpdatedAt             Timestamp      `json:"updatedAt"`
	User                  SessionUser    `json:"User"`
	Session               SessionStats   `json:"Session"`
	Art                   string         `json:"art"`
	AudienceRatingImage   string         `json:"audienceRatingImage"`
	ContentRating         string         `json:"contentRating"`
	GrandparentArt        string         `json:"grandparentArt"`
	GrandparentGUID       string         `json:"grandparentGuid"`
	GrandparentKey        string         `json:"grandparentKey"`
	GrandparentRatingKey  string         `json:"grandparentRatingKey"`
	GrandparentTheme      string         `json:"grandparentTheme"`
	GrandparentThumb      string         `json:"grandparentThumb"`
	GrandparentTitle      string         `json:"grandparentTitle"`
	GUID                  string         `json:"guid"`
	Key                   string         `json:"key"`
	LibrarySectionID      string         `json:"librarySectionID"`
	LibrarySectionKey     string         `json:"librarySectionKey"`
	LibrarySectionTitle   string         `json:"librarySectionTitle"`
	OriginallyAvailableAt string         `json:"originallyAvailableAt"`
	ParentGUID            string         `json:"parentGuid"`
	ParentKey             string         `json:"parentKey"`
	ParentRatingKey       string         `json:"parentRatingKey"`
	ParentThumb           string         `json:"parentThumb"`
	ParentTitle           string         `json:"parentTitle"`
	RatingKey             string         `json:"ratingKey"`
	SessionKey            string         `json:"sessionKey"`
	Summary               string         `json:"summary"`
	Thumb                 string         `json:"thumb"`
	Title                 string         `json:"title"`
	Type                  string         `json:"type"`
	Media                 []SessionMedia `json:"Media"`
	Director              []struct {
		Filter string `json:"filter"`
		ID     string `json:"id"`
		Tag    string `json:"tag"`
	} `json:"Director"`
	Writer []struct {
		Filter string `json:"filter"`
		ID     string `json:"id"`
		Tag    string `json:"tag"`
	} `json:"Writer"`
	Rating2 []struct {
		Image string `json:"image"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"Rating"`
	Role []struct {
		Filter string `json:"filter"`
		ID     string `json:"id"`
		Role   string `json:"role"`
		Tag    string `json:"tag"`
		Thumb  string `json:"thumb,omitempty"`
	} `json:"Role"`
	Player           SessionPlayer     `json:"Player"`
	TranscodeSession SessionTranscoder `json:"TranscodeSession"`
	AddedAt          int               `json:"addedAt"`
	AudienceRating   float64           `json:"audienceRating"`
	Duration         int               `json:"duration"`
	Index            int               `json:"index"`
	ParentIndex      int               `json:"parentIndex"`
	Rating           float64           `json:"rating"`
	ViewOffset       int               `json:"viewOffset"`
}

// SessionMedia contains one record in a Session's Media list
type SessionMedia struct {
	AudioProfile          string             `json:"audioProfile"`
	ID                    string             `json:"id"`
	VideoProfile          string             `json:"videoProfile"`
	AudioCodec            string             `json:"audioCodec"`
	Container             string             `json:"container"`
	Protocol              string             `json:"protocol"`
	VideoCodec            string             `json:"videoCodec"`
	VideoFrameRate        string             `json:"videoFrameRate"`
	VideoResolution       string             `json:"videoResolution"`
	Part                  []MediaSessionPart `json:"Part"`
	AudioChannels         int                `json:"audioChannels"`
	Bitrate               int                `json:"bitrate"`
	Duration              int                `json:"duration"`
	Height                int                `json:"height"`
	Width                 int                `json:"width"`
	OptimizedForStreaming bool               `json:"optimizedForStreaming"`
	Selected              bool               `json:"selected"`
}

// MediaSessionPart contains one record in a MediaSession's Part list
type MediaSessionPart struct {
	AudioProfile          string                   `json:"audioProfile"`
	ID                    string                   `json:"id"`
	VideoProfile          string                   `json:"videoProfile"`
	Container             string                   `json:"container"`
	Protocol              string                   `json:"protocol"`
	Decision              string                   `json:"decision"`
	Stream                []MediaSessionPartStream `json:"Stream"`
	Bitrate               int                      `json:"bitrate"`
	Duration              int                      `json:"duration"`
	Height                int                      `json:"height"`
	Width                 int                      `json:"width"`
	OptimizedForStreaming bool                     `json:"optimizedForStreaming"`
	Selected              bool                     `json:"selected"`
}

// MediaSessionPartStream contains one stream (video, audio, subtitles) in a MediaSession's Part list
type MediaSessionPartStream struct {
	Codec                string  `json:"codec"`
	DisplayTitle         string  `json:"displayTitle"`
	ExtendedDisplayTitle string  `json:"extendedDisplayTitle"`
	ID                   string  `json:"id"`
	Language             string  `json:"language"`
	LanguageCode         string  `json:"languageCode"`
	LanguageTag          string  `json:"languageTag"`
	Decision             string  `json:"decision"`
	Location             string  `json:"location"`
	AudioChannelLayout   string  `json:"audioChannelLayout,omitempty"`
	BitrateMode          string  `json:"bitrateMode,omitempty"`
	Profile              string  `json:"profile,omitempty"`
	Title                string  `json:"title,omitempty"`
	Container            string  `json:"container,omitempty"`
	Format               string  `json:"format,omitempty"`
	Bitrate              int     `json:"bitrate,omitempty"`
	FrameRate            float64 `json:"frameRate,omitempty"`
	Height               int     `json:"height,omitempty"`
	StreamType           int     `json:"streamType"`
	Width                int     `json:"width,omitempty"`
	Channels             int     `json:"channels,omitempty"`
	SamplingRate         int     `json:"samplingRate,omitempty"`
	Default              bool    `json:"default"`
	Selected             bool    `json:"selected,omitempty"`
}

// SessionUser contains the user details inside a Session
type SessionUser struct {
	ID    string `json:"id"`
	Thumb string `json:"thumb"`
	Title string `json:"title"`
}

// SessionPlayer contains the player details inside a Session
type SessionPlayer struct {
	Address             string `json:"address"`
	Device              string `json:"device"`
	MachineIdentifier   string `json:"machineIdentifier"`
	Model               string `json:"model"`
	Platform            string `json:"platform"`
	PlatformVersion     string `json:"platformVersion"`
	Product             string `json:"product"`
	Profile             string `json:"profile"`
	RemotePublicAddress string `json:"remotePublicAddress"`
	State               string `json:"state"`
	Title               string `json:"title"`
	Version             string `json:"version"`
	Local               bool   `json:"local"`
	Relayed             bool   `json:"relayed"`
	Secure              bool   `json:"secure"`
	UserID              int    `json:"userID"`
}

// SessionStats contains the session details inside a Session
type SessionStats struct {
	ID        string `json:"id"`
	Location  string `json:"location"`
	Bandwidth int    `json:"bandwidth"`
}

// SessionTranscoder contains the transcoder details inside a Session.
// If the session doesn't transcode any media streams, all fields will be blank.
type SessionTranscoder struct {
	Key                     string  `json:"key"`
	Context                 string  `json:"context"`
	SourceVideoCodec        string  `json:"sourceVideoCodec"`
	SourceAudioCodec        string  `json:"sourceAudioCodec"`
	VideoDecision           string  `json:"videoDecision"`
	AudioDecision           string  `json:"audioDecision"`
	SubtitleDecision        string  `json:"subtitleDecision"`
	Protocol                string  `json:"protocol"`
	Container               string  `json:"container"`
	VideoCodec              string  `json:"videoCodec"`
	AudioCodec              string  `json:"audioCodec"`
	Progress                float64 `json:"progress"`
	Size                    int     `json:"size"`
	Speed                   float64 `json:"speed"`
	Duration                int     `json:"duration"`
	AudioChannels           int     `json:"audioChannels"`
	TimeStamp               float64 `json:"timeStamp"`
	Throttled               bool    `json:"throttled"`
	Complete                bool    `json:"complete"`
	Error                   bool    `json:"error"`
	TranscodeHwRequested    bool    `json:"transcodeHwRequested"`
	TranscodeHwFullPipeline bool    `json:"transcodeHwFullPipeline"`
}

// GetTitle returns the title of the movie, tv episode being played.  For movies, this is just the title.
// For TV Shows, it returns the show, season & episode title.
func (s Session) GetTitle() string {
	if s.Type == "episode" {
		return fmt.Sprintf("%s - S%02dE%02d - %s", s.GrandparentTitle, s.ParentIndex, s.Index, s.Title)
	}
	return s.Title

}

// GetProgress returns the progress of the session, i.e. how much of the movie / tv episode has been watched.
// Returns a percentage between 0.0 and 1.0
func (s Session) GetProgress() float64 {
	return float64(s.ViewOffset) / float64(s.Duration)
}

// GetVideoMode returns the session's video mode (transcoding, direct play, etc).
func (s Session) GetVideoMode() string {
	decisions := set.New[string]()
	for _, media := range s.Media {
		for _, part := range media.Part {
			videoDecision := part.Decision
			if videoDecision == "transcode" {
				videoDecision = s.TranscodeSession.VideoDecision
			}
			if videoDecision == "" {
				videoDecision = "unknown"
			}
			decisions.Add(videoDecision)
		}
	}
	modes := decisions.ListOrdered()
	if len(modes) == 0 {
		return "unknown"
	}
	return strings.Join(modes, ",")
}
