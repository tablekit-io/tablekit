package oauth

import (
	"time"

	"core/services/store"

	"github.com/google/uuid"
)

// MintedBearer is the result of minting a long-lived bearer token: the prefixed
// token value to hand out, its id (needed to revoke it later), and its expiry.
type MintedBearer struct {
	Token     string
	ID        string
	ExpiresAt time.Time
}

// MintBearer registers a bearer-only client, issues its long-lived JWT, and
// records the token row. A bearer token registers as its own client: no redirect
// URIs, no name. The returned Token is already prefixed (TokenPrefix) and ready
// to hand out.
func MintBearer(st *store.Store, iss *Issuer) (MintedBearer, error) {
	clientID := uuid.NewString()
	tokenID := uuid.NewString()
	now := time.Now()

	if err := st.SaveClient(&store.Client{
		ClientID:     clientID,
		ClientName:   nil,
		RedirectURIs: []string{},
		Type:         "bearer",
		CreatedAt:    now,
	}); err != nil {
		return MintedBearer{}, err
	}

	token, expiresAt, err := iss.IssueBearer(clientID, tokenID)
	if err != nil {
		return MintedBearer{}, err
	}
	if err := st.PutBearerToken(&store.BearerToken{
		ID:        tokenID,
		ClientID:  clientID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}); err != nil {
		return MintedBearer{}, err
	}

	return MintedBearer{Token: TokenPrefix + token, ID: tokenID, ExpiresAt: expiresAt}, nil
}
