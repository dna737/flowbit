package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultCacheTTL = 5 * time.Minute
	clockSkew       = 30 * time.Second
)

var (
	ErrMissingToken = errors.New("missing auth token")
	ErrInvalidToken = errors.New("invalid auth token")
)

type Config struct {
	JWKSURL           string
	Issuer            string
	AuthorizedParties []string
}

type Verifier struct {
	jwksURL           string
	issuer            string
	authorizedParties map[string]struct{}
	client            *http.Client
	now               func() time.Time

	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

func NewVerifier(cfg Config) (*Verifier, error) {
	jwksURL := strings.TrimSpace(cfg.JWKSURL)
	if jwksURL == "" {
		return nil, errors.New("clerk jwks url is required")
	}
	parties := make(map[string]struct{}, len(cfg.AuthorizedParties))
	for _, party := range cfg.AuthorizedParties {
		party = strings.TrimSpace(party)
		if party != "" {
			parties[party] = struct{}{}
		}
	}
	return &Verifier{
		jwksURL:           jwksURL,
		issuer:            strings.TrimRight(strings.TrimSpace(cfg.Issuer), "/"),
		authorizedParties: parties,
		client:            &http.Client{Timeout: 10 * time.Second},
		now:               func() time.Time { return time.Now().UTC() },
		keys:              make(map[string]*rsa.PublicKey),
	}, nil
}

func TokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if token := bearerToken(r.Header.Get("Authorization")); token != "" {
		return token
	}
	if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
		return token
	}
	return ""
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func (v *Verifier) Verify(ctx context.Context, token string) (Claims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Claims{}, ErrMissingToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	var header jwtHeader
	if err := decodeSegment(parts[0], &header); err != nil {
		return Claims{}, fmt.Errorf("%w: decode header: %v", ErrInvalidToken, err)
	}
	if header.Algorithm != "RS256" || strings.TrimSpace(header.KeyID) == "" {
		return Claims{}, ErrInvalidToken
	}

	var payload jwtPayload
	if err := decodeSegment(parts[1], &payload); err != nil {
		return Claims{}, fmt.Errorf("%w: decode payload: %v", ErrInvalidToken, err)
	}
	if err := v.validatePayload(payload); err != nil {
		return Claims{}, err
	}

	key, err := v.key(ctx, header.KeyID)
	if err != nil {
		return Claims{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, fmt.Errorf("%w: decode signature: %v", ErrInvalidToken, err)
	}
	signed := []byte(parts[0] + "." + parts[1])
	digest := sha256.Sum256(signed)
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return Claims{}, ErrInvalidToken
	}

	return payload.toClaims(), nil
}

func decodeSegment(segment string, dst any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func (v *Verifier) validatePayload(payload jwtPayload) error {
	sub := strings.TrimSpace(payload.Subject)
	if sub == "" {
		return ErrInvalidToken
	}

	now := v.now()
	if payload.ExpiresAt == 0 || now.After(time.Unix(payload.ExpiresAt, 0).Add(clockSkew)) {
		return fmt.Errorf("%w: expired", ErrInvalidToken)
	}
	if payload.NotBefore != 0 && now.Add(clockSkew).Before(time.Unix(payload.NotBefore, 0)) {
		return fmt.Errorf("%w: not active", ErrInvalidToken)
	}
	if v.issuer != "" && strings.TrimRight(payload.Issuer, "/") != v.issuer {
		return fmt.Errorf("%w: issuer", ErrInvalidToken)
	}
	if len(v.authorizedParties) > 0 && !v.hasAllowedParty(payload) {
		return fmt.Errorf("%w: authorized party", ErrInvalidToken)
	}
	return nil
}

func (v *Verifier) hasAllowedParty(payload jwtPayload) bool {
	if _, ok := v.authorizedParties[payload.AuthorizedParty]; ok {
		return true
	}
	for _, aud := range payload.Audience {
		if _, ok := v.authorizedParties[aud]; ok {
			return true
		}
	}
	return false
}

func (v *Verifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	if key := v.keys[kid]; key != nil && v.now().Before(v.expires) {
		v.mu.Unlock()
		return key, nil
	}
	v.mu.Unlock()

	if err := v.refresh(ctx); err != nil {
		return nil, err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	key := v.keys[kid]
	if key == nil {
		return nil, fmt.Errorf("%w: jwk not found", ErrInvalidToken)
	}
	return key, nil
}

func (v *Verifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch jwks: status %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	var doc jwksDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}
	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, jwk := range doc.Keys {
		key, err := jwk.publicKey()
		if err != nil || jwk.KeyID == "" {
			continue
		}
		keys[jwk.KeyID] = key
	}
	if len(keys) == 0 {
		return fmt.Errorf("%w: no usable jwks", ErrInvalidToken)
	}

	v.mu.Lock()
	v.keys = keys
	v.expires = v.now().Add(defaultCacheTTL)
	v.mu.Unlock()
	return nil
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type jwtPayload struct {
	Subject         string   `json:"sub"`
	Issuer          string   `json:"iss"`
	ExpiresAt       int64    `json:"exp"`
	NotBefore       int64    `json:"nbf"`
	AuthorizedParty string   `json:"azp"`
	Audience        audience `json:"aud"`
	Email           string   `json:"email"`
	EmailAddress    string   `json:"email_address"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	ImageURL        string   `json:"image_url"`
}

func (p jwtPayload) toClaims() Claims {
	email := strings.TrimSpace(p.Email)
	if email == "" {
		email = strings.TrimSpace(p.EmailAddress)
	}
	return Claims{
		Subject:   strings.TrimSpace(p.Subject),
		Email:     email,
		FirstName: strings.TrimSpace(p.FirstName),
		LastName:  strings.TrimSpace(p.LastName),
		ImageURL:  strings.TrimSpace(p.ImageURL),
	}
}

type audience []string

func (a *audience) UnmarshalJSON(raw []byte) error {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		*a = []string{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyID   string `json:"kid"`
	KeyType string `json:"kty"`
	Use     string `json:"use"`
	Alg     string `json:"alg"`
	N       string `json:"n"`
	E       string `json:"e"`
}

func (j jwk) publicKey() (*rsa.PublicKey, error) {
	if j.KeyType != "RSA" || strings.TrimSpace(j.N) == "" || strings.TrimSpace(j.E) == "" {
		return nil, ErrInvalidToken
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, err
	}
	e := new(big.Int).SetBytes(eBytes).Int64()
	if e <= 0 {
		return nil, ErrInvalidToken
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(e),
	}, nil
}
