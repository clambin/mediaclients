package plexauth

import (
	"encoding/xml"
	"strconv"
	"time"
)

type PINResponse struct {
	CreatedAt        time.Time   `json:"createdAt"`
	ExpiresAt        time.Time   `json:"expiresAt"`
	AuthToken        interface{} `json:"authToken"`
	NewRegistration  interface{} `json:"newRegistration"`
	Code             string      `json:"code"`
	Product          string      `json:"product"`
	Qr               string      `json:"qr"`
	ClientIdentifier string      `json:"clientIdentifier"`
	Location         struct {
		Code                       string `json:"code"`
		ContinentCode              string `json:"continent_code"`
		Country                    string `json:"country"`
		City                       string `json:"city"`
		TimeZone                   string `json:"time_zone"`
		PostalCode                 string `json:"postal_code"`
		Subdivisions               string `json:"subdivisions"`
		Coordinates                string `json:"coordinates"`
		EuropeanUnionMember        bool   `json:"european_union_member"`
		InPrivacyRestrictedCountry bool   `json:"in_privacy_restricted_country"`
		InPrivacyRestrictedRegion  bool   `json:"in_privacy_restricted_region"`
	} `json:"location"`
	Id        int  `json:"id"`
	ExpiresIn int  `json:"expiresIn"`
	Trusted   bool `json:"trusted"`
}

type ValidatePINResponse struct {
	CreatedAt        time.Time `json:"createdAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
	NewRegistration  any       `json:"newRegistration"`
	AuthToken        *string   `json:"authToken"`
	Code             string    `json:"code"`
	Product          string    `json:"product"`
	Qr               string    `json:"qr"`
	ClientIdentifier string    `json:"clientIdentifier"`
	Location         struct {
		Code                       string `json:"code"`
		ContinentCode              string `json:"continent_code"`
		Country                    string `json:"country"`
		City                       string `json:"city"`
		TimeZone                   string `json:"time_zone"`
		PostalCode                 string `json:"postal_code"`
		Subdivisions               string `json:"subdivisions"`
		Coordinates                string `json:"coordinates"`
		EuropeanUnionMember        bool   `json:"european_union_member"`
		InPrivacyRestrictedCountry bool   `json:"in_privacy_restricted_country"`
		InPrivacyRestrictedRegion  bool   `json:"in_privacy_restricted_region"`
	} `json:"location"`
	Id        int  `json:"id"`
	ExpiresIn int  `json:"expiresIn"`
	Trusted   bool `json:"trusted"`
}

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

type PlexTimestamp time.Time

func (t *PlexTimestamp) UnmarshalXMLAttr(attr xml.Attr) error {
	epoc, err := strconv.ParseInt(attr.Value, 10, 64)
	if err != nil {
		return err
	}
	*t = PlexTimestamp(time.Unix(epoc, 0))
	return nil
}
