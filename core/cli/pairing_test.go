package cli

import (
	"context"
	"os"
	"testing"

	"core/db"
	"core/db/dbtest"
	"core/services/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available). The pairing commands construct services.New, which
// opens tablekit's own state database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func TestPairingEnable(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))

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
			pairingEnableCmd.SetContext(context.Background())
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
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
	pairingDisableCmd.SetContext(context.Background())
	require.NoError(t, pairingDisableCmd.RunE(pairingDisableCmd, nil))
	assertMode(t, store.PairingDisabled)
}

func assertMode(t *testing.T, want string) {
	t.Helper()
	database, err := db.Open(os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	storageService, err := store.New(os.Getenv("DATA_DIR"), database)
	require.NoError(t, err)
	mode, _, err := storageService.PairingStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, want, mode)
}
