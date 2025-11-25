package plex

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
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

const legacyAuthURL = "https://plex.tv/users/sign_in.xml"
const tokenLifetime = 7 * 24 * time.Hour
const baseURLV2 = "https://clients.plex.tv"

func AuthorizeDevice(ctx context.Context, httpClient *http.Client, username, password, clientID string, identity Device) (string, error) {
}

type tokenSource interface {
	Token(context.Context) (string, error)
}

var (
	_ tokenSource = (*fixedTokenSource)(nil)
	_ tokenSource = (*refreshingTokenSource)(nil)
)

// fixedTokenSource returns a tokenSource that always returns the same token.
type fixedTokenSource struct {
	token string
}

func (s *fixedTokenSource) Token(_ context.Context) (string, error) {
	return s.token, nil
}

type refreshingTokenSource struct {
	expiration time.Time
	httpClient *http.Client
	lock       sync.Mutex
	privateKey ed25519.PrivateKey
	clientID   string
	token      string
	kid        string
	baseURL    string
	uploadKeys sync.Once
}

func (s *refreshingTokenSource) Token(ctx context.Context) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// use the cached token if it's available and not expired yet
	if s.token != "" && time.Now().Before(s.expiration) {
		return s.token, nil
	}

	// generate keys & upload to plex
	var err error
	s.uploadKeys.Do(func() {
		err = s.generateKeys(ctx)
	})
	if err != nil {
		return "", fmt.Errorf("keys: %w", err)
	}

	// upload the token if necessary
	nonce, err := s.getNonce(ctx)
	if err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	// sign a jwt and send it to plex. response is the new token
	if s.token, err = s.getToken(ctx, nonce); err == nil {
		s.expiration = time.Now().Add(tokenLifetime)
	}
	return s.token, err
}

func (s *refreshingTokenSource) generateKeys(ctx context.Context) error {
	// generate keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}
	s.privateKey = priv

	// create a jwk from the public key
	key, err := jwk.Import(pub)
	if err != nil {
		return fmt.Errorf("import key: %w", err)
	}

	// Assign a key ID (kid) using thumbprint
	if err = jwk.AssignKeyID(key); err != nil {
		return fmt.Errorf("assign key id: %w", err)
	}
	var ok bool
	if s.kid, ok = key.KeyID(); !ok {
		panic("key id not set")
	}

	// Set use (sig) and algorithm
	_ = key.Set(jwk.KeyUsageKey, "sig")
	_ = key.Set(jwk.KeyIDKey, s.kid)
	_ = key.Set(jwk.AlgorithmKey, jwa.EdDSA().String())

	// Marshal to JSON
	jwkBody, err := json.MarshalIndent(map[string]any{"jwk": key}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cmp.Or(s.baseURL, baseURLV2)+"/api/v2/auth/jwk", bytes.NewReader(jwkBody))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", s.clientID)
	req.Header.Set("X-Plex-Token", s.token)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d - %s (%s)", resp.StatusCode, resp.Status, errBody)
	}
	return nil
}

func (s *refreshingTokenSource) getNonce(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cmp.Or(s.baseURL, baseURLV2)+"/api/v2/auth/nonce", nil)
	req.Header.Set("X-Plex-Client-Identifier", s.clientID)
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, resp.Status)
	}

	var response struct {
		Nonce string `json:"nonce"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response.Nonce, nil
}

func (s *refreshingTokenSource) getToken(ctx context.Context, nonce string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"nonce": nonce,
		"scope": "username,email,friendly_name",
		"aud":   "plex.tv",
		"iss":   s.clientID,
		"iat":   now.Unix(),
		"exp":   now.Add(tokenLifetime).Unix(),
	})
	token.Header["kid"] = s.kid

	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	var body bytes.Buffer
	if err = json.NewEncoder(&body).Encode(map[string]string{"jwt": signedToken}); err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cmp.Or(s.baseURL, baseURLV2)+"/api/v2/auth/token", &body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", s.clientID)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("post: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("post: unexpected status code: %d - %s (%s)", resp.StatusCode, resp.Status, string(b))
	}

	var response struct {
		AuthToken string `json:"auth_token"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	return response.AuthToken, nil
}
