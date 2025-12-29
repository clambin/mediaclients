package plextv

import (
	"errors"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

// plex.tv's time may drift from the client's time, meaning that if we verify the token,
// the issue time ("iat") may be in the future. clockSkewTolerance indicated how far off
// the time can be before we consider the token invalid.
//
// One minute is probably way too much.
const clockSkewTolerance = time.Minute

// Token represents a Plex authentication token. It can be either a legacy token (20-character string) or a JWT.
// Note: JWTs currently cannot be used to access Plex Media Servers.
type Token string

// String returns the token as a string.
func (t *Token) String() string {
	if t == nil {
		return ""
	}
	return string(*t)
}

// IsLegacy returns true if the token is a legacy token (20-character string).
func (t *Token) IsLegacy() bool {
	if t == nil {
		return false
	}
	return len(*t) == 20
}

// IsJWT returns true if the token is a JWT.
// Note: returns true even if the token is invalid (e.g. expired).
func (t *Token) IsJWT() bool {
	if t == nil {
		return false
	}
	_, err := t.parseJWT()
	return err == nil || errors.Is(err, jwt.TokenExpiredError())
}

// IsValid returns true if the token is valid.
// Note: for JWT, the token is valid if it is not expired. The signature, if present, is not verified.
func (t *Token) IsValid() bool {
	if t == nil {
		return false
	}
	if t.IsLegacy() {
		return true
	}
	tok, err := t.parseJWT()
	if err != nil {
		return false
	}
	exp, ok := tok.Expiration()
	return ok && exp.After(time.Now())
}

func (t *Token) parseJWT() (jwt.Token, error) {
	return jwt.Parse([]byte(*t), jwt.WithVerify(false), jwt.WithAcceptableSkew(clockSkewTolerance))
}
