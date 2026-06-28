package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClients(t *testing.T) {
	storageService := newStore(t)

	got, err := storageService.GetClient("nope")
	require.NoError(t, err)
	assert.Nil(t, got)

	c := &Client{ClientID: "abc", RedirectURIs: []string{"http://x/cb"}, CreatedAt: time.Now()}
	require.NoError(t, storageService.SaveClient(c))

	got, err = storageService.GetClient("abc")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []string{"http://x/cb"}, got.RedirectURIs)
}

func TestPairingModes(t *testing.T) {
	t.Run("once admits one then locks", func(t *testing.T) {
		storageService := newStore(t)
		// Fresh state defaults to "once".
		mode, paired, err := storageService.PairingStatus()
		require.NoError(t, err)
		assert.Equal(t, PairingOnce, mode)
		assert.Empty(t, paired)

		ok, err := storageService.TryPair("client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// Same client is idempotently allowed.
		ok, err = storageService.TryPair("client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// A different client is rejected; mode is now disabled.
		ok, err = storageService.TryPair("client-b")
		require.NoError(t, err)
		assert.False(t, ok)

		mode, paired, err = storageService.PairingStatus()
		require.NoError(t, err)
		assert.Equal(t, PairingDisabled, mode)
		assert.Equal(t, []string{"client-a"}, paired)
	})

	t.Run("indefinite admits many", func(t *testing.T) {
		storageService := newStore(t)
		require.NoError(t, storageService.SetPairingMode(PairingIndefinite))
		for _, id := range []string{"a", "b", "c"} {
			ok, err := storageService.TryPair(id)
			require.NoError(t, err)
			assert.True(t, ok)
		}
		mode, paired, err := storageService.PairingStatus()
		require.NoError(t, err)
		assert.Equal(t, PairingIndefinite, mode)
		assert.ElementsMatch(t, []string{"a", "b", "c"}, paired)
	})

	t.Run("disabled rejects new but allows paired", func(t *testing.T) {
		storageService := newStore(t)
		require.NoError(t, storageService.SetPairingMode(PairingIndefinite))
		_, err := storageService.TryPair("known")
		require.NoError(t, err)

		require.NoError(t, storageService.SetPairingMode(PairingDisabled))
		ok, err := storageService.TryPair("stranger")
		require.NoError(t, err)
		assert.False(t, ok)

		ok, err = storageService.TryPair("known")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown mode rejected", func(t *testing.T) {
		storageService := newStore(t)
		assert.Error(t, storageService.SetPairingMode("bogus"))
	})
}

func TestLegacyPairedClientIDMigration(t *testing.T) {
	directory := t.TempDir()
	// Hand-write a pre-multi-client clients.json.
	legacy := `{"paired_client_id":"old-client","clients":{}}`
	require.NoError(t, os.WriteFile(filepath.Join(directory, "clients.json"), []byte(legacy), 0o600))

	storageService, err := New(directory)
	require.NoError(t, err)

	mode, paired, err := storageService.PairingStatus()
	require.NoError(t, err)
	assert.Equal(t, PairingOnce, mode) // default applied
	assert.Equal(t, []string{"old-client"}, paired)

	// A resave drops the legacy key in favor of the list.
	require.NoError(t, storageService.SetPairingMode(PairingDisabled))
	raw, err := os.ReadFile(filepath.Join(directory, "clients.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "paired_client_id\"")
	assert.Contains(t, string(raw), "paired_client_ids")
}

func TestEmptyPairedListMarshalsToArray(t *testing.T) {
	directory := t.TempDir()
	storageService, err := New(directory)
	require.NoError(t, err)

	// SaveClient writes clients.json with an empty (but non-nil) paired list.
	require.NoError(t, storageService.SaveClient(&Client{
		ClientID: "x", RedirectURIs: []string{"http://x/cb"}, CreatedAt: time.Now(),
	}))

	raw, err := os.ReadFile(filepath.Join(directory, "clients.json"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"paired_client_ids": []`)

	// Reopening validates against the schema; [] (not null) must pass.
	_, err = New(directory)
	require.NoError(t, err)
}

func TestBearerClientNullName(t *testing.T) {
	storageService := newStore(t)

	// A bearer client: nil name, empty redirect URIs, type "bearer".
	require.NoError(t, storageService.SaveClient(&Client{
		ClientID: "bearer-1", ClientName: nil, RedirectURIs: []string{},
		Type: "bearer", CreatedAt: time.Now(),
	}))

	raw, err := os.ReadFile(filepath.Join(storageService.directory, "clients.json"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"client_name": null`)
	assert.Contains(t, string(raw), `"type": "bearer"`)
	assert.Contains(t, string(raw), `"redirect_uris": []`)

	// Reopening validates the (empty redirect_uris, null name) shape against the schema.
	reopened, err := New(storageService.directory)
	require.NoError(t, err)
	got, err := reopened.GetClient("bearer-1")
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

	const n = 32
	var wins int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ok, err := storageService.TryPair(clientIDFor(i))
			if err == nil && ok {
				atomic.AddInt32(&wins, 1)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&wins))
	mode, paired, err := storageService.PairingStatus()
	require.NoError(t, err)
	assert.Equal(t, PairingDisabled, mode)
	assert.Len(t, paired, 1)
}

func clientIDFor(i int) string {
	b, _ := json.Marshal(i)
	return "client-" + string(b)
}
