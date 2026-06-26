package e2e

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pairs reports whether a fresh client can complete /authorize (gets a code)
// or is turned away with the "already paired" page.
func pairs(t *testing.T, appURL string) bool {
	t.Helper()
	clientID := register(t, appURL)
	_, challenge := pkcePair(t)
	resp := authorize(t, appURL, clientID, challenge, "")
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusFound &&
		resp.Header.Get("Location") != ""
}

func TestPairingDefaultOnce(t *testing.T) {
	srv := startServer(t)

	// First client pairs.
	assert.True(t, pairs(t, srv.appURL))

	// Second (different) client is refused with the already-paired page that
	// bounces back to the client with an OAuth error.
	clientID := register(t, srv.appURL)
	_, challenge := pkcePair(t)
	resp := authorize(t, srv.appURL, clientID, challenge, "")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	html, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(html), "Already paired")
	assert.Contains(t, string(html), "error=access_denied")
}

func TestPairingEnableIndefinitely(t *testing.T) {
	srv := startServer(t)
	runCLI(t, srv.dataDir, "pairing", "enable", "--indefinitely")

	// Multiple distinct clients can now pair.
	assert.True(t, pairs(t, srv.appURL))
	assert.True(t, pairs(t, srv.appURL))
}

func TestPairingDisable(t *testing.T) {
	srv := startServer(t)

	// Pair one client while in the default "once" mode.
	clientID := register(t, srv.appURL)
	verifier, challenge := pkcePair(t)
	code := authorizeCode(t, srv.appURL, clientID, challenge)
	exchangeCode(t, srv.appURL, clientID, code, verifier)

	runCLI(t, srv.dataDir, "pairing", "disable")

	// A new client is blocked...
	assert.False(t, pairs(t, srv.appURL))

	// ...but the already-paired client is still admitted (re-authorize).
	_, challenge2 := pkcePair(t)
	resp := authorize(t, srv.appURL, clientID, challenge2, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusFound, resp.StatusCode)
}
