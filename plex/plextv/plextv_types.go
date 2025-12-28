package plextv

import (
	"encoding/xml"
	"strconv"
	"time"
)

// PlexTimestamp is a custom type for parsing Plex timestamps. It's mainly used by legacy API endpoints.
type PlexTimestamp time.Time

func (t *PlexTimestamp) UnmarshalXMLAttr(attr xml.Attr) error {
	epoc, err := strconv.ParseInt(attr.Value, 10, 64)
	if err != nil {
		return err
	}
	*t = PlexTimestamp(time.Unix(epoc, 0))
	return nil
}

// User represents a Plex TV user. It is the response from the /api/v2/user endpoint.
type User struct {
	Profile struct {
		DefaultAudioLanguages        interface{} `json:"defaultAudioLanguages"`
		DefaultSubtitleLanguages     interface{} `json:"defaultSubtitleLanguages"`
		MediaReviewsLanguages        interface{} `json:"mediaReviewsLanguages"`
		DefaultAudioLanguage         string      `json:"defaultAudioLanguage"`
		DefaultSubtitleLanguage      string      `json:"defaultSubtitleLanguage"`
		DefaultAudioAccessibility    int         `json:"defaultAudioAccessibility"`
		AutoSelectSubtitle           int         `json:"autoSelectSubtitle"`
		DefaultSubtitleAccessibility int         `json:"defaultSubtitleAccessibility"`
		DefaultSubtitleForced        int         `json:"defaultSubtitleForced"`
		WatchedIndicator             int         `json:"watchedIndicator"`
		MediaReviewsVisibility       int         `json:"mediaReviewsVisibility"`
		AutoSelectAudio              bool        `json:"autoSelectAudio"`
	} `json:"profile"`
	Locale                  interface{} `json:"locale"`
	AttributionPartner      interface{} `json:"attributionPartner"`
	Uuid                    string      `json:"uuid"`
	Username                string      `json:"username"`
	Title                   string      `json:"title"`
	Email                   string      `json:"email"`
	FriendlyName            string      `json:"friendlyName"`
	Thumb                   string      `json:"thumb"`
	AuthToken               string      `json:"authToken"`
	MailingListStatus       string      `json:"mailingListStatus"`
	ScrobbleTypes           string      `json:"scrobbleTypes"`
	Country                 string      `json:"country"`
	SubscriptionDescription string      `json:"subscriptionDescription"`
	Subscription            struct {
		SubscribedAt   time.Time `json:"subscribedAt"`
		Status         string    `json:"status"`
		PaymentService string    `json:"paymentService"`
		Plan           string    `json:"plan"`
		Features       []string  `json:"features"`
		Active         bool      `json:"active"`
	} `json:"subscription"`
	Entitlements []string `json:"entitlements"`
	Roles        []string `json:"roles"`
	Services     []struct {
		Identifier string  `json:"identifier"`
		Endpoint   string  `json:"endpoint"`
		Token      *string `json:"token"`
		Secret     *string `json:"secret"`
		Status     string  `json:"status"`
	} `json:"services"`
	Id                   int  `json:"id"`
	JoinedAt             int  `json:"joinedAt"`
	HomeSize             int  `json:"homeSize"`
	MaxHomeSize          int  `json:"maxHomeSize"`
	RememberExpiresAt    int  `json:"rememberExpiresAt"`
	AdsConsentSetAt      int  `json:"adsConsentSetAt"`
	AdsConsentReminderAt int  `json:"adsConsentReminderAt"`
	Confirmed            bool `json:"confirmed"`
	EmailOnlyAuth        bool `json:"emailOnlyAuth"`
	HasPassword          bool `json:"hasPassword"`
	Protected            bool `json:"protected"`
	MailingListActive    bool `json:"mailingListActive"`
	Restricted           bool `json:"restricted"`
	Anonymous            bool `json:"anonymous"`
	Home                 bool `json:"home"`
	Guest                bool `json:"guest"`
	HomeAdmin            bool `json:"homeAdmin"`
	AdsConsent           bool `json:"adsConsent"`
	ExperimentalFeatures bool `json:"experimentalFeatures"`
	TwoFactorEnabled     bool `json:"twoFactorEnabled"`
	BackupCodesCreated   bool `json:"backupCodesCreated"`
}

