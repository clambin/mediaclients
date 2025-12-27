package plextv

import (
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

func TestToken(t *testing.T) {
	genJWT := func(expiration time.Duration) string {
		now := time.Now()

		tok := jwt.New()
		_ = tok.Set("nonce", "8c460102-eb0b-44f0-8179-38131db6b4f2")
		_ = tok.Set("thumbprint", "OubVBsn-rQWLASHgI1k70zSe0rJP_0IlIe8SgMY8uhk")
		_ = tok.Set(jwt.IssuerKey, "plex.tv")
		_ = tok.Set(jwt.AudienceKey, []string{"plex.tv", "my-client"})
		_ = tok.Set(jwt.IssuedAtKey, now)
		_ = tok.Set(jwt.ExpirationKey, now.Add(expiration))
		_ = tok.Set("user", map[string]any{
			"id":            475814,
			"uuid":          "35c1a6fd2b630943",
			"username":      "clambin",
			"email":         "christophe.lambin@gmail.com",
			"friendly_name": nil,
		})

		serialized, err := jwt.NewSerializer().Serialize(tok)
		if err != nil {
			panic(err)
		}

		return string(serialized)
	}
	validJWT := genJWT(7 * 24 * time.Hour)
	expiredJWT := genJWT(-1 * time.Hour)

	tests := []struct {
		name         string
		token        Token
		wantString   string
		wantIsLegacy bool
		wantIsJWT    bool
		wantValid    bool
	}{
		{"AuthToken", legacyToken, legacyToken, true, false, true},
		{"JWTToken", Token(validJWT), validJWT, false, true, true},
		{"expired JWTToken", Token(expiredJWT), expiredJWT, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.String(); got != tt.wantString {
				t.Errorf("Token.String() = %v, want %v", got, tt.wantString)
			}
			if got := tt.token.IsLegacy(); got != tt.wantIsLegacy {
				t.Errorf("Token.IsLegacy() = %v, want %v", got, tt.wantIsLegacy)
			}
			if got := tt.token.IsJWT(); got != tt.wantIsJWT {
				t.Errorf("Token.IsJWT() = %v, want %v", got, tt.wantIsJWT)
			}
			if got := tt.token.IsValid(); got != tt.wantValid {
				t.Errorf("Token.IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}
