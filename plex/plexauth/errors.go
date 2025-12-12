package plexauth

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrInvalidToken = fmt.Errorf("invalid token")

var _ error = (*HTTPError)(nil)

// HTTPError represents an error returned by Plex.
type HTTPError struct {
	// Reason contains the error message(s) returned by Plex, if present, otherwise Status.
	Reason string
	// Status is the HTTP status returned by Plex.
	Status string
	// Body contains the raw response body returned by Plex.
	Body []byte
	// StatusCode is the HTTP status code returned by Plex.
	StatusCode int
}

func (h *HTTPError) Error() string {
	return "plex: " + cmp.Or(h.Reason, h.Status)
}

func ParsePlexError(r *http.Response) error {
	var (
		errorBody struct {
			Error  string `json:"error"`
			Errors []struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
				Status  int    `json:"status"`
			} `json:"errors"`
		}
		buf bytes.Buffer
	)

	var reason string
	if r.Body != nil {
		_ = json.NewDecoder(io.TeeReader(r.Body, &buf)).Decode(&errorBody)
	}

	switch {
	case errorBody.Error != "":
		// single error message
		reason = errorBody.Error
	case len(errorBody.Errors) > 0:
		// multi-message error
		messages := make([]string, len(errorBody.Errors))
		for i, entry := range errorBody.Errors {
			messages[i] = fmt.Sprintf("%d - %s", entry.Code, entry.Message)
		}
		reason = strings.Join(messages, ", ")
	default:
		reason = r.Status
	}

	return &HTTPError{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Reason:     reason,
		Body:       buf.Bytes()}
}
