package plexauth

import (
	"bytes"
	"cmp"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

var (
	// DefaultConfig contains the default configuration required to authenticate with Plex.
	DefaultConfig = Config{
		URL:      "https://plex.tv",
		V2URL:    "https://clients.plex.tv", // TODO: do any endpoints mandate clients.plex.tv?
		Scopes:   []string{"username", "email", "friendly_name", "restricted", "anonymous"},
		ClientID: uuid.New().String(),
		aud:      "plex.tv",
		tokenTTL: 7 * 24 * time.Hour,
	}

	defaultHTTPClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: http.DefaultTransport,
	}
)

type httpClientType struct{}

// ContextWithHTTPClient returns a new context with an added HTTP client. When passed to [Config]'s methods,
// they use that HTTP client to perform their authentication calls.
// If no HTTP client is set, a default HTTP client is used.
func ContextWithHTTPClient(ctx context.Context, httpClient *http.Client) context.Context {
	return context.WithValue(ctx, httpClientType{}, httpClient)
}

// httpClient returns the HTTP set in the context. If none is set, it returns a default client.
func httpClient(ctx context.Context) *http.Client {
	if c, ok := ctx.Value(httpClientType{}).(*http.Client); ok {
		return c
	}
	return defaultHTTPClient
}

// Device identifies the client when using Plex username/password credentials.
// Although this package provides a default, it is recommended to set this yourself.
//
// Limitation: currently the Device attributes are only registered when the device is registered
// (during [Config.RegisterWithCredentials] or [Config.RegisterWithPIN]).  If the attributes change
// after registration (e.g., the Version is updated), this is not (yet) reflected in the registered
// device on plex.tv.
type Device struct {
	// Product is the name of the client product.
	// Passed as X-Plex-Product header.
	// In Authorized Devices, it is shown on line 3.
	Product string
	// Version is the version of the client application.
	// Passed as X-Plex-Version header.
	// In Authorized Devices, it is shown on line 2.
	Version string
	// Platform is the operating system or compiler of the client application.
	// Passed as X-Plex-Platform header.
	Platform string
	// PlatformVersion is the version of the platform.
	// Passed as X-Plex-Platform-Version header.
	PlatformVersion string
	// Device is a relatively friendly name for the client device.
	// Passed as X-Plex-Device header.
	// In Authorized Devices, it is shown on line 4.
	Device string
	// Model is a potentially less friendly identifier for the device model.
	// Passed as X-Plex-Model header.
	Model string
	// DeviceVendor is the name of the device vendor.
	// Passed as X-Plex-Device-Vendor header.
	DeviceVendor string
	// DeviceName is a friendly name for the client.
	// Passed as X-Plex-Device-Name header.
	// In Authorized Devices, it is shown on line 1.
	DeviceName string
	// Provides describes the type of device.
	// Passed as X-Plex-Provides header.
	Provides string
}

// populateRequest populates the request headers with the device information.
func (id Device) populateRequest(req *http.Request) {
	for key, value := range id.values() {
		if value != "" {
			req.Header.Set(key, value)
		}
	}
}

// Query returns the device information as url.Values, suitable for use in a URL query.
//
// TODO: exported for now for debugging purpose, but will be unexported for GA.
func (id Device) Query() url.Values {
	v := make(url.Values)
	for key, value := range id.values() {
		v.Add(key, value)
	}
	return v
}

func (id Device) values() iter.Seq2[string, string] {
	headers := map[string]string{
		"X-Plex-Product":          id.Product,
		"X-Plex-Version":          id.Version,
		"X-Plex-Platform":         id.Platform,
		"X-Plex-Platform-Version": id.PlatformVersion,
		"X-Plex-Device":           id.Device,
		"X-Plex-Device-Vendor":    id.DeviceVendor,
		"X-Plex-Device-Name":      id.DeviceName,
		"X-Plex-Model":            id.Model,
		"X-Plex-Provides":         id.Provides,
	}
	return func(yield func(string, string) bool) {
		for key, value := range headers {
			if value != "" {
				if !yield(key, value) {
					return
				}
			}
		}
	}
}

// Config contains the configuration required to authenticate with Plex.
type Config struct {
	// Device information used during username/password authentication.
	Device Device
	// URL is the base URL of the legacy Plex authentication endpoint.
	// Defaults to https://plex.tv and should not need to be changed.
	URL string
	// V2URL is the base URL of the new Plex authentication endpoint.
	// Defaults to https://clients.plex.tv and should not need to be changed.
	V2URL string
	// ClientID is the unique identifier of the client application.
	ClientID string
	aud      string
	// Scopes is a list of scopes to request.
	// This may become non-exported in the future.
	Scopes []string
	// TokenTTL is the duration of the authentication token.
	// Defaults to 7 days, in line with Plex specifications.
	// Normally, this should not need to be changed.
	tokenTTL time.Duration
}

// WithClientID sets the Client ID.
func (c Config) WithClientID(clientID string) Config {
	c.ClientID = clientID
	return c
}

