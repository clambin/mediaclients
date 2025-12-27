package plextv

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// RegisterWithCredentials registers a device using username/password credentials and returns a Token.
func (c Config) RegisterWithCredentials(ctx context.Context, username, password string) (Token, error) {
	// credentials are passed in the request body in url-encoded form
	v := make(url.Values)
	v.Set("user[login]", username)
	v.Set("user[password]", password)

	// call the auth endpoint
	resp, err := c.do(ctx, http.MethodPost, c.URL+"/users/sign_in.xml", strings.NewReader(v.Encode()), http.StatusCreated, func(req *http.Request) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/xml")
		c.Device.populateRequest(req)
	})
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// a successful response contains an XML document with an authentication token
	var authResponse struct {
		XMLName             xml.Name `xml:"user"`
		AuthenticationToken string   `xml:"authenticationToken,attr"`
	}
	err = xml.NewDecoder(resp.Body).Decode(&authResponse)
	return Token(authResponse.AuthenticationToken), err
}
