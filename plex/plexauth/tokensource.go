package plexauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

// TokenSourceOption provides the configuration to determine the desired TokenSource.
type TokenSourceOption func(*tokenSourceConfiguration)

// WithToken configures a TokenSource to use an existing, fixed token.
func WithToken(token AuthToken) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.token = token
	}
}

// WithCredentials uses the given credentials to register a device and get an auth token.
func WithCredentials(username, password string) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.registrar = authTokenSourceFunc(func(ctx context.Context) (AuthToken, error) {
			return c.config.RegisterWithCredentials(ctx, username, password)
		})
	}
}

// WithPIN uses the PIN flow to register a device and get an auth token.
// Use the callback to inform the user of the PIN URL and to confirm the PIN.
func WithPIN(cb func(PINResponse, string), pollInterval time.Duration) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.registrar = authTokenSourceFunc(func(ctx context.Context) (AuthToken, error) {
			return c.config.RegisterWithPIN(ctx, cb, pollInterval)
		})
	}
}

// WithLogger configures an optional logger.
func WithLogger(logger *slog.Logger) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.logger = logger
	}
}

// WithJWT configures the TokenSource to use a JWT token to request an auth token.
//
// Using JWT requires persistence. storePath is the path to where the secure data will be stored;
// passphrase is the passphrase used to encrypt the secure data.
//
// Note: once you set up JWT authentication, you can't use credentials or PIN anymore for the device's ClientIdentifier.
// If you lose the secure data stored at storePath, you'll need to re-register the device with a new ClientIdentifier.
//
// See [Config.JWTToken] for more details.
func WithJWT(storePath, passphrase string) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.vault = newJWTDataStore(storePath, passphrase, c.config.ClientID)
	}
}

// WithPMS specifies the name of the Plex Media Server for which to obtain an auth token.
// If not specified, the first Plex Media Server found in the account will be used. If you have multiple Plex Media Servers,
// you should specify the name of the one you want to use.
func WithPMS(pmsName string) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.pmsName = pmsName
		c.usePMSToken = true
	}
}

type tokenSourceConfiguration struct {
	config      *Config
	registrar   AuthTokenSource
	token       AuthToken
	logger      *slog.Logger
	vault       *jwtDataStore
	pmsName     string
	usePMSToken bool
}

func (c tokenSourceConfiguration) TokenSource() AuthTokenSource {
	if c.token != "" {
		return fixedTokenSource{token: c.token}
	}

	// the registrar is the basis of every TokenSource.
	source := c.registrar

	// If we're not using the Plex Media Server token, we're done.
	if !c.usePMSToken {
		// cache the token. we only register once
		return &cachingTokenSource{authTokenSource: source}
	}

	// If we're using JWT tokens, use jwtTokenSource to register the device (if needed) and obtain a JWT token.
	if c.vault != nil {
		jwts := &jwtTokenSource{
			registrar: source,
			vault:     c.vault,
			logger:    c.logger.With("component", "jwtTokenSource"),
			config:    c.config,
		}
		source = authTokenSourceFunc(func(ctx context.Context) (AuthToken, error) {
			t, err := jwts.token(ctx)
			if err != nil {
				return "", err
			}
			return AuthToken(t.String()), nil
		})
	}

	return &cachingTokenSource{
		authTokenSource: &pmsTokenSource{
			tokenSource: source,
			config:      c.config,
			pmsName:     c.pmsName,
			logger:      c.logger.With("component", "pmsTokenSource"),
		},
	}
}

// A AuthTokenSource returns a Plex authentication Token
type AuthTokenSource interface {
	Token(ctx context.Context) (AuthToken, error)
}

var _ AuthTokenSource = (*authTokenSourceFunc)(nil)

// authTokenSourceFunc is an adapter to convert a function with the correct signature into an AuthTokenSource.
type authTokenSourceFunc func(context.Context) (AuthToken, error)

func (a authTokenSourceFunc) Token(ctx context.Context) (AuthToken, error) {
	return a(ctx)
}

var (
	_ AuthTokenSource = fixedTokenSource{}
	_ AuthTokenSource = (*cachingTokenSource)(nil)
	_ AuthTokenSource = (*pmsTokenSource)(nil)
)

// fixedTokenSource returns a fixed token.
type fixedTokenSource struct {
	token AuthToken
}

func (f fixedTokenSource) Token(_ context.Context) (AuthToken, error) {
	return f.token, nil
}

// A cachingTokenSource caches the auth token obtained by the underlying AuthTokenSource.
type cachingTokenSource struct {
	authTokenSource AuthTokenSource
	authToken       AuthToken
	lock            sync.Mutex
}

func (s *cachingTokenSource) Token(ctx context.Context) (AuthToken, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var err error
	if !s.authToken.IsValid() {
		s.authToken, err = s.authTokenSource.Token(ctx)
	}
	return s.authToken, err
}

// A pmsTokenSource returns the Plex authentication token for a given Plex Media Server.
type pmsTokenSource struct {
	tokenSource AuthTokenSource
	config      *Config
	pmsName     string
	logger      *slog.Logger
}

func (p pmsTokenSource) Token(ctx context.Context) (AuthToken, error) {
	// get a token to access the Plex Cloud API
	token, err := p.tokenSource.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	p.logger.Debug("got cloud token")
	var mediaServers []RegisteredDevice
	if mediaServers, err = p.config.MediaServers(ctx, token); err == nil {
		for _, server := range mediaServers {
			p.logger.Debug("media server found", "name", server.Name)
			if server.Name == p.pmsName || p.pmsName == "" {
				p.logger.Debug("media server matched")
				return AuthToken(server.Token), nil
			}
		}
		err = fmt.Errorf("media server %q not found", p.pmsName)
	}
	p.logger.Debug("no media server found", "err", err)
	return "", fmt.Errorf("media servers: %w", err)
}

// A jwtTokenSource returns a Plex JWT Token. If needed, it registers a new device using the configured Registrar.
type jwtTokenSource struct {
	registrar   AuthTokenSource
	vault       secureDataVault
	logger      *slog.Logger
	config      *Config
	secureData  jwtSecureData
	initialized bool
	lock        sync.Mutex
}

type secureDataVault interface {
	Load() (jwtSecureData, error)
	Save(jwtSecureData) error
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

	var authToken AuthToken
	authToken, err = s.registrar.Token(ctx)
	if err != nil {
		return fmt.Errorf("registrar: %w", err)
	}

	s.logger.Debug("device registered successfully")

	s.secureData.ClientID = s.config.ClientID
	if s.secureData.PrivateKey, s.secureData.KeyID, err = s.config.GenerateAndUploadPublicKey(ctx, authToken); err != nil {
		return fmt.Errorf("publish key: %w", err)
	}

	s.logger.Debug("public key published successfully")

	if err = s.vault.Save(s.secureData); err != nil {
		return fmt.Errorf("save token data: %w", err)
	}

	s.logger.Debug("token data saved successfully")
	return nil
}

func (s *jwtTokenSource) token(ctx context.Context) (token JWTToken, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.initialized {
		if err = s.initialize(ctx); err != nil {
			return token, fmt.Errorf("init: %w", err)
		}
		s.initialized = true
	}
	return s.config.JWTToken(ctx, s.secureData.PrivateKey, s.secureData.KeyID)
}
