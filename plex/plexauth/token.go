package plexauth

import (
	"errors"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

type Token string

// String returns the token as a string.
func (t Token) String() string {
	return string(t)
}

// IsLegacy returns true if the token is a legacy token (20-character string).
func (t Token) IsLegacy() bool {
	return len(t) == 20
}

// IsJWT returns true if the token is a JWT.
// Note: returns true even if the token is invalid (e.g. expired).
func (t Token) IsJWT() bool {
	_, err := jwt.Parse([]byte(t), jwt.WithVerify(false))
	return err == nil || errors.Is(err, jwt.TokenExpiredError())
}

// IsValid returns true if the token is valid.
// Note: for JWT, the token is valid if it is not expired. The signature, if present, is not verified.
func (t Token) IsValid() bool {
	if t.IsLegacy() {
		return true
	}
	tok, err := jwt.Parse([]byte(t), jwt.WithVerify(false))
	if err != nil {
		return false
	}
	exp, ok := tok.Expiration()
	if !ok {
		return false
	}
	return exp.After(time.Now())
}
