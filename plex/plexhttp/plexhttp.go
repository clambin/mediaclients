package plexhttp

import (
	"encoding/json"
	"net/http"
)

var _ error = HTTPError{}

type HTTPError struct {
	StatusCode int
	StatusText string
	Body       string
}

type AuthError struct {
	HTTPError
}

func Parse(r *http.Response) error {
	var errorBody struct {
		Error string `json:"error"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&errorBody)
	}
	if r.StatusCode == http.StatusUnauthorized {
		return AuthError{HTTPError: HTTPError{StatusCode: r.StatusCode, StatusText: r.Status, Body: errorBody.Error}}
	}
	return HTTPError{StatusCode: r.StatusCode, StatusText: r.Status, Body: errorBody.Error}
}

func (h HTTPError) Error() string {
	body := /*fmt.Sprintf("%03d - %s", h.StatusCode, */ h.StatusText //)
	if h.Body != "" {
		body += ": " + h.Body
	}
	return body
}
