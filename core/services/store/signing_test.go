package store

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
