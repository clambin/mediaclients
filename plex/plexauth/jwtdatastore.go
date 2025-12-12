package plexauth

import (
	"crypto/ed25519"
	"fmt"

	"github.com/clambin/mediaclients/plex/internal/vault"
)

var (
	ErrInvalidClientID = fmt.Errorf("data store contains invalid client ID")
)

// jwtSecureData contains the data required to request a JWTToken.
type jwtSecureData struct {
	KeyID      string             `json:"key-id"`
	ClientID   string             `json:"client-id"`
	PrivateKey ed25519.PrivateKey `json:"private-key"`
}

type jwtDataStore struct {
	vault    *vault.Vault[jwtSecureData]
	clientID string
}

func newJWTDataStore(filePath string, passphrase string, clientID string) *jwtDataStore {
	return &jwtDataStore{
		vault:    vault.New[jwtSecureData](filePath, passphrase),
		clientID: clientID,
	}
}

// Save saves the given data to the data store.
func (s *jwtDataStore) Save(data jwtSecureData) error {
	return s.vault.Save(data)
}

// Load loads the data from the data store. It returns ErrInvalidClientID if the data's client ID does not match.
func (s *jwtDataStore) Load() (jwtSecureData, error) {
	data, err := s.vault.Load()
	if err != nil {
		return jwtSecureData{}, err
	}
	if data.ClientID != s.clientID {
		err = ErrInvalidClientID
	}
	return data, err
}
