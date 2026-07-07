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
		pairing := NewPairingRepository(newDB(t))
		// Fresh state defaults to "once".
		mode, paired, err := pairing.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingOnce, mode)
		assert.Empty(t, paired)

		clientA := uuid.New()
		clientB := uuid.New()
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
		pairing := NewPairingRepository(newDB(t))
		require.NoError(t, pairing.SetPairingMode(ctx, PairingIndefinite))
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
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
		pairing := NewPairingRepository(newDB(t))
		require.NoError(t, pairing.SetPairingMode(ctx, PairingIndefinite))
		known := uuid.New()
		_, err := pairing.TryPair(ctx, known)
		require.NoError(t, err)

		require.NoError(t, pairing.SetPairingMode(ctx, PairingDisabled))
		ok, err := pairing.TryPair(ctx, uuid.New())
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
	pairing := NewPairingRepository(newDB(t)) // defaults to once
	ctx := context.Background()

	const n = 32
	var wins int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ok, err := pairing.TryPair(ctx, clientIDFor(i))
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

// clientIDFor returns a distinct client id per goroutine index; the concurrent
// pairing race only needs the ids to differ, and every uuid.New() does.
func clientIDFor(int) uuid.UUID {
	return uuid.New()
}
