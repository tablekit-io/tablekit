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
	response := authorize(t, appURL, clientID, challenge, "")
	defer response.Body.Close()
	return response.StatusCode == http.StatusFound &&
		response.Header.Get("Location") != ""
}

func TestPairingDefaultOnce(t *testing.T) {
	server := startServer(t)

	// First client pairs.
	assert.True(t, pairs(t, server.appURL))

	// Second (different) client is refused with the already-paired page that
	// bounces back to the client with an OAuth error.
	clientID := register(t, server.appURL)
	_, challenge := pkcePair(t)
	response := authorize(t, server.appURL, clientID, challenge, "")
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	html, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(html), "Already paired")
	assert.Contains(t, string(html), "error=access_denied")
}

func TestPairingEnableIndefinitely(t *testing.T) {
	server := startServer(t)
	runCLI(t, server.dataDir, "pairing", "enable", "--indefinitely")

	// Multiple distinct clients can now pair.
	assert.True(t, pairs(t, server.appURL))
	assert.True(t, pairs(t, server.appURL))
}

func TestPairingDisable(t *testing.T) {
	server := startServer(t)

	// Pair one client while in the default "once" mode.
	clientID := register(t, server.appURL)
	verifier, challenge := pkcePair(t)
	code := authorizeCode(t, server.appURL, clientID, challenge)
	exchangeCode(t, server.appURL, clientID, code, verifier)

	runCLI(t, server.dataDir, "pairing", "disable")

	// A new client is blocked...
	assert.False(t, pairs(t, server.appURL))

	// ...but the already-paired client is still admitted (re-authorize).
	_, challenge2 := pkcePair(t)
	response := authorize(t, server.appURL, clientID, challenge2, "")
	defer response.Body.Close()
	require.Equal(t, http.StatusFound, response.StatusCode)
}
