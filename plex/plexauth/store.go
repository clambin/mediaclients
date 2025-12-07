package plexauth

import (
	"crypto/ed25519"

	"github.com/clambin/mediaclients/plex/internal/vault"
)

// jwtSecureData contains the data required to request a JWTToken.
type jwtSecureData struct {
	KeyID      string             `json:"key-id"`
	ClientID   string             `json:"client-id"`
	PrivateKey ed25519.PrivateKey `json:"private-key"`
}

func newJWTDataStore(filePath string, passphrase string) *vault.Vault[jwtSecureData] {
	return vault.New[jwtSecureData](filePath, passphrase)
}
