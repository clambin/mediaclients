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

// A AuthTokenSource returns a Plex authentication Token
type AuthTokenSource interface {
	Token(ctx context.Context) (AuthToken, error)
}

var (
	_ AuthTokenSource = fixedTokenSource{}
	_ AuthTokenSource = (*cachingTokenSource)(nil)
	_ AuthTokenSource = (*legacyTokenSource)(nil)
	_ AuthTokenSource = (*pmsTokenSource)(nil)
)

// TokenSourceFactory provides methods to create AuthTokenSources.
// This struct should not be created directly but rather by calling [Config.TokenSource].
type TokenSourceFactory struct {
	config *Config
}

// FixedToken returns an AuthTokenSource that always returns the given token.
func (f TokenSourceFactory) FixedToken(token AuthToken) AuthTokenSource {
	return &cachingTokenSource{AuthTokenSource: fixedTokenSource{token: token}}
}

// LegacyToken returns an AuthTokenSource that uses the given Registrar to register a new device and get an auth token.
func (f TokenSourceFactory) LegacyToken(r Registrar) AuthTokenSource {
	return &cachingTokenSource{
		AuthTokenSource: &legacyTokenSource{Registrar: r},
	}
}

// PMSToken returns an AuthTokenSource that returns the auth Token for the given Plex Media Server.
// It uses Plex's legacy authentication flows to register the device and get an authToken, which is then used
// to retrieve the Plex Media Server token.
func (f TokenSourceFactory) PMSToken(r Registrar, pmsName string) AuthTokenSource {
	return &cachingTokenSource{
		AuthTokenSource: &pmsTokenSource{
			tokenSource:     &registrarAsTokenSource{Registrar: r},
			Config:          f.config,
			MediaServerName: pmsName,
		},
	}
}

// PMSTokenWithJWT returns an AuthTokenSource that returns the auth Token for the given Plex Media Server.
// It uses Plex's new authentication mechanism, based on JSON Web Tokens (JWT).
// This means the device is only registered once; later requests for an auth Token use a temporary JWT token,
// using a public key uploaded to Plex Cloud when the device was registered.
//
// This requires storage of the associated private data in the given storePath.
// The data is encrypted using the passphrase.
func (f TokenSourceFactory) PMSTokenWithJWT(r Registrar, pmsName string, storePath string, passphrase string, logger *slog.Logger) AuthTokenSource {
	return &cachingTokenSource{
		AuthTokenSource: &pmsTokenSource{
			tokenSource: &jwtTokenSource{
				Config:    f.config,
				Vault:     newJWTDataStore(storePath, passphrase, f.config.ClientID),
				Registrar: r,
				Logger:    logger,
			},
			Config:          f.config,
			MediaServerName: pmsName,
		},
	}
}

// A cachingTokenSource caches the auth token obtained by the underlying AuthTokenSource.
type cachingTokenSource struct {
	AuthTokenSource
	authToken AuthToken
	lock      sync.Mutex
}

func (s *cachingTokenSource) Token(ctx context.Context) (AuthToken, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var err error
	if !s.authToken.IsValid() {
		s.authToken, err = s.AuthTokenSource.Token(ctx)
	}
	return s.authToken, err
}

// fixedTokenSource returns a fixed token.
type fixedTokenSource struct {
	token AuthToken
}

func (f fixedTokenSource) Token(_ context.Context) (AuthToken, error) {
	return f.token, nil
}

// A legacyTokenSource uses a Registrar to register a new device and get an auth token.
type legacyTokenSource struct {
	Registrar
}

func (s *legacyTokenSource) Token(ctx context.Context) (AuthToken, error) {
	return s.Register(ctx)
}

// A pmsTokenSource returns the Plex authentication token for a given Plex Media Server.
type pmsTokenSource struct {
	tokenSource
	Config          *Config
	MediaServerName string
}

