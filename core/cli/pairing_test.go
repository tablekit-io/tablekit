package cli

import (
	"context"
	"encoding/base64"
	"os"
	"testing"

	"core/db"
	"core/db/dbtest"
	"core/services/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSigningKey is a fixed base64 HS256 secret; services.New builds the issuer,
// which now requires SIGNING_KEY. The plaintext is exactly 32 bytes.
var testSigningKey = base64.StdEncoding.EncodeToString([]byte("tablekit-cli-test-signing-key-32"))

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available). The pairing commands construct services.New, which
// opens tablekit's own state database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func TestPairingEnable(t *testing.T) {
	t.Setenv("SIGNING_KEY", testSigningKey)
	useIsolatedDatabase(t)

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
	t.Setenv("SIGNING_KEY", testSigningKey)
	useIsolatedDatabase(t)
	pairingDisableCmd.SetContext(context.Background())
	require.NoError(t, pairingDisableCmd.RunE(pairingDisableCmd, nil))
	assertMode(t, store.PairingDisabled)
}

// useIsolatedDatabase points the command under test at a fresh, isolated state
// database via DATABASE_URL. It also clears the structured DB_* vars (which the
// dev container sets, and which otherwise take precedence in DatabaseDSN), so
// both the command and assertMode resolve to this same test database.
func useIsolatedDatabase(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
}

func assertMode(t *testing.T, want string) {
	t.Helper()
	database, err := db.Open(os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	storageService := store.New(database)
	mode, _, err := storageService.PairingStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, want, mode)
}
