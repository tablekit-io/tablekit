package store

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClients(t *testing.T) {
	storageService := newStore(t)
	ctx := context.Background()

	got, err := storageService.GetClient(ctx, "nope")
	require.NoError(t, err)
	assert.Nil(t, got)

	c := &Client{ClientID: "abc", RedirectURIs: []string{"http://x/cb"}, CreatedAt: time.Now()}
	require.NoError(t, storageService.SaveClient(ctx, c))

	got, err = storageService.GetClient(ctx, "abc")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []string{"http://x/cb"}, got.RedirectURIs)
}

func TestPairingModes(t *testing.T) {
	ctx := context.Background()

	t.Run("once admits one then locks", func(t *testing.T) {
		storageService := newStore(t)
		// Fresh state defaults to "once".
		mode, paired, err := storageService.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingOnce, mode)
		assert.Empty(t, paired)

		ok, err := storageService.TryPair(ctx, "client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// Same client is idempotently allowed.
		ok, err = storageService.TryPair(ctx, "client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// A different client is rejected; mode is now disabled.
		ok, err = storageService.TryPair(ctx, "client-b")
		require.NoError(t, err)
		assert.False(t, ok)

		mode, paired, err = storageService.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingDisabled, mode)
		assert.Equal(t, []string{"client-a"}, paired)
	})

	t.Run("indefinite admits many", func(t *testing.T) {
		storageService := newStore(t)
		require.NoError(t, storageService.SetPairingMode(ctx, PairingIndefinite))
		for _, id := range []string{"a", "b", "c"} {
			ok, err := storageService.TryPair(ctx, id)
			require.NoError(t, err)
			assert.True(t, ok)
		}
		mode, paired, err := storageService.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingIndefinite, mode)
		assert.ElementsMatch(t, []string{"a", "b", "c"}, paired)
	})

	t.Run("disabled rejects new but allows paired", func(t *testing.T) {
		storageService := newStore(t)
		require.NoError(t, storageService.SetPairingMode(ctx, PairingIndefinite))
		_, err := storageService.TryPair(ctx, "known")
		require.NoError(t, err)

		require.NoError(t, storageService.SetPairingMode(ctx, PairingDisabled))
		ok, err := storageService.TryPair(ctx, "stranger")
		require.NoError(t, err)
		assert.False(t, ok)

		ok, err = storageService.TryPair(ctx, "known")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown mode rejected", func(t *testing.T) {
		storageService := newStore(t)
		assert.Error(t, storageService.SetPairingMode(ctx, "bogus"))
	})
}

// TestBearerClient round-trips a bearer client (nil name, empty redirect URIs,
// type "bearer") and confirms it survives a reopen over the same database.
func TestBearerClient(t *testing.T) {
	storageService := newStore(t)
	ctx := context.Background()

	require.NoError(t, storageService.SaveClient(ctx, &Client{
		ClientID: "bearer-1", ClientName: nil, RedirectURIs: []string{},
		Type: "bearer", CreatedAt: time.Now(),
	}))

	reopened := reopen(t, storageService)
	got, err := reopened.GetClient(ctx, "bearer-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.ClientName)
	assert.Equal(t, "bearer", got.Type)
	assert.Empty(t, got.RedirectURIs)
}

// TestTryPairConcurrent is the race centerpiece: under mode=once, many
// goroutines racing to pair distinct clients must yield exactly one winner.
// Run with `go test -race`.
func TestTryPairConcurrent(t *testing.T) {
	storageService := newStore(t) // defaults to once
	ctx := context.Background()

	const n = 32
	var wins int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ok, err := storageService.TryPair(ctx, clientIDFor(i))
			if err == nil && ok {
				atomic.AddInt32(&wins, 1)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&wins))
	mode, paired, err := storageService.PairingStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, PairingDisabled, mode)
	assert.Len(t, paired, 1)
}

func clientIDFor(i int) string {
	b, _ := json.Marshal(i)
	return "client-" + string(b)
}