// PlexTVDevice represents a device registered on PlexTV. It is the response from the /api/v2/devices endpoint.
type PlexTVDevice struct {
	CreatedAt        time.Time `json:"createdAt"`
	LastSeenAt       time.Time `json:"lastSeenAt"`
	Platform         *string   `json:"platform"`
	Device           *string   `json:"device"`
	Model            *string   `json:"model"`
	Vendor           *string   `json:"vendor"`
	Provides         *string   `json:"provides"`
	PlatformVersion  *string   `json:"platformVersion"`
	ScreenDensity    *int      `json:"screenDensity"`
	Name             string    `json:"name"`
	Product          string    `json:"product"`
	Token            string    `json:"token"`
	PublicAddress    string    `json:"publicAddress"`
	ClientIdentifier string    `json:"clientIdentifier"`
	ScreenResolution string    `json:"screenResolution"`
	SyncLists        []struct {
		Version            int `json:"version"`
		ItemsCompleteCount int `json:"itemsCompleteCount"`
		TotalSize          int `json:"totalSize"`
	} `json:"syncLists"`
	Connections []struct {
		Uri string `json:"uri"`
	} `json:"connections"`
	Id       int  `json:"id"`
	Presence bool `json:"presence"`
}

// RegisteredDevice represents a registered device on Plex.
type RegisteredDevice struct {
	CreatedAt  PlexTimestamp `xml:"createdAt,attr"`
	LastSeenAt PlexTimestamp `xml:"lastSeenAt,attr"`
	SyncList   *SyncList     `xml:"SyncList"`
	// Attributes
	Name            string `xml:"name,attr"`
	PublicAddress   string `xml:"publicAddress,attr"`
	Product         string `xml:"product,attr"`
	ProductVersion  string `xml:"productVersion,attr"`
	Platform        string `xml:"platform,attr"`
	PlatformVersion string `xml:"platformVersion,attr"`
	Device          string `xml:"device,attr"`
	Model           string `xml:"model,attr"`
	Vendor          string `xml:"vendor,attr"`
	Provides        string `xml:"provides,attr"`
	ClientID        string `xml:"clientIdentifier,attr"`
	Version         string `xml:"version,attr"`
	ID              string `xml:"id,attr"`
	Token           string `xml:"token,attr"`
	ScreenRes       string `xml:"screenResolution,attr"`
	ScreenDensity   string `xml:"screenDensity,attr"`

	// Optional nested elements
	Connections []Connection `xml:"Connection"`
}

type Connection struct {
	URI string `xml:"uri,attr"`
}

type SyncList struct {
	ItemsComplete int `xml:"itemsCompleteCount,attr"`
	TotalSize     int `xml:"totalSize,attr"`
	Version       int `xml:"version,attr"`
}

// Resource represents a registered device on Plex. It's the response to /api/v2/resources endpoint.
//
// Use the AccessToken to interact with the PMS instance and the list of connection URLs to locate it.
// Connections labeled as local should be preferred over those that are not,
// and relay should only be used as a last resort as bandwidth on relay connections is limited.
type Resource struct {
	CreatedAt        time.Time   `json:"createdAt"`
	LastSeenAt       time.Time   `json:"lastSeenAt"`
	OwnerId          interface{} `json:"ownerId"`
	SourceTitle      interface{} `json:"sourceTitle"`
	Name             string      `json:"name"`
	Product          string      `json:"product"`
	ProductVersion   string      `json:"productVersion"`
	Platform         string      `json:"platform"`
	PlatformVersion  string      `json:"platformVersion"`
	Device           string      `json:"device"`
	ClientIdentifier string      `json:"clientIdentifier"`
	Provides         string      `json:"provides"`
	PublicAddress    string      `json:"publicAddress"`
	AccessToken      string      `json:"accessToken"`
	Connections      []struct {
		Protocol string `json:"protocol"`
		Address  string `json:"address"`
		Uri      string `json:"uri"`
		Port     int    `json:"port"`
		Local    bool   `json:"local"`
		Relay    bool   `json:"relay"`
		IPv6     bool   `json:"IPv6"`
	} `json:"connections"`
	SearchEnabled          bool `json:"searchEnabled"`
	Owned                  bool `json:"owned"`
	Home                   bool `json:"home"`
	Synced                 bool `json:"synced"`
	Relay                  bool `json:"relay"`
	Presence               bool `json:"presence"`
	HttpsRequired          bool `json:"httpsRequired"`
	PublicAddressMatches   bool `json:"publicAddressMatches"`
	DnsRebindingProtection bool `json:"dnsRebindingProtection"`
	NatLoopbackSupported   bool `json:"natLoopbackSupported"`
}
