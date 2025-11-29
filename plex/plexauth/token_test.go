package plexauth

import (
	"testing"
	"time"
)

func TestToken(t *testing.T) {
	tests := []struct {
		name       string
		token      Token
		wantString string
		wantValid  bool
	}{
		{"AuthToken", AuthToken("abcd1234"), "abcd1234", true},
		{"JWTToken", JWTToken{expiration: time.Now().Add(24 * time.Hour), AuthToken: "abcd1234"}, "abcd1234", true},
		{"expired JWTToken", JWTToken{AuthToken: "abcd1234"}, "abcd1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.String(); got != tt.wantString {
				t.Errorf("Token.String() = %v, want %v", got, tt.wantString)
			}
			if got := tt.token.IsValid(); got != tt.wantValid {
				t.Errorf("Token.IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}
