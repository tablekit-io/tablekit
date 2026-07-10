package oauth

import (
	"context"
	"time"

	"core/services/store"

	"github.com/google/uuid"
)

// MintedStatic is the result of minting a long-lived static token: the prefixed
// token value to hand out, its id (needed to revoke it later), and its expiry.
type MintedStatic struct {
	Token     string
	ID        uuid.UUID
	ExpiresAt time.Time
}

// MintStatic registers a static-only client, issues its long-lived JWT, and
// records the token row. A static token registers as its own client: no redirect
// URIs, no name. This is what guarantees every static-token request carries a
// client id. The returned Token is already prefixed (TokenPrefix) and ready to
// hand out.
func MintStatic(ctx context.Context, clients store.ClientRepository, tokens store.StaticTokenRepository, iss *Issuer) (MintedStatic, error) {
	clientID, err := uuid.NewV7()
	if err != nil {
		return MintedStatic{}, err
	}
	tokenID, err := uuid.NewV7()
	if err != nil {
		return MintedStatic{}, err
	}
	now := time.Now()

	if err := clients.SaveClient(ctx, &store.Client{
		ClientID:     clientID,
		ClientName:   nil,
		RedirectURIs: []string{},
		Type:         store.ClientTypeStatic,
		CreatedAt:    now,
	}); err != nil {
		return MintedStatic{}, err
	}

	token, expiresAt, err := iss.IssueStatic(clientID.String(), tokenID.String())
	if err != nil {
		return MintedStatic{}, err
	}
	if err := tokens.PutStaticToken(ctx, &store.StaticToken{
		ID:        tokenID,
		ClientID:  clientID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}); err != nil {
		return MintedStatic{}, err
	}

	return MintedStatic{Token: TokenPrefix + token, ID: tokenID, ExpiresAt: expiresAt}, nil
}