func (p pmsTokenSource) Token(ctx context.Context) (AuthToken, error) {
	// get a token to access the Plex Cloud API
	token, err := p.token(ctx)
	if err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	mediaServers, err := p.Config.MediaServers(ctx, token)
	if err != nil {
		return "", fmt.Errorf("media servers: %w", err)
	}
	for _, server := range mediaServers {
		if server.Name == p.MediaServerName || p.MediaServerName == "" {
			return AuthToken(server.Token), nil
		}
	}
	return "", fmt.Errorf("no media server %q found", p.MediaServerName)
}

// tokenSource is a common interface for the underlying token source for a pmsTokenSource.
type tokenSource interface {
	token(ctx context.Context) (Token, error)
}

var (
	_ tokenSource = (*jwtTokenSource)(nil)
	_ tokenSource = (*registrarAsTokenSource)(nil)
)

// A jwtTokenSource returns a Plex JWT Token. If needed, it registers a new device using the configured Registrar.
type jwtTokenSource struct {
	Registrar  Registrar
	Vault      secureDataVault
	Logger     *slog.Logger
	Config     *Config
	secureData jwtSecureData
	init       sync.Once
}

type secureDataVault interface {
	Load() (jwtSecureData, error)
	Save(jwtSecureData) error
}

func (s *jwtTokenSource) initialize(ctx context.Context) (err error) {
	// set up logger if not done already
	if s.Logger == nil {
		s.Logger = slog.New(slog.DiscardHandler)
	}

	// load the client's jwt token data
	s.secureData, err = s.Vault.Load()
	switch {
	case err == nil:
		// valid secure data found
		return nil
	case errors.Is(err, ErrInvalidClientID):
		// secure data found, but not for this client ID
		s.Logger.Warn("client ID mismatch, secure data found but not for this client. Overwriting secure data",
			slog.String("want", s.Config.ClientID),
			slog.String("found", s.secureData.ClientID),
		)
	case errors.Is(err, os.ErrNotExist):
		s.Logger.Info("no secure data found. Initializing")
	default:
		return fmt.Errorf("load token data: %w", err)
	}

	s.Logger.Debug("registering device")

	var authToken AuthToken
	authToken, err = s.Registrar.Register(ctx)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	s.Logger.Debug("device registered successfully")

	s.secureData.ClientID = s.Config.ClientID
	if s.secureData.PrivateKey, s.secureData.KeyID, err = s.Config.GenerateAndUploadPublicKey(ctx, authToken); err != nil {
		return fmt.Errorf("publish key: %w", err)
	}

	s.Logger.Debug("public key published successfully")

	if err = s.Vault.Save(s.secureData); err != nil {
		return fmt.Errorf("save token data: %w", err)
	}

	s.Logger.Debug("token data saved successfully")
	return nil
}

func (s *jwtTokenSource) token(ctx context.Context) (token Token, err error) {
	// load the client's jwt token data. We only need to do this once.
	if s.init.Do(func() { err = s.initialize(ctx) }); err != nil {
		return token, fmt.Errorf("init: %w", err)
	}
	// create a jwt.
	return s.Config.JWTToken(ctx, s.secureData.PrivateKey, s.secureData.KeyID)
}

// registrarAsTokenSource is an adapter that allows the use of any Registrar as an tokenSource..
type registrarAsTokenSource struct {
	Registrar
}

func (r registrarAsTokenSource) token(ctx context.Context) (Token, error) {
	return r.Register(ctx)
}

// A Registrar registers a new device with Plex Cloud and returns an auth token for
// the provided credentials & Client Identifier.
type Registrar interface {
	Register(context.Context) (AuthToken, error)
}

var (
	_ Registrar = CredentialsRegistrar{}
	_ Registrar = PINRegistrar{}
)

type CredentialsRegistrar struct {
	Config   *Config
	Username string
	Password string
}

func (r CredentialsRegistrar) Register(ctx context.Context) (AuthToken, error) {
	return r.Config.RegisterWithCredentials(ctx, r.Username, r.Password)
}

type PINRegistrar struct {
	Callback     func(PINResponse, string)
	Config       *Config
	PollInterval time.Duration
}

func (r PINRegistrar) Register(ctx context.Context) (AuthToken, error) {
	return r.Config.RegisterWithPIN(ctx, r.Callback, r.PollInterval)
}
