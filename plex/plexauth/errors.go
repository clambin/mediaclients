package plexauth

import (
	"cmp"
	"fmt"
)

// TODO: do we need ErrInvalidToken? Do we need more?
// TODO: should we add more info to PlexError? HTTP Code, etc?

var (
	ErrInvalidToken = fmt.Errorf("invalid token")
)

var _ error = (*PlexError)(nil)

type PlexError struct {
	Reason     string
	Status     string
	Body       []byte
	StatusCode int
}

func (h *PlexError) Error() string {
	return "plex: " + cmp.Or(h.Reason, h.Status)
}

type AuthError struct {
	*PlexError
}
