package oauth

import (
	"io"
	"net/http"
	"testing"

	"core/e2e/harness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pairs reports whether a fresh client can complete /authorize (gets a code)
// or is turned away with the "already paired" page.
func pairs(t *testing.T, appURL string) bool {
	t.Helper()
	clientID := harness.Register(t, appURL)
	_, challenge := harness.PKCEPair(t)
	response := harness.Authorize(t, appURL, clientID, challenge, "")
	defer response.Body.Close()
	return response.StatusCode == http.StatusFound &&
		response.Header.Get("Location") != ""
}

func TestPairingDefaultOnce(t *testing.T) {
	server := harness.StartServer(t)

	// First client pairs.
	assert.True(t, pairs(t, server.AppURL))

	// Second (different) client is refused with the already-paired page that
	// bounces back to the client with an OAuth error.
	clientID := harness.Register(t, server.AppURL)
	_, challenge := harness.PKCEPair(t)
	response := harness.Authorize(t, server.AppURL, clientID, challenge, "")
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	html, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(html), "Already paired")
	assert.Contains(t, string(html), "error=access_denied")
}

func TestPairingEnableIndefinitely(t *testing.T) {
	server := harness.StartServer(t)
	harness.RunCLI(t, server.DataDir, "pairing", "enable", "--indefinitely")

	// Multiple distinct clients can now pair.
	assert.True(t, pairs(t, server.AppURL))
	assert.True(t, pairs(t, server.AppURL))
}

func TestPairingDisable(t *testing.T) {
	server := harness.StartServer(t)

	// Pair one client while in the default "once" mode.
	clientID := harness.Register(t, server.AppURL)
	verifier, challenge := harness.PKCEPair(t)
	code := harness.AuthorizeCode(t, server.AppURL, clientID, challenge)
	harness.ExchangeCode(t, server.AppURL, clientID, code, verifier)

	harness.RunCLI(t, server.DataDir, "pairing", "disable")

	// A new client is blocked...
	assert.False(t, pairs(t, server.AppURL))

	// ...but the already-paired client is still admitted (re-authorize).
	_, challenge2 := harness.PKCEPair(t)
	response := harness.Authorize(t, server.AppURL, clientID, challenge2, "")
	defer response.Body.Close()
	require.Equal(t, http.StatusFound, response.StatusCode)
}
