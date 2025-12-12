package plexauth

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlexError(t *testing.T) {
	tests := []struct {
		name       string
		resp       *http.Response
		wantErrStr string
	}{
		{
			name: "json body parsed",
			resp: &http.Response{
				Status:     "400 Bad Request",
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"invalid input"}`)),
			},
			wantErrStr: "plex: invalid input",
		},
		{
			name: "auth - json body parsed",
			resp: &http.Response{
				Status:     "401 Unauthorized",
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewBufferString(`{"errors": [ {"code":1001, "message": "invalid user"} ] }`)),
			},
			wantErrStr: "plex: 1001 - invalid user",
		},
		{
			name: "non-json body ignored",
			resp: &http.Response{
				Status:     "500 Internal Server Error",
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("not-json")),
			},
			wantErrStr: "plex: 500 Internal Server Error",
		},
		{
			name: "no body",
			resp: &http.Response{
				Status:     "500 Internal Server Error",
				StatusCode: http.StatusInternalServerError,
			},
			wantErrStr: "plex: 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, ParsePlexError(tt.resp), tt.wantErrStr)
		})
	}
}
