package plexauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/afero"
)

// JWTSecureData contains the data required to request a JWTToken.
type JWTSecureData struct {
	KeyID      string             `json:"key-id"`
	ClientID   string             `json:"client-id"`
	PrivateKey ed25519.PrivateKey `json:"private-key"`
}

// JWTDataStore provides a basic way to store the Client ID, private key and public key ID to request a JWTToken.
// This implementation uses AES256 to encrypt the data at rest.
type JWTDataStore struct {
	store         afero.Fs
	cachedData    *JWTSecureData
	filepath      string
	encryptionKey []byte
	lock          sync.Mutex
}

func NewJWTDataStore(filePath string, passphrase string) *JWTDataStore {
	hash := sha256.Sum256([]byte(passphrase))
	return &JWTDataStore{
		encryptionKey: hash[:],
		filepath:      filePath,
		store:         afero.NewOsFs(),
	}
}

func (s *JWTDataStore) Get() (JWTSecureData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// if we have the data cached, return it
	if s.cachedData != nil {
		return *s.cachedData, nil
	}

	encryptedData, err := afero.ReadFile(s.store, s.filepath)
	if err != nil {
		return JWTSecureData{}, err
	}
	decryptedData, err := decryptAES(encryptedData, s.encryptionKey)
	if err != nil {
		return JWTSecureData{}, fmt.Errorf("decrypt: %w", err)
	}
	var data JWTSecureData
	if err = json.Unmarshal(decryptedData, &data); err != nil {
		return JWTSecureData{}, fmt.Errorf("decode: %w", err)
	}
	s.cachedData = &data
	return data, nil
}

func (s *JWTDataStore) Set(data JWTSecureData) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	decryptedData, _ := json.Marshal(data)
	encryptedData, err := encryptAES(decryptedData, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}
	if err = afero.WriteFile(s.store, s.filepath, encryptedData, 0600); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	s.cachedData = &data
	return nil
}

// AES encryption
func encryptAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return aesGCM.Seal(nonce, nonce, data, nil), nil
}

// AES decryption
func decryptAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("invalid ciphertext")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
