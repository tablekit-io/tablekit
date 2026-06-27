package store

import (
	"encoding/base64"
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

// newStore returns a Store backed by a fresh temp dir.
func newStore(t *testing.T) *Store {
	t.Helper()
	storageService, err := New(t.TempDir())
	require.NoError(t, err)
	return storageService
}

func TestSigningKeyGeneratesBase64(t *testing.T) {
	storageService := newStore(t)

	key1, err := storageService.SigningKey()
	require.NoError(t, err)
	assert.Len(t, key1, 32)

	// The file is base64 text that decodes back to the key, not a binary blob.
	raw, err := os.ReadFile(filepath.Join(storageService.directory, "signing.key"))
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	require.NoError(t, err)
	assert.Equal(t, key1, decoded)

	// Stable across calls and across a fresh Store over the same dir.
	key2, err := storageService.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, key1, key2)

	reopened, err := New(storageService.directory)
	require.NoError(t, err)
	key3, err := reopened.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, key1, key3)
}

func TestSigningKeyLegacyMigrationAtBoot(t *testing.T) {
	directory := t.TempDir()
	keyPath := filepath.Join(directory, "signing.key")

	// Pre-base64 format: exactly 32 raw bytes.
	legacy := make([]byte, 32)
	for i := range legacy {
		legacy[i] = byte(i + 1)
	}
	require.NoError(t, os.WriteFile(keyPath, legacy, 0o600))

	// New() migrates the file in place at boot.
	storageService, err := New(directory)
	require.NoError(t, err)

	raw, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	require.NoError(t, err)
	assert.Equal(t, legacy, decoded, "file rewritten as base64 of the legacy key")

	key, err := storageService.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, legacy, key, "same key bytes preserved across migration")

	// Re-opening is idempotent.
	storageService2, err := New(directory)
	require.NoError(t, err)
	key2, err := storageService2.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, legacy, key2)
}

func TestDecodeSigningKey(t *testing.T) {
	key32 := make([]byte, 32)
	for i := range key32 {
		key32[i] = byte(i)
	}
	key40 := append(append([]byte{}, key32...), []byte("abcdefgh")...)

	tests := []struct {
		name    string
		in      string
		want    []byte
		wantErr bool
	}{
		{"valid 32", base64.StdEncoding.EncodeToString(key32), key32, false},
		{"raw unpadded", base64.RawStdEncoding.EncodeToString(key32), key32, false},
		{"longer kept", base64.StdEncoding.EncodeToString(key40), key40, false},
		{
			"short padded",
			base64.StdEncoding.EncodeToString([]byte("sixteen-byte-key")),
			append([]byte("sixteen-byte-key"), make([]byte, 16)...),
			false,
		},
		{"empty", "", nil, true},
		{"not base64", "!!!not-base64!!!", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeSigningKey(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.GreaterOrEqual(t, len(got), 32)
		})
	}
}

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

func TestAuthCodeSingleUse(t *testing.T) {
	storageService := newStore(t)
	code := &AuthCode{
		Code: "c1", ClientID: "a", RedirectURI: "http://x/cb",
		CodeChallenge: "chal", Scope: "mcp", UserID: "owner",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	require.NoError(t, storageService.PutCode(code))

	got, err := storageService.ConsumeCode("c1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "chal", got.CodeChallenge)

	// Second consume returns nil — codes are one-time.
	got, err = storageService.ConsumeCode("c1")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestChains(t *testing.T) {
	storageService := newStore(t)
	chain := &Chain{
		ID: "ch1", ClientID: "a", UserID: "owner", Scope: "mcp",
		RedirectURI: "http://x/cb", InvalidatedBefore: time.Unix(0, 0), CreatedAt: time.Now(),
	}
	require.NoError(t, storageService.NewChain(chain))

	got, err := storageService.GetChain("ch1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Revoked)

	cutoff := time.Now().Truncate(time.Second)
	require.NoError(t, storageService.BumpCutoff("ch1", cutoff))
	got, err = storageService.GetChain("ch1")
	require.NoError(t, err)
	assert.True(t, got.InvalidatedBefore.Equal(cutoff))

	require.NoError(t, storageService.RevokeChain("ch1"))
	got, err = storageService.GetChain("ch1")
	require.NoError(t, err)
	assert.True(t, got.Revoked)
}

func TestBearerTokens(t *testing.T) {
	storageService := newStore(t)

	got, err := storageService.GetBearerToken("nope")
	require.NoError(t, err)
	assert.Nil(t, got)

	token := &BearerToken{
		ID: "tok-1", ClientID: "client-1",
		CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 6, 0),
	}
	require.NoError(t, storageService.PutBearerToken(token))

	// Survives a reopen (and passes schema validation on load).
	reopened, err := New(storageService.directory)
	require.NoError(t, err)
	got, err = reopened.GetBearerToken("tok-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "client-1", got.ClientID)
	assert.False(t, got.Revoked)

	// Revoke flips the flag; revoking an unknown id errors.
	require.NoError(t, reopened.RevokeBearerToken("tok-1"))
	got, err = reopened.GetBearerToken("tok-1")
	require.NoError(t, err)
	assert.True(t, got.Revoked)

	assert.Error(t, reopened.RevokeBearerToken("ghost"))
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

func TestNewFailsOnCorruptState(t *testing.T) {
	directory := t.TempDir()
	// Client object missing the required client_id field.
	bad := `{"clients":{"x":{"redirect_uris":["http://x/cb"],"created_at":"2026-01-01T00:00:00Z"}}}`
	require.NoError(t, os.WriteFile(filepath.Join(directory, "clients.json"), []byte(bad), 0o600))

	_, err := New(directory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
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
