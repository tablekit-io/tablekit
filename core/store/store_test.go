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
	st, err := New(t.TempDir())
	require.NoError(t, err)
	return st
}

func TestSigningKeyGeneratesBase64(t *testing.T) {
	st := newStore(t)

	k1, err := st.SigningKey()
	require.NoError(t, err)
	assert.Len(t, k1, 32)

	// The file is base64 text that decodes back to the key, not a binary blob.
	raw, err := os.ReadFile(filepath.Join(st.dir, "signing.key"))
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	require.NoError(t, err)
	assert.Equal(t, k1, decoded)

	// Stable across calls and across a fresh Store over the same dir.
	k2, err := st.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, k1, k2)

	reopened, err := New(st.dir)
	require.NoError(t, err)
	k3, err := reopened.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, k1, k3)
}

func TestSigningKeyLegacyMigrationAtBoot(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "signing.key")

	// Pre-base64 format: exactly 32 raw bytes.
	legacy := make([]byte, 32)
	for i := range legacy {
		legacy[i] = byte(i + 1)
	}
	require.NoError(t, os.WriteFile(keyPath, legacy, 0o600))

	// New() migrates the file in place at boot.
	st, err := New(dir)
	require.NoError(t, err)

	raw, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	require.NoError(t, err)
	assert.Equal(t, legacy, decoded, "file rewritten as base64 of the legacy key")

	key, err := st.SigningKey()
	require.NoError(t, err)
	assert.Equal(t, legacy, key, "same key bytes preserved across migration")

	// Re-opening is idempotent.
	st2, err := New(dir)
	require.NoError(t, err)
	key2, err := st2.SigningKey()
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
	st := newStore(t)

	got, err := st.GetClient("nope")
	require.NoError(t, err)
	assert.Nil(t, got)

	c := &Client{ClientID: "abc", RedirectURIs: []string{"http://x/cb"}, CreatedAt: time.Now()}
	require.NoError(t, st.SaveClient(c))

	got, err = st.GetClient("abc")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []string{"http://x/cb"}, got.RedirectURIs)
}

func TestPairingModes(t *testing.T) {
	t.Run("once admits one then locks", func(t *testing.T) {
		st := newStore(t)
		// Fresh state defaults to "once".
		mode, paired, err := st.PairingStatus()
		require.NoError(t, err)
		assert.Equal(t, PairingOnce, mode)
		assert.Empty(t, paired)

		ok, err := st.TryPair("client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// Same client is idempotently allowed.
		ok, err = st.TryPair("client-a")
		require.NoError(t, err)
		assert.True(t, ok)

		// A different client is rejected; mode is now disabled.
		ok, err = st.TryPair("client-b")
		require.NoError(t, err)
		assert.False(t, ok)

		mode, paired, err = st.PairingStatus()
		require.NoError(t, err)
		assert.Equal(t, PairingDisabled, mode)
		assert.Equal(t, []string{"client-a"}, paired)
	})

	t.Run("indefinite admits many", func(t *testing.T) {
		st := newStore(t)
		require.NoError(t, st.SetPairingMode(PairingIndefinite))
		for _, id := range []string{"a", "b", "c"} {
			ok, err := st.TryPair(id)
			require.NoError(t, err)
			assert.True(t, ok)
		}
		mode, paired, err := st.PairingStatus()
		require.NoError(t, err)
		assert.Equal(t, PairingIndefinite, mode)
		assert.ElementsMatch(t, []string{"a", "b", "c"}, paired)
	})

	t.Run("disabled rejects new but allows paired", func(t *testing.T) {
		st := newStore(t)
		require.NoError(t, st.SetPairingMode(PairingIndefinite))
		_, err := st.TryPair("known")
		require.NoError(t, err)

		require.NoError(t, st.SetPairingMode(PairingDisabled))
		ok, err := st.TryPair("stranger")
		require.NoError(t, err)
		assert.False(t, ok)

		ok, err = st.TryPair("known")
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown mode rejected", func(t *testing.T) {
		st := newStore(t)
		assert.Error(t, st.SetPairingMode("bogus"))
	})
}

func TestLegacyPairedClientIDMigration(t *testing.T) {
	dir := t.TempDir()
	// Hand-write a pre-multi-client clients.json.
	legacy := `{"paired_client_id":"old-client","clients":{}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "clients.json"), []byte(legacy), 0o600))

	st, err := New(dir)
	require.NoError(t, err)

	mode, paired, err := st.PairingStatus()
	require.NoError(t, err)
	assert.Equal(t, PairingOnce, mode) // default applied
	assert.Equal(t, []string{"old-client"}, paired)

	// A resave drops the legacy key in favor of the list.
	require.NoError(t, st.SetPairingMode(PairingDisabled))
	raw, err := os.ReadFile(filepath.Join(dir, "clients.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "paired_client_id\"")
	assert.Contains(t, string(raw), "paired_client_ids")
}

func TestEmptyPairedListMarshalsToArray(t *testing.T) {
	dir := t.TempDir()
	st, err := New(dir)
	require.NoError(t, err)

	// SaveClient writes clients.json with an empty (but non-nil) paired list.
	require.NoError(t, st.SaveClient(&Client{
		ClientID: "x", RedirectURIs: []string{"http://x/cb"}, CreatedAt: time.Now(),
	}))

	raw, err := os.ReadFile(filepath.Join(dir, "clients.json"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"paired_client_ids": []`)

	// Reopening validates against the schema; [] (not null) must pass.
	_, err = New(dir)
	require.NoError(t, err)
}

