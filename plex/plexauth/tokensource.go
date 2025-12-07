package plexauth

import (
	"context"
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

// NewFixedTokenSource returns an AuthTokenSource that always returns the given token.
func NewFixedTokenSource(token AuthToken) AuthTokenSource {
	return fixedTokenSource{token: token}
}

type fixedTokenSource struct {
	token AuthToken
}

func (f fixedTokenSource) Token(_ context.Context) (AuthToken, error) {
	return f.token, nil
}

// NewLegacyTokenSource returns an AuthTokenSource that returns the Plex Auth Token using the given Registrar.
func NewLegacyTokenSource(registrar Registrar) AuthTokenSource {
	return &cachingTokenSource{
		AuthTokenSource: &legacyTokenSource{Registrar: registrar},
	}
}

// NewPMSTokenStore returns an AuthTokenSource that returns the Plex Auth Token for the given media server.
// It uses Plex's legacy authentication flows to register the device and get an authToken, which is then used
// to retrieve the Plex Media Server token. This means the device is registered on each first call.
func NewPMSTokenStore(cfg Config, registrar Registrar, mediaServerName string) AuthTokenSource {
	return &cachingTokenSource{
		AuthTokenSource: &pmsTokenSource{
			tokenSource:     &registrarAsTokenSource{Registrar: registrar},
			Config:          &cfg,
			MediaServerName: mediaServerName,
		},
	}
}

// NewPMSTokenSourceWithJWT returns an AuthTokenSource that returns the Plex Auth Token for the given media server.
// It uses Plex's new authentication mechanism, based on JSON Web Tokens (JWT). This means the device is only registered once;
// later calls use a temporary JWT token using a public key uploaded to Plex Cloud when the device was registered.
//
// This requires storage of the associated private data in the given storePath. The data is encrypted using the passphrase.
func NewPMSTokenSourceWithJWT(cfg Config, registrar Registrar, mediaServerName string, storePath string, passphrase string, logger *slog.Logger) AuthTokenSource {
	return &cachingTokenSource{
		AuthTokenSource: &pmsTokenSource{
			tokenSource: &jwtTokenSource{
				Config:    &cfg,
				Store:     NewJWTDataStore(storePath, passphrase),
				Registrar: registrar,
				Logger:    logger,
			},
			Config:          &cfg,
			MediaServerName: mediaServerName,
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
	Store      *JWTDataStore
	Logger     *slog.Logger
	Config     *Config
	secureData JWTSecureData
	init       sync.Once
}

func (s *jwtTokenSource) token(ctx context.Context) (token Token, err error) {
	// load the client's jwt token data. We only need to do this once.
	if s.init.Do(func() { err = s.initialize(ctx) }); err != nil {
		return token, fmt.Errorf("failed to initialize: %w", err)
	}
	// create a jwt.
	return s.Config.JWTToken(ctx, s.secureData.PrivateKey, s.secureData.KeyID)
}

func (s *jwtTokenSource) initialize(ctx context.Context) (err error) {
	// set up logger if not done already
	if s.Logger == nil {
		s.Logger = slog.New(slog.DiscardHandler)
	}

	// load the client's jwt token data
	s.secureData, err = s.Store.Get()
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to load token data: %w", err)
	}

	s.Logger.Debug("token data not found, creating new one")

	var authToken AuthToken
	authToken, err = s.Registrar.Register(ctx)
	if err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}

	s.Logger.Debug("device registered successfully")

	if s.secureData.PrivateKey, s.secureData.KeyID, err = s.Config.GenerateAndUploadPublicKey(ctx, authToken); err != nil {
		return fmt.Errorf("failed to publish key: %w", err)
	}

	s.Logger.Debug("public key published successfully")

	if err = s.Store.Set(s.secureData); err != nil {
		return fmt.Errorf("failed to save token data: %w", err)
	}

	s.Logger.Debug("token data saved successfully")
	return nil
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
