// Package oauth implements a minimal single-client OAuth 2.1 authorization
// server (DCR + PKCE + JWT access/refresh tokens with rotation/replay
// detection), modeled on the dbctx reference but backed by JSON files.
package oauth

import (
	"fmt"
	"time"

	"core/config"
	"core/services"
	"core/store"

	"github.com/golang-jwt/jwt/v5"
)

// Token audiences distinguish access from refresh tokens so a refresh token
// can never be presented as an access token (and vice versa).
const (
	audienceAccess  = "mcp"
	audienceRefresh = "mcp-refresh"
	// Subject is fixed: this is a single-user server.
	subject = "user:owner"
	// UserID is the bare identifier carried into the MCP session.
	UserID = "owner"
	// Scope is the only scope this server issues.
	Scope = "mcp"
)

// Claims is the JWT payload. cid = client row id, chain = refresh chain id.
type Claims struct {
	CID   string `json:"cid"`
	Chain string `json:"chain"`
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

// Issuer builds and validates the server's JWTs against the shared HS256 key.
type Issuer struct {
	cfg *config.Config
	key []byte
}

func init() {
	// Refresh rotation bumps a chain's cutoff to the used token's iat. At the
	// default second precision two refreshes within the same second collide and
	// the second would look like a replay. Microsecond precision removes that.
	jwt.TimePrecision = time.Microsecond
}

// NewIssuer resolves the signing key: an externally provided base64 key
// (config.SigningKey) wins, otherwise the store's generated/persisted key is used.
func NewIssuer(appServices *services.Services) (*Issuer, error) {
	configService := appServices.Config
	var key []byte
	var err error
	if configService.SigningKey != "" {
		key, err = store.DecodeSigningKey(configService.SigningKey)
	} else {
		key, err = appServices.Store.SigningKey()
	}
	if err != nil {
		return nil, err
	}
	return &Issuer{cfg: configService, key: key}, nil
}

type issueArgs struct {
	clientID string
	chainID  string
	scope    string
}

func (i *Issuer) sign(a issueArgs, aud string, ttl time.Duration) (token string, iat time.Time, err error) {
	now := time.Now()
	claims := Claims{
		CID:   a.clientID,
		Chain: a.chainID,
		Scope: a.scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.cfg.PublicBaseURL,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{aud},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(i.key)
	return signed, now, err
}

// IssueAccess mints a short-lived access token.
func (i *Issuer) IssueAccess(clientID, chainID, scope string) (string, error) {
	t, _, err := i.sign(issueArgs{clientID, chainID, scope}, audienceAccess, i.cfg.AccessTTL)
	return t, err
}

// IssueRefresh mints a refresh token and returns its iat (needed for the chain
// rotation cutoff).
func (i *Issuer) IssueRefresh(clientID, chainID, scope string) (token string, iat time.Time, err error) {
	return i.sign(issueArgs{clientID, chainID, scope}, audienceRefresh, i.cfg.RefreshTTL)
}

func (i *Issuer) verify(token, aud string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return i.key, nil
	},
		jwt.WithIssuer(i.cfg.PublicBaseURL),
		jwt.WithAudience(aud),
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// VerifyAccess validates an access token and returns its claims.
func (i *Issuer) VerifyAccess(token string) (*Claims, error) {
	return i.verify(token, audienceAccess)
}

// VerifyRefresh validates a refresh token and returns its claims.
func (i *Issuer) VerifyRefresh(token string) (*Claims, error) {
	return i.verify(token, audienceRefresh)
}
