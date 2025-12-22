package plexauth

import (
	"time"
)

// PINResponse is the response from the PINRequest endpoint
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

// ValidatePINResponse is the response from the ValidatePIN endpoint.
// When AuthToken is not null, the user has been authenticated.
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
