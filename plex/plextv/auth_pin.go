package plextv

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

// RegisterWithPIN is a helper function that registers a device using the PIN authentication flow and gets a Token.
// It requests a PIN from Plex, calls the callback with the PINResponse and PIN URL and blocks until the PIN is confirmed.
// Use a context with a timeout to ensure it doesn't block forever.
//
// The callback can be used to inform the user/application of the URL to confirm the PINRequest.
func (c Config) RegisterWithPIN(ctx context.Context, callback func(PINResponse, string), pollInterval time.Duration) (token Token, err error) {
	pinResponse, pinURL, err := c.PINRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("pin: %w", err)
	}
	callback(pinResponse, pinURL)
	for {
		if token, _, err = c.ValidatePIN(ctx, pinResponse.Id); err == nil && token.IsValid() {
			return token, nil
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(cmp.Or(pollInterval, 15*time.Second)):
		}
	}
}

// PINRequest requests a PINRequest from Plex.
//
// Currently only supports strong=false. Support for strong=true is planned, but this requires https://app.plex.tv/auth,
// which is currently broken.
func (c Config) PINRequest(ctx context.Context) (PINResponse, string, error) {
	resp, err := c.do(ctx, http.MethodPost, c.V2URL+"/api/v2/pins" /*?strong=false"*/, nil, http.StatusCreated)
	if err != nil {
		return PINResponse{}, "", fmt.Errorf("pin request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var response PINResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return PINResponse{}, "", fmt.Errorf("decode: %w", err)
	}
	// legacy endpoint. once https://app.plex.tv/auth is fixed, this can be adapted accordingly.
	return response, "https://plex.tv/pin?pin=" + response.Code, nil
}

// ValidatePIN checks if the user has confirmed the PINRequest.  It returns the full Plex response.
// When the user has confirmed the PINRequest, the AuthToken field will be populated.
func (c Config) ValidatePIN(ctx context.Context, id int) (Token, ValidatePINResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, c.V2URL+"/api/v2/pins/"+strconv.Itoa(id), nil, http.StatusOK)
	if err != nil {
		return "", ValidatePINResponse{}, fmt.Errorf("validate pin: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var response ValidatePINResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", ValidatePINResponse{}, fmt.Errorf("decode: %w", err)
	}
	var token Token
	if response.AuthToken != nil {
		token = Token(*response.AuthToken)
	}
	return token, response, err
}
