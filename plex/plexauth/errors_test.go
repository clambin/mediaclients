package plexauth

import (
	"bytes"
	"errors"
	"fmt"
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
			name: "single error",
			resp: &http.Response{
				Status:     "400 Bad Request",
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"invalid input"}`)),
			},
			wantErrStr: "plex: invalid input",
		},
		{
			name: "multi error with 1 error",
			resp: &http.Response{
				Status:     "401 Unauthorized",
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewBufferString(`{"errors": [ {"code":1001, "message": "invalid user"} ] }`)),
			},
			wantErrStr: "plex: user could not be authenticated",
		},
		{
			name: "multi error with multiple errors",
			resp: &http.Response{
				Status:     "401 Unauthorized",
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewBufferString(`{"errors": [ {"code":1001, "message": "invalid user"}, {"code":-1, "message": "I just don't feel like it"} ] }`)),
			},
			wantErrStr: "plex: user could not be authenticated\n-1 - I just don't feel like it",
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

func TestPlexError_Unwrap(t *testing.T) {
	r := http.Response{
		StatusCode: http.StatusTooManyRequests,
		Status:     http.StatusText(http.StatusTooManyRequests),
		Body:       io.NopCloser(bytes.NewBufferString(`{"errors": [ {"code":1003, "message": "too many requests"} ] }`)),
	}
	err := ParsePlexError(&r)
	if err == nil {
		t.Fatal("expected error")
	}
	var err2 *PlexError
	if !errors.As(err, &err2) {
		t.Fatal("expected *PlexError")
	}
	fmt.Printf("%T\n", err2.Error())

	if !errors.Is(err, ErrTooManyRequests) {
		t.Fatal("expected error")
	}
}
