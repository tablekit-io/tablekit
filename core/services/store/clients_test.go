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
	c := &Client{ClientID: clientID, RedirectURIs: []string{"http://x/cb"}, Type: ClientTypeOAuth, CreatedAt: time.Now()}
	require.NoError(t, clients.SaveClient(ctx, c))

	got, err = clients.GetClient(ctx, clientID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []string{"http://x/cb"}, got.RedirectURIs)
}

// TestStaticClient round-trips a static client (nil name, empty redirect URIs,
// type "static") and confirms it survives a fresh repository over the same
// database (state persists in Postgres, standing in for a process restart).
func TestStaticClient(t *testing.T) {
	database := newDB(t)
	clients := NewClientRepository(database)
	ctx := context.Background()

	staticID := uuid.New()
	require.NoError(t, clients.SaveClient(ctx, &Client{
		ClientID: staticID, ClientName: nil, RedirectURIs: []string{},
		Type: ClientTypeStatic, CreatedAt: time.Now(),
	}))

	reopened := NewClientRepository(database)
	got, err := reopened.GetClient(ctx, staticID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.ClientName)
	assert.Equal(t, ClientTypeStatic, got.Type)
	assert.Empty(t, got.RedirectURIs)
}
