package cli

import (
	"os"
	"testing"

	"core/services/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPairingEnable(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())

	tests := []struct {
		name     string
		once     bool
		indef    bool
		wantErr  bool
		wantMode string
	}{
		{"neither flag", false, false, true, ""},
		{"both flags", true, true, true, ""},
		{"once", true, false, false, store.PairingOnce},
		{"indefinitely", false, true, false, store.PairingIndefinite},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairingOnce, pairingIndefinitely = tt.once, tt.indef
			err := pairingEnableCmd.RunE(pairingEnableCmd, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assertMode(t, tt.wantMode)
		})
	}
}

func TestPairingDisable(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	require.NoError(t, pairingDisableCmd.RunE(pairingDisableCmd, nil))
	assertMode(t, store.PairingDisabled)
}

func assertMode(t *testing.T, want string) {
	t.Helper()
	storageService, err := store.New(os.Getenv("DATA_DIR"))
	require.NoError(t, err)
	mode, _, err := storageService.PairingStatus()
	require.NoError(t, err)
	assert.Equal(t, want, mode)
}