func TestAuthCodeSingleUse(t *testing.T) {
	st := newStore(t)
	code := &AuthCode{
		Code: "c1", ClientID: "a", RedirectURI: "http://x/cb",
		CodeChallenge: "chal", Scope: "mcp", UserID: "owner",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	require.NoError(t, st.PutCode(code))

	got, err := st.ConsumeCode("c1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "chal", got.CodeChallenge)

	// Second consume returns nil — codes are one-time.
	got, err = st.ConsumeCode("c1")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestChains(t *testing.T) {
	st := newStore(t)
	ch := &Chain{
		ID: "ch1", ClientID: "a", UserID: "owner", Scope: "mcp",
		RedirectURI: "http://x/cb", InvalidatedBefore: time.Unix(0, 0), CreatedAt: time.Now(),
	}
	require.NoError(t, st.NewChain(ch))

	got, err := st.GetChain("ch1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Revoked)

	cutoff := time.Now().Truncate(time.Second)
	require.NoError(t, st.BumpCutoff("ch1", cutoff))
	got, err = st.GetChain("ch1")
	require.NoError(t, err)
	assert.True(t, got.InvalidatedBefore.Equal(cutoff))

	require.NoError(t, st.RevokeChain("ch1"))
	got, err = st.GetChain("ch1")
	require.NoError(t, err)
	assert.True(t, got.Revoked)
}

func TestNewFailsOnCorruptState(t *testing.T) {
	dir := t.TempDir()
	// Client object missing the required client_id field.
	bad := `{"clients":{"x":{"redirect_uris":["http://x/cb"],"created_at":"2026-01-01T00:00:00Z"}}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "clients.json"), []byte(bad), 0o600))

	_, err := New(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

// TestTryPairConcurrent is the race centerpiece: under mode=once, many
// goroutines racing to pair distinct clients must yield exactly one winner.
// Run with `go test -race`.
func TestTryPairConcurrent(t *testing.T) {
	st := newStore(t) // defaults to once

	const n = 32
	var wins int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ok, err := st.TryPair(clientIDFor(i))
			if err == nil && ok {
				atomic.AddInt32(&wins, 1)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&wins))
	mode, paired, err := st.PairingStatus()
	require.NoError(t, err)
	assert.Equal(t, PairingDisabled, mode)
	assert.Len(t, paired, 1)
}

func clientIDFor(i int) string {
	b, _ := json.Marshal(i)
	return "client-" + string(b)
}
