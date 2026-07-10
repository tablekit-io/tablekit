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
	database := newDB(t)
	clientID := seedClient(t, database)
	authCodes := NewAuthCodeRepository(database)
	ctx := context.Background()
	codeID := uuid.New()
	code := &AuthCode{
		ID: codeID, Code: codeID.String(), ClientID: clientID, RedirectURI: "http://x/cb",
		CodeChallenge: "chal", Scope: "mcp", UserID: "owner",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	require.NoError(t, authCodes.PutCode(ctx, code))

	got, err := authCodes.ConsumeCode(ctx, codeID.String())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "chal", got.CodeChallenge)

	// Second consume returns nil — codes are one-time.
	got, err = authCodes.ConsumeCode(ctx, codeID.String())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestChains(t *testing.T) {
	database := newDB(t)
	clientID := seedClient(t, database)
	chains := NewTokenChainRepository(database)
	ctx := context.Background()
	chainID := uuid.New()
	chain := &Chain{
		ID: chainID, ClientID: clientID, UserID: "owner", Scope: "mcp",
		RedirectURI: "http://x/cb", InvalidatedBefore: time.Unix(0, 0), CreatedAt: time.Now(),
	}
	require.NoError(t, chains.NewChain(ctx, chain))

	got, err := chains.GetChain(ctx, chainID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Revoked())

	cutoff := time.Now().Truncate(time.Second)
	require.NoError(t, chains.BumpCutoff(ctx, chainID, cutoff))
	got, err = chains.GetChain(ctx, chainID)
	require.NoError(t, err)
	assert.True(t, got.InvalidatedBefore.Equal(cutoff))

	require.NoError(t, chains.RevokeChain(ctx, chainID))
	got, err = chains.GetChain(ctx, chainID)
	require.NoError(t, err)
	assert.True(t, got.Revoked())
}

func TestStaticTokens(t *testing.T) {
	database := newDB(t)
	tokens := NewStaticTokenRepository(database)
	ctx := context.Background()

	got, err := tokens.GetStaticToken(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)

	tokenID := uuid.New()
	clientID := seedClient(t, database)
	token := &StaticToken{
		ID: tokenID, ClientID: clientID,
		CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 6, 0),
	}
	require.NoError(t, tokens.PutStaticToken(ctx, token))

	// Survives a fresh repository over the same database.
	reopened := NewStaticTokenRepository(database)
	got, err = reopened.GetStaticToken(ctx, tokenID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, clientID, got.ClientID)
	assert.False(t, got.Revoked())

	// Revoke stamps revoked_at; revoking an unknown id errors.
	require.NoError(t, reopened.RevokeStaticToken(ctx, tokenID))
	got, err = reopened.GetStaticToken(ctx, tokenID)
	require.NoError(t, err)
	assert.True(t, got.Revoked())

	assert.Error(t, reopened.RevokeStaticToken(ctx, uuid.New()))
}
