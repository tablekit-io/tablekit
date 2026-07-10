package oauth

import (
	"io"
	"net/http"
	"testing"

	"core/e2e/harness"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthorizeSelfHealInDevelopment: in development, /authorize recreates a
// client_id it has never seen (as happens when a wiped dev DB drops the row the
// MCP client still has cached) and issues a usable code, so reconnect works
// without re-registering.
func TestAuthorizeSelfHealInDevelopment(t *testing.T) {
	server := harness.StartServerEnv(t, "TABLEKIT_ENV=development")

	// A client_id the server never registered: the stale-cached-id scenario.
	unknownClient := uuid.NewString()
	verifier, challenge := harness.PKCEPair(t)

	// Development self-heal: /authorize creates the client and returns a code.
	code := harness.AuthorizeCode(t, server.AppURL, unknownClient, challenge)

	// The healed client is now real — the code redeems for tokens. A successful
	// exchange proves the client row exists and the auth code was persisted
	// against it (both would fail otherwise, including the dev foreign keys).
	tokens := harness.ExchangeCode(t, server.AppURL, unknownClient, code, verifier)
	assert.NotEmpty(t, tokens["access_token"])
}

// TestAuthorizeRejectsUnknownClientOutsideDevelopment: with no TABLEKIT_ENV the
// server keeps the strict production behavior — an unknown client_id at
// /authorize is rejected and no client is created.
func TestAuthorizeRejectsUnknownClientOutsideDevelopment(t *testing.T) {
	server := harness.StartServer(t) // no TABLEKIT_ENV => production

	unknownClient := uuid.NewString()
	_, challenge := harness.PKCEPair(t)

	response := harness.Authorize(t, server.AppURL, unknownClient, challenge, "")
	defer response.Body.Close()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "unknown client_id")
}
