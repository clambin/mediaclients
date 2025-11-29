package plexauth

import (
	"time"
)

type Token interface {
	String() string
	IsValid() bool
}

var (
	_ Token = AuthToken("")
	_ Token = JWTToken{}
)

// AuthToken is a Plex authentication token. It gives access to both the Plex Media Server and the Plex Cloud API.
type AuthToken string

func (t AuthToken) String() string {
	return string(t)
}

func (t AuthToken) IsValid() bool {
	return t != ""
}

// JWTToken is a new authentication mechanism introduced in Plex Cloud, based on JSON Web Tokens (JWT).
type JWTToken struct {
	expiration time.Time
	AuthToken
}

func (j JWTToken) IsValid() bool {
	return j.AuthToken.IsValid() && !j.Expired()
}

func (j JWTToken) Expired() bool {
	return time.Now().After(j.expiration)
}
