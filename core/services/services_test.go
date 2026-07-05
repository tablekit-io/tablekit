package services

import (
	"encoding/base64"
	"os"
	"testing"

	"core/db/dbtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSigningKey is a fixed base64 HS256 secret (SIGNING_KEY is required). The
// plaintext is exactly 32 bytes.
var testSigningKey = base64.StdEncoding.EncodeToString([]byte("tablekit-svc-test-signing-key-32"))

// TestMain starts one throwaway Postgres for the whole package (skipped where
// docker isn't available); New opens tablekit's own state database.
func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func TestNewWiresConfigAndStore(t *testing.T) {
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
	t.Setenv("SIGNING_KEY", testSigningKey)

	appServices, err := New()
	require.NoError(t, err)

	require.NotNil(t, appServices.Config)
	require.NotNil(t, appServices.Clients)
	require.NotNil(t, appServices.AuthCodes)
	require.NotNil(t, appServices.TokenChains)
	require.NotNil(t, appServices.BearerTokens)
	require.NotNil(t, appServices.Pairing)
	require.NotNil(t, appServices.Engine)
	// The JWT issuer is constructed once here and shared across the app.
	require.NotNil(t, appServices.Issuer)
}

func TestNewFailsWithoutSigningKey(t *testing.T) {
	// A reachable database gets New past db.Open, so it reaches issuer
	// construction, which requires SIGNING_KEY → New errors when it is empty.
	t.Setenv("DATABASE_URL", dbtest.NewDSN(t))
	t.Setenv("SIGNING_KEY", "")

	_, err := New()
	assert.Error(t, err)
}
