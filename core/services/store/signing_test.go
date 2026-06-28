package store

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
