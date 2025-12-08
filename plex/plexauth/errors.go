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

var _ error = (*ErrPlex)(nil)

type ErrPlex struct {
	Reason     string
	Status     string
	Body       []byte
	StatusCode int
}

func (h *ErrPlex) Error() string {
	return "plex: " + cmp.Or(h.Reason, h.Status)
}

type AuthError struct {
	*ErrPlex
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

	return &ErrPlex{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Reason:     reason,
		Body:       buf.Bytes()}
}