// WithDevice sets the device information used during username/password and pin authentication.
//
// See the [Device] type for details on what each field means.
//
// Limitation: currently the Device attributes are only registered when the device is registered
// // (during [Config.RegisterWithCredentials] or [Config.RegisterWithPIN]).  If the attributes change
// // after registration (e.g., the Version is updated), this is not (yet) reflected in the registered
// // device on plex.tv.
func (c Config) WithDevice(device Device) Config {
	c.Device = device
	return c
}

// RegisterWithCredentials registers a device using username/password credentials and returns a Token.
func (c Config) RegisterWithCredentials(ctx context.Context, username, password string) (Token, error) {
	// credentials are passed in the request body in url-encoded form
	v := make(url.Values)
	v.Set("user[login]", username)
	v.Set("user[password]", password)

	// call the auth endpoint
	resp, err := c.do(ctx, http.MethodPost, c.URL+"/users/sign_in.xml", strings.NewReader(v.Encode()), http.StatusCreated, func(req *http.Request) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/xml")
		c.Device.populateRequest(req)
	})
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// a successful response contains an XML document with an authentication token
	var authResponse struct {
		XMLName             xml.Name `xml:"user"`
		AuthenticationToken string   `xml:"authenticationToken,attr"`
	}
	err = xml.NewDecoder(resp.Body).Decode(&authResponse)
	return Token(authResponse.AuthenticationToken), err
}

// RegisterWithPIN is a helper function that registers a device using the PIN authentication flow and gets a Token.
// It requests a PIN from Plex, calls the callback with the PINResponse and PIN URL and blocks until the PIN is confirmed.
// Use a context with a timeout to ensure it doesn't block forever.
//
// The callback can be used to inform the user/application of the URL to confirm the PINRequest.
func (c Config) RegisterWithPIN(ctx context.Context, callback func(PINResponse, string), pollInterval time.Duration) (token Token, err error) {
	pinResponse, pinURL, err := c.PINRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("pin: %w", err)
	}
	callback(pinResponse, pinURL)
	for {
		if token, _, err = c.ValidatePIN(ctx, pinResponse.Id); err == nil && token.IsValid() {
			return token, nil
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(cmp.Or(pollInterval, 15*time.Second)):
		}
	}
}

// PINRequest requests a PINRequest from Plex.
//
// Currently only supports strong=false. Support for strong=true is planned, but this requires https://app.plex.tv/auth,
// which is currently broken.
func (c Config) PINRequest(ctx context.Context) (PINResponse, string, error) {
	resp, err := c.do(ctx, http.MethodPost, c.V2URL+"/api/v2/pins" /*?strong=false"*/, nil, http.StatusCreated, func(req *http.Request) {
		c.Device.populateRequest(req)
	})
	if err != nil {
		return PINResponse{}, "", fmt.Errorf("pin request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var response PINResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return PINResponse{}, "", fmt.Errorf("decode: %w", err)
	}
	// legacy endpoint. once https://app.plex.tv/auth is fixed, this can be adapted accordingly.
	return response, "https://plex.tv/pin?pin=" + response.Code, nil
}

// ValidatePIN checks if the user has confirmed the PINRequest.  It returns the full Plex response.
// When the user has confirmed the PINRequest, the AuthToken field will be populated.
func (c Config) ValidatePIN(ctx context.Context, id int) (Token, ValidatePINResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, c.V2URL+"/api/v2/pins/"+strconv.Itoa(id), nil, http.StatusOK, func(req *http.Request) {
		// this is only needed once we start using the new response url (https://app.plex.tv/auth),
		// but leaving it here for now, as it doesn't do any harm.
		c.Device.populateRequest(req)
	})
	if err != nil {
		return "", ValidatePINResponse{}, fmt.Errorf("validate pin: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var response ValidatePINResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", ValidatePINResponse{}, fmt.Errorf("decode: %w", err)
	}
	var token Token
	if response.AuthToken != nil {
		token = Token(*response.AuthToken)
	}
	return token, response, err
}

// GenerateAndUploadPublicKey is a helper function to set up JWT Tokens.
// It generates a new ed25519 keypair, uploads the private key to the Plex server and
// returns the private key and associated public key ID to be used for generating a new JWT token.
//
// Token must be a valid Plex token, either generated by [Config.RegisterWithCredentials]/[Config.RegisterWithPIN] or
// obtained from a previous [Config.JWTToken] call.
func (c Config) GenerateAndUploadPublicKey(ctx context.Context, token Token) (ed25519.PrivateKey, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generate keypair: %w", err)
	}
	keyID, err := c.UploadPublicKey(ctx, publicKey, token)
	return privateKey, keyID, err
}

