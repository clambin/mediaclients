package plextv

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

// TokenSource creates a Plex authentication Token.
type TokenSource interface {
	Token(ctx context.Context) (Token, error)
}

// TokenSource returns an [TokenSource] that can be passed to plex.New() to create an authenticated Plex client.
func (c Config) TokenSource(opts ...TokenSourceOption) TokenSource {
	cfg := tokenSourceConfiguration{
		config: &c,
		logger: slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg.tokenSource()
}

// TokenSourceOption provides the configuration to create the desired TokenSource.
type TokenSourceOption func(*tokenSourceConfiguration)

// WithLogger configures an optional logger.
func WithLogger(logger *slog.Logger) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.logger = logger
	}
}

// WithToken configures a TokenSource to use an existing, fixed token.
func WithToken(token Token) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.token = token
	}
}

// WithCredentials uses the given credentials to register a device and get a token.
func WithCredentials(username, password string) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.registrar = tokenSourceFunc(func(ctx context.Context) (Token, error) {
			return c.config.RegisterWithCredentials(ctx, username, password)
		})
	}
}

// WithPIN uses the PIN flow to register a device and get a token.
// Use the callback to inform the user of the PIN URL and to confirm the PIN.
func WithPIN(cb func(PINResponse, string), pollInterval time.Duration) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.registrar = tokenSourceFunc(func(ctx context.Context) (Token, error) {
			return c.config.RegisterWithPIN(ctx, cb, pollInterval)
		})
	}
}

// WithJWT configures the TokenSource to use a JWT token to request a token.
//
// Using JWT requires persistence. storePath is the path to where the secure data will be stored;
// passphrase is the passphrase used to encrypt the secure data.
//
// Note: once you set up JWT authentication, you can't use credentials or PIN anymore for the device's ClientIdentifier.
// If you lose the secure data stored at storePath, you'll need to re-register the device with a new ClientIdentifier.
//
// Note 2: JWT tokens are relatively new. And in my experience, their use is not always intuitive. The main advantage
// right now is that a JWT-enabled TokenSource does not reregister each time it starts.  But it comes with an
// operational burden (the need for persistent data) and some risk. Approach with caution.
// See [Config.JWTToken] for more details.
func WithJWT(store JWTSecureDataStore) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.vault = &jwtDataStore{
			vault:    store,
			clientID: c.config.ClientID,
		}
	}
}

type tokenSourceConfiguration struct {
	registrar TokenSource
	config    *Config
	logger    *slog.Logger
	vault     *jwtDataStore
	token     Token
}

func (c tokenSourceConfiguration) tokenSource() TokenSource {
	// if we have a fixed token, we're done.
	if c.token != "" {
		return fixedTokenSource{token: c.token}
	}

	// the registrar is the basis of every other tokenSource.
	source := c.registrar

	// If we're using JWT tokens, use jwtTokenSource to register the device (if needed) and get a JWT token.
	if c.vault != nil {
		source = &jwtTokenSource{
			registrar: source,
			vault:     c.vault,
			logger:    c.logger.With("component", "jwtTokenSource"),
			config:    c.config,
		}
	}

	// return the final token source: cachingTokenSource -> [ jwtTokenSource -> ] registrar
	return &cachingTokenSource{
		tokenSource: source,
	}
}

var _ TokenSource = (*tokenSourceFunc)(nil)

// tokenSourceFunc is an adapter to convert a function with the correct signature into an TokenSource.
type tokenSourceFunc func(context.Context) (Token, error)

func (a tokenSourceFunc) Token(ctx context.Context) (Token, error) {
	return a(ctx)
}

var (
	_ TokenSource = fixedTokenSource{}
	_ TokenSource = (*cachingTokenSource)(nil)
	_ TokenSource = (*jwtTokenSource)(nil)
)

// fixedTokenSource returns a fixed token.
type fixedTokenSource struct {
	token Token
}

func (f fixedTokenSource) Token(_ context.Context) (Token, error) {
	return f.token, nil
}

// A cachingTokenSource caches the token obtained by the underlying TokenSource.
type cachingTokenSource struct {
	tokenSource TokenSource
	token       *Token
	lock        sync.Mutex
}

func (s *cachingTokenSource) Token(ctx context.Context) (Token, error) {
	if s.tokenSource == nil {
		return "", ErrNoTokenSource
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	// if we have a valid token cached, return it
	// Note: IsValid parses the token on each call, which is quite expensive. But we must test it here, since a JWT may expire.
	if s.token != nil && s.token.IsValid() {
		return *s.token, nil
	}

	// no valid token cached, get a new one
	token, err := s.tokenSource.Token(ctx)
	if err != nil {
		return "", err
	}
	s.token = &token
	return token, nil
}

// A jwtTokenSource returns a Plex JWT Token. If needed, it registers a new device using the configured registrar.
type jwtTokenSource struct {
	registrar  TokenSource
	vault      secureDataVault
	logger     *slog.Logger
	config     *Config
	secureData JWTSecureData
	init       untilSuccessful
}

type secureDataVault interface {
	Load() (JWTSecureData, error)
	Save(JWTSecureData) error
}

func (s *jwtTokenSource) Token(ctx context.Context) (Token, error) {
	if err := s.init.Do(func() error { return s.initialize(ctx) }); err != nil {
		return "", fmt.Errorf("init: %w", err)
	}
	return s.config.JWTToken(ctx, s.secureData.PrivateKey, s.secureData.KeyID)
}

func (s *jwtTokenSource) initialize(ctx context.Context) (err error) {
	// set up logger if not done already
	if s.logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	}

	// load the client's jwt token data
	s.secureData, err = s.vault.Load()
	switch {
	case err == nil:
		// valid secure data found
		return nil
	case errors.Is(err, ErrInvalidClientID):
		// secure data found, but not for this client ID
		s.logger.Warn("client ID mismatch, secure data found but not for this client. Overwriting secure data",
			slog.String("want", s.config.ClientID),
			slog.String("found", s.secureData.ClientID),
		)
	case errors.Is(err, os.ErrNotExist):
		s.logger.Info("no secure data found. Initializing")
	default:
		return fmt.Errorf("load token data: %w", err)
	}

	s.logger.Debug("registering device")

	if s.registrar == nil {
		return ErrNoTokenSource
	}

	var token Token
	token, err = s.registrar.Token(ctx)
	if err != nil {
		return fmt.Errorf("registrar: %w", err)
	}

	s.logger.Debug("device registered successfully")

	s.secureData.ClientID = s.config.ClientID
	if s.secureData.PrivateKey, s.secureData.KeyID, err = s.config.GenerateAndUploadPublicKey(ctx, token); err != nil {
		return fmt.Errorf("publish key: %w", err)
	}

	s.logger.Debug("public key published successfully")

	if err = s.vault.Save(s.secureData); err != nil {
		return fmt.Errorf("save token data: %w", err)
	}

	s.logger.Debug("token data saved successfully")
	return nil
}

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
