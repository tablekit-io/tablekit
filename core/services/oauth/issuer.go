// Package oauth implements a minimal single-client OAuth 2.1 authorization
// server (DCR + PKCE + JWT access/refresh tokens with rotation/replay
// detection), modeled on the dbctx reference but backed by JSON files.
package oauth

import (
	"fmt"
	"time"

	"core/services/config"
	"core/services/store"

	"github.com/golang-jwt/jwt/v5"
)

// Token audiences distinguish access from refresh tokens so a refresh token
// can never be presented as an access token (and vice versa).
const (
	audienceAccess  = "mcp"
	audienceRefresh = "mcp-refresh"
	// audienceBearer is the audience for CLI-minted long-lived bearer tokens. It
	// is deliberately distinct from audienceAccess: a bearer token therefore
	// fails VerifyAccess, so stripping the TokenPrefix and presenting the raw JWT
	// on the OAuth path cannot bypass the revocation check.
	audienceBearer = "mcp-bearer"
	// audienceExport is the audience for short-lived signed export links handed to
	// get_export_url. Distinct from the others so an export token is only ever
	// honoured by the export endpoint, never on the MCP/OAuth paths.
	audienceExport = "mcp-export"
	// Subject is fixed: this is a single-user server.
	subject = "user:owner"
	// UserID is the bare identifier carried into the MCP session.
	UserID = "owner"
	// Scope is the only scope this server issues.
	Scope = "mcp"
)

// TokenPrefix marks a CLI-minted bearer token. It wraps the JWT string
// (Authorization: Bearer <TokenPrefix><jwt>) so the MCP guard can branch on the
// prefix and route the token to the bearer verifier instead of the OAuth one.
const TokenPrefix = "tablekit_pat_"

// bearerMonths is the validity window of a CLI-minted bearer token, in calendar
// months.
const bearerMonths = 6

// exportTTL is the validity window of a signed export link. Short: the link is
// handed straight to the user to click, not stored.
const exportTTL = 5 * time.Minute

// Claims is the JWT payload. cid = client row id, chain = refresh chain id,
// qk = the stored query key an export token authorizes.
type Claims struct {
	CID   string `json:"cid"`
	Chain string `json:"chain"`
	Scope string `json:"scope"`
	QK    string `json:"qk,omitempty"`
	jwt.RegisteredClaims
}

// Issuer builds and validates the server's JWTs against the shared HS256 key.
type Issuer struct {
	configService *config.Config
	key           []byte
}

func init() {
	// Refresh rotation bumps a chain's cutoff to the used token's iat. At the
	// default second precision two refreshes within the same second collide and
	// the second would look like a replay. Microsecond precision removes that.
	jwt.TimePrecision = time.Microsecond
}

// NewIssuer resolves the signing key: an externally provided base64 key
// (config.SigningKey) wins, otherwise the store's generated/persisted key is used.
func NewIssuer(configService *config.Config, storageService *store.Store) (*Issuer, error) {
	var key []byte
	var err error
	if configService.SigningKey != "" {
		key, err = store.DecodeSigningKey(configService.SigningKey)
	} else {
		key, err = storageService.SigningKey()
	}
	if err != nil {
		return nil, err
	}
	return &Issuer{configService: configService, key: key}, nil
}

type issueArgs struct {
	clientID string
	chainID  string
	scope    string
	tokenID  string
	queryKey string
}

func (i *Issuer) sign(a issueArgs, aud string, expiresAt time.Time) (token string, iat time.Time, err error) {
	now := time.Now()
	claims := Claims{
		CID:   a.clientID,
		Chain: a.chainID,
		Scope: a.scope,
		QK:    a.queryKey,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        a.tokenID,
			Issuer:    i.configService.PublicBaseURL,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{aud},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(i.key)
	return signed, now, err
}

// IssueAccess mints a short-lived access token.
func (i *Issuer) IssueAccess(clientID, chainID, scope string) (string, error) {
	args := issueArgs{clientID: clientID, chainID: chainID, scope: scope}
	t, _, err := i.sign(args, audienceAccess, time.Now().Add(i.configService.AccessTTL))
	return t, err
}

// IssueRefresh mints a refresh token and returns its iat (needed for the chain
// rotation cutoff).
func (i *Issuer) IssueRefresh(clientID, chainID, scope string) (token string, iat time.Time, err error) {
	args := issueArgs{clientID: clientID, chainID: chainID, scope: scope}
	return i.sign(args, audienceRefresh, time.Now().Add(i.configService.RefreshTTL))
}

// IssueBearer mints a long-lived bearer token (valid bearerMonths calendar
// months) carrying tokenID as its jti. The returned expiresAt matches the JWT's
// exp exactly, so the caller can persist it on the BearerToken row. The returned
// token is the raw JWT; the caller prepends TokenPrefix before handing it out.
func (i *Issuer) IssueBearer(clientID, tokenID string) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().AddDate(0, bearerMonths, 0)
	args := issueArgs{clientID: clientID, scope: Scope, tokenID: tokenID}
	token, _, err = i.sign(args, audienceBearer, expiresAt)
	return token, expiresAt, err
}

// IssueExport mints a short-lived token authorizing a CSV/JSON export of the
// given stored query. The query key travels in the qk claim under the export
// audience, so the token is useless anywhere but the export endpoint.
func (i *Issuer) IssueExport(queryKey string) (string, error) {
	args := issueArgs{scope: Scope, queryKey: queryKey}
	t, _, err := i.sign(args, audienceExport, time.Now().Add(exportTTL))
	return t, err
}

func (i *Issuer) verify(token, aud string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return i.key, nil
	},
		jwt.WithIssuer(i.configService.PublicBaseURL),
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

// VerifyBearer validates a long-lived bearer token (signature, issuer, audience,
// expiry) and returns its claims. The caller still has to check revocation via
// the token id (claims.ID) in the store.
func (i *Issuer) VerifyBearer(token string) (*Claims, error) {
	return i.verify(token, audienceBearer)
}

// VerifyExport validates a signed export token (signature, issuer, audience,
// expiry) and returns its claims; claims.QK is the stored query key to export.
func (i *Issuer) VerifyExport(token string) (*Claims, error) {
	return i.verify(token, audienceExport)
}

// BearerTokenID extracts the jti from a bearer JWT WITHOUT verifying its
// signature or expiry, so `token:revoke` can resolve an id from a token even if
// it has expired. The id is only used to look up our own row, so trusting it is
// safe — revoking a forged id simply revokes nothing.
func BearerTokenID(token string) (string, error) {
	claims := &Claims{}
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(token, claims); err != nil {
		return "", err
	}
	if claims.ID == "" {
		return "", fmt.Errorf("token has no id")
	}
	return claims.ID, nil
}