// UploadPublicKey uploads a public key to the Plex server. It returns a generated key ID for the public key,
// which can be used to generate a new token with [Config.JWTToken].
func (c Config) UploadPublicKey(ctx context.Context, publicKey ed25519.PublicKey, token Token) (string, error) {
	// check we have a valid token
	// TODO: IsValid should return err, not bool, so the caller can determine the root cause of the error.
	if !token.IsValid() {
		return "", ErrInvalidToken
	}

	// create a jwk from the public key
	jwKey, err := jwk.Import(publicKey)
	if err != nil {
		return "", fmt.Errorf("import key: %w", err)
	}
	// Assign a key ID (kid) using thumbprint
	if err = jwk.AssignKeyID(jwKey); err != nil {
		return "", fmt.Errorf("assign key id: %w", err)
	}
	keyID, _ := jwKey.KeyID()
	// Set use (sig) and algorithm
	_ = jwKey.Set(jwk.KeyUsageKey, "sig")
	_ = jwKey.Set(jwk.KeyIDKey, keyID)
	_ = jwKey.Set(jwk.AlgorithmKey, jwa.EdDSA().String())

	// Marshal to JSON
	jwkBody, _ := json.Marshal(map[string]any{"jwk": jwKey})

	// upload the key to the Plex server
	resp, err := c.do(ctx, http.MethodPost, c.V2URL+"/api/v2/auth/jwk", bytes.NewReader(jwkBody), http.StatusCreated, func(req *http.Request) {
		req.Header.Set("X-Plex-Token", token.String())
	})
	if err != nil {
		return "", fmt.Errorf("jwk: %w", err)
	}
	_ = resp.Body.Close()
	return keyID, nil
}

// JWTToken is a new authentication mechanism introduced in plex.tv, based on JSON Web Tokens (JWT).
//
// JWTTokens increase security by eliminating the need for a PIN or username/password combination each time
// a client starts. As setting up JWT requires a valid token, a client only needs to register once.
//
// To create a JWTToken, a client generates a new ed25519 keypair and uploads the public key to Plex
// (using [Config.GenerateAndUploadPublicKey] or [Config.UploadPublicKey], using a valid Plex token).
// The client can then use the private key and the public key's ID to generate a new JWTToken.
//
// This does require persistence, as the Client ID, private key, and public key ID must be kept in sync
// with plex.tv: once a JWTToken has been requested for the ClientID, further requests to re-register that ClientID
// ([Config.RegisterWithCredentials]/[Config.RegisterWithPIN]) will fail. You will need to generate a new ClientID
// and re-register.
//
// JWTTokens are valid for 7 days.
//
// Note: a JWTToken can only be used to access the plex.tv API; it cannot be used to access Plex Media Servers.
func (c Config) JWTToken(ctx context.Context, privateKey ed25519.PrivateKey, keyID string) (Token, error) {
	nonce, err := c.nonce(ctx)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}
	// create a jwt
	tok := jwt.New()
	_ = tok.Set("nonce", nonce)
	_ = tok.Set("scope", strings.Join(c.Scopes, ","))
	_ = tok.Set("aud", c.aud)
	_ = tok.Set("iss", c.ClientID)
	headers := jws.NewHeaders()
	_ = headers.Set(jws.KeyIDKey, keyID)
	signed, err := jwt.Sign(tok,
		jwt.WithKey(
			jwa.EdDSA(),
			privateKey,
			jws.WithProtectedHeaders(headers),
		),
	)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	// request a new jwtToken
	return c.jwtToken(ctx, string(signed))
}

func (c Config) nonce(ctx context.Context) (string, error) {
	resp, err := c.do(ctx, http.MethodGet, c.V2URL+"/api/v2/auth/nonce", nil, http.StatusOK)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var response struct {
		Nonce string `json:"nonce"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return response.Nonce, nil
}

func (c Config) jwtToken(ctx context.Context, signedJWToken string) (Token, error) {
	// send the signed token to the auth endpoint
	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(map[string]string{"jwt": signedJWToken})
	resp, err := c.do(ctx, http.MethodPost, c.V2URL+"/api/v2/auth/token", &body, http.StatusOK)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// extract the new token from the response
	var response struct {
		AuthToken string `json:"auth_token"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return Token(response.AuthToken), nil
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

// PlexTVClient returns a [PlexTVClient] that can be used to query the plex.tv API.
func (c Config) PlexTVClient(opts ...TokenSourceOption) PlexTVClient {
	return PlexTVClient{
		config:      &c,
		tokenSource: c.TokenSource(opts...),
	}
}

// requestFormatter modifies a request before [Config.do] sends to its destination.
type requestFormatter func(*http.Request)

// do builds a new HTTP request and sends it to the destination URL.
// note: url cannot include query parameters
func (c Config) do(ctx context.Context, method string, url string, body io.Reader, wantStatusCode int, formatters ...requestFormatter) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("X-Plex-Client-Identifier", c.ClientID)
	c.Device.populateRequest(req)
	for _, formatter := range formatters {
		formatter(req)
	}
	resp, err := httpClient(ctx).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != wantStatusCode {
		defer func() { _ = resp.Body.Close() }()
		return nil, ParsePlexError(resp)
	}
	return resp, nil
}
