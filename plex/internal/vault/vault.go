// Package vault provides a simple means of storing secure data on disk.
// Data at rest is encrypted using AES-256-GCM, using a salted hash.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/spf13/afero"
	"golang.org/x/crypto/hkdf"
)

const (
	currentVersion = 1
)

var (
	// hash function to use for deriving the encryption key
	hash = sha256.New
	// size of the salt to use for deriving the encryption key
	saltSize = hash().Size()
)

type ErrInvalidKey struct {
	Err error
}

func (e *ErrInvalidKey) Error() string {
	if e.Err != nil {
		return "invalid key: " + e.Err.Error()
	}
	return "invalid key"
}

func (e *ErrInvalidKey) Unwrap() error {
	return e.Err
}

type Vault[T any] struct {
	content    *T
	fs         afero.Fs
	passphrase string
	filePath   string
	lock       sync.Mutex
}

func New[T any](filePath string, passphrase string) *Vault[T] {
	return newWithFS[T](afero.NewOsFs(), filePath, passphrase)
}

func newWithFS[T any](fs afero.Fs, filePath string, passphrase string) *Vault[T] {
	c := Vault[T]{
		fs:         fs,
		filePath:   filePath,
		passphrase: passphrase,
	}
	return &c
}

func (c *Vault[T]) Load() (T, error) {
	var zero T
	c.lock.Lock()
	defer c.lock.Unlock()

	// if we have the content cached, return it
	if c.content != nil {
		return *c.content, nil
	}

	// read the file
	data, err := afero.ReadFile(c.fs, c.filePath)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return zero, os.ErrNotExist
		}
		return zero, err
	}

	// determine file version
	var versionReq map[string]any
	if err = json.Unmarshal(data, &versionReq); err != nil {
		return zero, fmt.Errorf("unrecognized file format: %w", err)
	}
	version, ok := versionReq["version"].(float64)
	if !ok {
		return zero, fmt.Errorf("unrecognized file format: missing version")
	}
	// decode the file based on the version
	var record content
	switch int(version) {
	case currentVersion:
		if err = json.Unmarshal(data, &record); err != nil {
			return zero, fmt.Errorf("unrecognized file format: decode: %w", err)
		}
	default:
		return zero, fmt.Errorf("unsupported version %d", int(version))
	}

	// create the encryption key based on the file's salt and the passphrase
	encryptionKey, err := deriveEncryptionKey(c.passphrase, record.Salt)
	if err != nil {
		return zero, fmt.Errorf("derive encryption key: %w", err)
	}

	// decrypt the data
	clearData, err := decryptAES(record.Data, encryptionKey[:])
	if err != nil {
		return zero, &ErrInvalidKey{Err: err}
	}

	// decode the data
	var t T
	if err = json.Unmarshal(clearData, &t); err != nil {
		return zero, fmt.Errorf("decode data: %w", err)
	}
	c.content = &t
	return t, nil
}

func (c *Vault[T]) Save(t T) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// record to write to disk
	var record = content{
		Version: currentVersion,
		Salt:    make([]byte, saltSize),
	}

	// generate a new random salt
	if _, err := rand.Read(record.Salt); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	// encode the data
	body, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	// encrypt the data
	encryptionKey, err := deriveEncryptionKey(c.passphrase, record.Salt)
	if err != nil {
		return fmt.Errorf("derive encryption key: %w", err)
	}
	if record.Data, err = encryptAES(body, encryptionKey[:]); err != nil {
		return &ErrInvalidKey{Err: err}
	}

	// encode the record
	if body, err = json.MarshalIndent(record, "", "  "); err != nil {
		return fmt.Errorf("encode record: %w", err)
	}

	// write the file
	if err = afero.WriteFile(c.fs, c.filePath, body, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// cache the content for reading
	c.content = &t

	return nil
}

type content struct {
	Salt    []byte `json:"salt"`
	Data    []byte `json:"data"`
	Version int    `json:"version"`
}

// deriveKey generates the encryption key from a passphrase and a salt value
func deriveEncryptionKey(passphrase string, salt []byte) ([]byte, error) {
	r := hkdf.New(hash, []byte(passphrase), salt, nil)
	key := make([]byte, 32)
	_, err := io.ReadFull(r, key)
	return key, err
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
