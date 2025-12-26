package plexauth

import (
	"crypto/ed25519"
	"fmt"
)

var (
	ErrInvalidClientID = fmt.Errorf("data store contains invalid client ID")
)

type JWTSecureDataStore interface {
	Save(data JWTSecureData) error
	Load() (JWTSecureData, error)
}

// JWTSecureData contains the data required to request a JWTToken.
type JWTSecureData struct {
	KeyID      string             `json:"key-id"`
	ClientID   string             `json:"client-id"`
	PrivateKey ed25519.PrivateKey `json:"private-key"`
}

type jwtDataStore struct {
	vault    JWTSecureDataStore
	clientID string
}

// Save saves the given data to the data store.
func (s *jwtDataStore) Save(data JWTSecureData) error {
	return s.vault.Save(data)
}

// Load loads the data from the data store. It returns ErrInvalidClientID if the data's client ID does not match.
func (s *jwtDataStore) Load() (JWTSecureData, error) {
	data, err := s.vault.Load()
	if err != nil {
		return JWTSecureData{}, err
	}
	if data.ClientID != s.clientID {
		err = ErrInvalidClientID
	}
	return data, err
}
