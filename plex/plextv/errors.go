package plextv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrInvalidToken indicates that a Config method was passed an invalid token.
	ErrInvalidToken = fmt.Errorf("invalid token")
	// ErrInvalidClientID indicates that the data in a JWTSecureDataStore is for a different Client ID
	// than the one used to register the device.
	// The stored private key and public key ID can't be used to generate a new JWT token.
	// You will need to re-register the device with a new Client ID.
	ErrInvalidClientID = fmt.Errorf("data store contains invalid client ID")
	// ErrNoTokenSource indicates that a token source needs a child token source, but none was provided.
	// A typical example is JWTTokenSource or PMSTokenSource needing a registrar to get a legacy token,
	// but none is provided in [Config.TokenSource].
	ErrNoTokenSource = errors.New("no token source provided")
	// ErrUnauthorized indicates that plex.tv could not authenticate the user.
	ErrUnauthorized = errors.New("user could not be authenticated")
	// ErrTooManyRequests indicates that the plex.tv API rate limit has been reached.
	// Typically, this only happens when generating a new JWT token.
	ErrTooManyRequests = errors.New("too many requests")
	// ErrJWKMissing indicates that the JWT token request could not find a public key to verify the JWT request.
	// Typically, this means the Registered Device for the associated Client ID has been removed from Plex's Registered Devices.
	// You will need to re-register the device with a new Client ID and a new public/private key pair.
	ErrJWKMissing = errors.New("jwk missing. no public key to verify jwt request")
)

var _ error = &PlexError{}

type PlexError struct {
	errors     error
	Status     string
	Body       []byte
	StatusCode int
}

func (p *PlexError) Error() string {
	txt := p.Status
	if p.errors != nil {
		txt = p.errors.Error()
	}
	return "plex: " + txt
}

func (p *PlexError) Unwrap() error {
	return p.errors
}

// TODO: probably more errors that could help users/apps figure out what's going on
var plexErrors = map[int]error{
	1001: ErrUnauthorized,
	1003: ErrTooManyRequests,
	1097: ErrJWKMissing,
}

// ParsePlexError parses the errors text returned by plex.tv and return a PlexError.
func ParsePlexError(r *http.Response) error {
	var errorBody struct {
		Error  string `json:"error"`
		Errors []struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
			Status  int    `json:"status"`
		} `json:"errors"`
	}

	var buf bytes.Buffer
	if r.Body != nil {
		_ = json.NewDecoder(io.TeeReader(r.Body, &buf)).Decode(&errorBody)
	}

	e := PlexError{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Body:       buf.Bytes(),
	}

	switch {
	case errorBody.Error != "":
		// single-message error
		e.errors = errors.New(errorBody.Error)
	case len(errorBody.Errors) > 0:
		// multi-message error
		errs := make([]error, len(errorBody.Errors))
		for i, entry := range errorBody.Errors {
			var ok bool
			if errs[i], ok = plexErrors[entry.Code]; !ok {
				errs[i] = errors.Join(e.errors, fmt.Errorf("%d - %s", entry.Code, entry.Message))
			}
		}
		e.errors = errors.Join(errs...)
	}
	return &e
}
