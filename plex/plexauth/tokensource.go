package plexauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// TokenSourceOption provides the configuration to determine the desired TokenSource.
type TokenSourceOption func(*tokenSourceConfiguration)

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

// WithLogger configures an optional logger.
func WithLogger(logger *slog.Logger) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.logger = logger
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
//
// TODO: use an interface to a secure vault so users can override it
func WithJWT(storePath, passphrase string) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.vault = newJWTDataStore(storePath, passphrase, c.config.ClientID)
	}
}

// WithPMS specifies the name of the Plex Media Server for which to obtain a token.
// If not specified, the first Plex Media Server found in the account will be used. If you have multiple Plex Media Servers,
// you should specify the name of the one you want to use.
//
// TODO: this isn't an auth function. auth should just return a PlexTV token (legacy or JWT).
// Move this to a PlexTV client.
func WithPMS(pmsName string) TokenSourceOption {
	return func(c *tokenSourceConfiguration) {
		c.pmsName = pmsName
		c.usePMSToken = true
	}
}

type tokenSourceConfiguration struct {
	config      *Config
	registrar   TokenSource
	token       Token
	logger      *slog.Logger
	vault       *jwtDataStore
	pmsName     string
	usePMSToken bool
}

func (c tokenSourceConfiguration) tokenSource() TokenSource {
	// if we have a fixed token, we're done.
	if c.token != "" {
		return fixedTokenSource{token: c.token}
	}

	// the registrar is the basis of every TokenSource.
	source := c.registrar

	// If we're not using the Plex Media Server token, we're done.
	// cache the registrar: we only need to register once.
	if !c.usePMSToken {
		// cache the token. we only register once
		return &cachingTokenSource{tokenSource: source}
	}

	// If we're using JWT tokens, use jwtTokenSource to register the device (if needed) and obtain a JWT token.
	if c.vault != nil {
		source = &jwtTokenSource{
			registrar: source,
			vault:     c.vault,
			logger:    c.logger.With("component", "jwtTokenSource"),
			config:    c.config,
		}
	}

	// return the final token source: cachingTokenSource -> pmsTokenSource -> [ jwtTokenSource -> ] registrar
	return &cachingTokenSource{
		tokenSource: &pmsTokenSource{
			tokenSource: source,
			config:      c.config,
			pmsName:     c.pmsName,
			logger:      c.logger.With("component", "pmsTokenSource"),
		},
	}
}

// TokenSource creates a Plex authentication Token.
type TokenSource interface {
	Token(ctx context.Context) (Token, error)
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
	_ TokenSource = (*pmsTokenSource)(nil)
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
	token       Token
	init        untilSuccessful
}

func (s *cachingTokenSource) Token(ctx context.Context) (Token, error) {
	if s.tokenSource == nil {
		return "", ErrNoTokenSource
	}
	err := s.init.Do(func() error {
		var err error
		s.token, err = s.tokenSource.Token(ctx)
		return err
	})
	return s.token, err
}

// A jwtTokenSource returns a Plex JWT Token. If needed, it registers a new device using the configured registrar.
type jwtTokenSource struct {
	registrar  TokenSource
	vault      secureDataVault
	logger     *slog.Logger
	config     *Config
	secureData jwtSecureData
	init       untilSuccessful
}

type secureDataVault interface {
	Load() (jwtSecureData, error)
	Save(jwtSecureData) error
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

// A pmsTokenSource returns the Plex authentication token for a given Plex Media Server.
type pmsTokenSource struct {
	tokenSource TokenSource
	config      *Config
	logger      *slog.Logger
	pmsName     string
}

func (p pmsTokenSource) Token(ctx context.Context) (Token, error) {
	// get a token to access the plex.tv API
	token, err := p.tokenSource.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	p.logger.Debug("got plex.tv token", "jwt", token.IsJWT())
	var mediaServers []RegisteredDevice
	if mediaServers, err = p.config.MediaServers(ctx, token); err == nil {
		for _, server := range mediaServers {
			p.logger.Debug("media server found", "name", server.Name)
			if server.Name == p.pmsName || p.pmsName == "" {
				p.logger.Debug("got media server token")
				return Token(server.Token), nil
			}
		}
		err = fmt.Errorf("media server %q not found", p.pmsName)
	}
	p.logger.Debug("no media server found", "err", err)
	return "", fmt.Errorf("media servers: %w", err)
}
