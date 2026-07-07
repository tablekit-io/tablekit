package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClients(t *testing.T) {
	clients := NewClientRepository(newDB(t))
	ctx := context.Background()

	got, err := clients.GetClient(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)

	clientID := uuid.New()
	c := &Client{ClientID: clientID, RedirectURIs: []string{"http://x/cb"}, CreatedAt: time.Now()}
	require.NoError(t, clients.SaveClient(ctx, c))

	got, err = clients.GetClient(ctx, clientID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []string{"http://x/cb"}, got.RedirectURIs)
}

// TestBearerClient round-trips a bearer client (nil name, empty redirect URIs,
// type "bearer") and confirms it survives a fresh repository over the same
// database (state persists in Postgres, standing in for a process restart).
func TestBearerClient(t *testing.T) {
	database := newDB(t)
	clients := NewClientRepository(database)
	ctx := context.Background()

	bearerID := uuid.New()
	require.NoError(t, clients.SaveClient(ctx, &Client{
		ClientID: bearerID, ClientName: nil, RedirectURIs: []string{},
		Type: "bearer", CreatedAt: time.Now(),
	}))

	reopened := NewClientRepository(database)
	got, err := reopened.GetClient(ctx, bearerID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.ClientName)
	assert.Equal(t, "bearer", got.Type)
	assert.Empty(t, got.RedirectURIs)
}
