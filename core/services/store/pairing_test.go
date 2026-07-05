package store

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

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

		ok, err := pairing.TryPair(ctx, "client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// Same client is idempotently allowed.
		ok, err = pairing.TryPair(ctx, "client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// A different client is rejected; mode is now disabled.
		ok, err = pairing.TryPair(ctx, "client-b")
		require.NoError(t, err)
		assert.False(t, ok)

		mode, paired, err = pairing.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingDisabled, mode)
		assert.Equal(t, []string{"client-a"}, paired)
	})

	t.Run("indefinite admits many", func(t *testing.T) {
		pairing := NewPairingRepository(newDB(t))
		require.NoError(t, pairing.SetPairingMode(ctx, PairingIndefinite))
		for _, id := range []string{"a", "b", "c"} {
			ok, err := pairing.TryPair(ctx, id)
			require.NoError(t, err)
			assert.True(t, ok)
		}
		mode, paired, err := pairing.PairingStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, PairingIndefinite, mode)
		assert.ElementsMatch(t, []string{"a", "b", "c"}, paired)
	})

	t.Run("disabled rejects new but allows paired", func(t *testing.T) {
		pairing := NewPairingRepository(newDB(t))
		require.NoError(t, pairing.SetPairingMode(ctx, PairingIndefinite))
		_, err := pairing.TryPair(ctx, "known")
		require.NoError(t, err)

		require.NoError(t, pairing.SetPairingMode(ctx, PairingDisabled))
		ok, err := pairing.TryPair(ctx, "stranger")
		require.NoError(t, err)
		assert.False(t, ok)

		ok, err = pairing.TryPair(ctx, "known")
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

func clientIDFor(i int) string {
	b, _ := json.Marshal(i)
	return "client-" + string(b)
}
