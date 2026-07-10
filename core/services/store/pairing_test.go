package store

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPairingModes(t *testing.T) {
	ctx := context.Background()

	t.Run("once admits one then locks", func(t *testing.T) {
		database := newDB(t)
		pairing := NewPairingRepository(database)
		// Fresh state defaults to "once".
		mode, paired, err := pairing.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingOnce, mode)
		assert.Empty(t, paired)

		clientA := seedClient(t, database)
		clientB := seedClient(t, database)
		ok, err := pairing.TryPair(ctx, clientA)
		require.NoError(t, err)
		assert.True(t, ok)

		// Same client is idempotently allowed.
		ok, err = pairing.TryPair(ctx, clientA)
		require.NoError(t, err)
		assert.True(t, ok)

		// A different client is rejected; mode is now disabled.
		ok, err = pairing.TryPair(ctx, clientB)
		require.NoError(t, err)
		assert.False(t, ok)

		mode, paired, err = pairing.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingDisabled, mode)
		assert.Equal(t, []uuid.UUID{clientA}, paired)
	})

	t.Run("indefinite admits many", func(t *testing.T) {
		database := newDB(t)
		pairing := NewPairingRepository(database)
		require.NoError(t, pairing.SetPairingMode(ctx, PairingIndefinite))
		ids := []uuid.UUID{seedClient(t, database), seedClient(t, database), seedClient(t, database)}
		for _, id := range ids {
			ok, err := pairing.TryPair(ctx, id)
			require.NoError(t, err)
			assert.True(t, ok)
		}
		mode, paired, err := pairing.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingIndefinite, mode)
		assert.ElementsMatch(t, ids, paired)
	})

	t.Run("disabled rejects new but allows paired", func(t *testing.T) {
		database := newDB(t)
		pairing := NewPairingRepository(database)
		require.NoError(t, pairing.SetPairingMode(ctx, PairingIndefinite))
		known := seedClient(t, database)
		_, err := pairing.TryPair(ctx, known)
		require.NoError(t, err)

		require.NoError(t, pairing.SetPairingMode(ctx, PairingDisabled))
		newClient := seedClient(t, database)
		ok, err := pairing.TryPair(ctx, newClient)
		require.NoError(t, err)
		assert.False(t, ok)

		ok, err = pairing.TryPair(ctx, known)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown mode rejected", func(t *testing.T) {
		pairing := NewPairingRepository(newDB(t))
		assert.Error(t, pairing.SetPairingMode(ctx, "bogus"))
	})
}

// TestTryPairConcurrent is the race centerpiece: under mode=once, many
// goroutines racing to pair distinct clients must yield exactly one winner.
// Run with `go test -race`.
func TestTryPairConcurrent(t *testing.T) {
	database := newDB(t)
	pairing := NewPairingRepository(database) // defaults to once
	ctx := context.Background()

	const n = 32
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = seedClient(t, database)
	}

	var wins int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ok, err := pairing.TryPair(ctx, ids[i])
			if err == nil && ok {
				atomic.AddInt32(&wins, 1)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&wins))
	mode, paired, err := pairing.PairingStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, PairingDisabled, mode)
	assert.Len(t, paired, 1)
}
