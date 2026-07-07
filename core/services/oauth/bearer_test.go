package oauth

import (
	"context"
	"strings"
	"testing"

	"core/services/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMintBearer(t *testing.T) {
	database := openTestDB(t)
	clients := store.NewClientRepository(database)
	tokens := store.NewBearerTokenRepository(database)
	issuer := newIssuer(t, testConfig())
	ctx := context.Background()

	minted, err := MintBearer(ctx, clients, tokens, issuer)
	require.NoError(t, err)

	// The handed-out token is prefixed and carries the returned id/expiry.
	assert.True(t, strings.HasPrefix(minted.Token, TokenPrefix), "token must carry the bearer prefix")
	assert.NotEmpty(t, minted.ID)
	assert.False(t, minted.ExpiresAt.IsZero())

	// The token row is persisted, retrievable by id, and not revoked.
	row, err := tokens.GetBearerToken(ctx, minted.ID)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.False(t, row.Revoked)
	assert.Equal(t, minted.ID, row.ID)

	// The raw JWT (prefix stripped) verifies as a bearer token with that jti.
	rawJWT := strings.TrimPrefix(minted.Token, TokenPrefix)
	claims, err := issuer.VerifyBearer(rawJWT)
	require.NoError(t, err)
	assert.Equal(t, minted.ID.String(), claims.ID)
}
