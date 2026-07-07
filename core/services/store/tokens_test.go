package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthCodeSingleUse(t *testing.T) {
	authCodes := NewAuthCodeRepository(newDB(t))
	ctx := context.Background()
	codeID := uuid.New()
	code := &AuthCode{
		Code: codeID, ClientID: uuid.New(), RedirectURI: "http://x/cb",
		CodeChallenge: "chal", Scope: "mcp", UserID: "owner",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	require.NoError(t, authCodes.PutCode(ctx, code))

	got, err := authCodes.ConsumeCode(ctx, codeID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "chal", got.CodeChallenge)

	// Second consume returns nil — codes are one-time.
	got, err = authCodes.ConsumeCode(ctx, codeID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestChains(t *testing.T) {
	chains := NewTokenChainRepository(newDB(t))
	ctx := context.Background()
	chainID := uuid.New()
	chain := &Chain{
		ID: chainID, ClientID: uuid.New(), UserID: "owner", Scope: "mcp",
		RedirectURI: "http://x/cb", InvalidatedBefore: time.Unix(0, 0), CreatedAt: time.Now(),
	}
	require.NoError(t, chains.NewChain(ctx, chain))

	got, err := chains.GetChain(ctx, chainID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Revoked)

	cutoff := time.Now().Truncate(time.Second)
	require.NoError(t, chains.BumpCutoff(ctx, chainID, cutoff))
	got, err = chains.GetChain(ctx, chainID)
	require.NoError(t, err)
	assert.True(t, got.InvalidatedBefore.Equal(cutoff))

	require.NoError(t, chains.RevokeChain(ctx, chainID))
	got, err = chains.GetChain(ctx, chainID)
	require.NoError(t, err)
	assert.True(t, got.Revoked)
}

func TestBearerTokens(t *testing.T) {
	database := newDB(t)
	tokens := NewBearerTokenRepository(database)
	ctx := context.Background()

	got, err := tokens.GetBearerToken(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)

	tokenID := uuid.New()
	clientID := uuid.New()
	token := &BearerToken{
		ID: tokenID, ClientID: clientID,
		CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 6, 0),
	}
	require.NoError(t, tokens.PutBearerToken(ctx, token))

	// Survives a fresh repository over the same database.
	reopened := NewBearerTokenRepository(database)
	got, err = reopened.GetBearerToken(ctx, tokenID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, clientID, got.ClientID)
	assert.False(t, got.Revoked)

	// Revoke flips the flag; revoking an unknown id errors.
	require.NoError(t, reopened.RevokeBearerToken(ctx, tokenID))
	got, err = reopened.GetBearerToken(ctx, tokenID)
	require.NoError(t, err)
	assert.True(t, got.Revoked)

	assert.Error(t, reopened.RevokeBearerToken(ctx, uuid.New()))
}
