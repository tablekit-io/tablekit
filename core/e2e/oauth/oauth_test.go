package oauth

import (
	"net/http"
	"net/url"
	"testing"

	"core/e2e/harness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthHappyPath(t *testing.T) {
	server := harness.StartServer(t)
	_, tokens := harness.FullHandshake(t, server.AppURL)

	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
	assert.Equal(t, "Bearer", tokens["token_type"])
	assert.Equal(t, "mcp", tokens["scope"])
}

func TestRefreshRotationAndReplay(t *testing.T) {
	server := harness.StartServer(t)
	clientID, tokens := harness.FullHandshake(t, server.AppURL)
	oldRefresh := tokens["refresh_token"].(string)

	// Rotation: old refresh yields a new pair.
	status, body := harness.PostForm(t, server.AppURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {oldRefresh},
	})
	require.Equal(t, http.StatusOK, status, body)
	newRefresh := body["refresh_token"].(string)
	assert.NotEqual(t, oldRefresh, newRefresh)

	// Replay: reusing the old refresh is rejected as theft.
	status, body = harness.PostForm(t, server.AppURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {oldRefresh},
	})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", body["error"])

	// And the whole chain is now revoked: the rotated refresh fails too.
	status, body = harness.PostForm(t, server.AppURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {newRefresh},
	})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", body["error"])
}

func TestMetadataEndpoints(t *testing.T) {
	server := harness.StartServer(t)

	response, err := http.Get(server.AppURL + "/.well-known/oauth-authorization-server")
	require.NoError(t, err)
	var authServerMeta map[string]any
	require.NoError(t, harness.Decode(response, &authServerMeta))
	assert.Equal(t, server.AppURL, authServerMeta["issuer"])
	assert.Equal(t, server.AppURL+"/oauth/token", authServerMeta["token_endpoint"])

	response, err = http.Get(server.AppURL + "/.well-known/oauth-protected-resource")
	require.NoError(t, err)
	var protectedResourceMeta map[string]any
	require.NoError(t, harness.Decode(response, &protectedResourceMeta))
	assert.Equal(t, server.AppURL, protectedResourceMeta["resource"])
}

func TestWelcomeRoutes(t *testing.T) {
	server := harness.StartServer(t)

	response, err := http.Get(server.AppURL + "/")
	require.NoError(t, err)
	var app map[string]any
	require.NoError(t, harness.Decode(response, &app))
	assert.Contains(t, app["message"], "MCP server")

	response, err = http.Get(server.ControlURL + "/health")
	require.NoError(t, err)
	var health map[string]any
	require.NoError(t, harness.Decode(response, &health))
	assert.Equal(t, "OK", health["status"])
}
