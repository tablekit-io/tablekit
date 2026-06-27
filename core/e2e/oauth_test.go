package e2e

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthHappyPath(t *testing.T) {
	server := startServer(t)
	_, tokens := fullHandshake(t, server.appURL)

	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
	assert.Equal(t, "Bearer", tokens["token_type"])
	assert.Equal(t, "mcp", tokens["scope"])
}

func TestRefreshRotationAndReplay(t *testing.T) {
	server := startServer(t)
	clientID, tokens := fullHandshake(t, server.appURL)
	oldRefresh := tokens["refresh_token"].(string)

	// Rotation: old refresh yields a new pair.
	status, body := postForm(t, server.appURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {oldRefresh},
	})
	require.Equal(t, http.StatusOK, status, body)
	newRefresh := body["refresh_token"].(string)
	assert.NotEqual(t, oldRefresh, newRefresh)

	// Replay: reusing the old refresh is rejected as theft.
	status, body = postForm(t, server.appURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {oldRefresh},
	})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", body["error"])

	// And the whole chain is now revoked: the rotated refresh fails too.
	status, body = postForm(t, server.appURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {newRefresh},
	})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", body["error"])
}

func TestMetadataEndpoints(t *testing.T) {
	server := startServer(t)

	response, err := http.Get(server.appURL + "/.well-known/oauth-authorization-server")
	require.NoError(t, err)
	var authServerMeta map[string]any
	require.NoError(t, decode(response, &authServerMeta))
	assert.Equal(t, server.appURL, authServerMeta["issuer"])
	assert.Equal(t, server.appURL+"/oauth/token", authServerMeta["token_endpoint"])

	response, err = http.Get(server.appURL + "/.well-known/oauth-protected-resource")
	require.NoError(t, err)
	var protectedResourceMeta map[string]any
	require.NoError(t, decode(response, &protectedResourceMeta))
	assert.Equal(t, server.appURL, protectedResourceMeta["resource"])
}

func TestMCPRequiresBearer(t *testing.T) {
	server := startServer(t)
	response, err := http.Post(server.appURL+"/mcp", "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	assert.Contains(t, response.Header.Get("WWW-Authenticate"), "resource_metadata")
}

func TestWelcomeRoutes(t *testing.T) {
	server := startServer(t)

	response, err := http.Get(server.appURL + "/")
	require.NoError(t, err)
	var app map[string]any
	require.NoError(t, decode(response, &app))
	assert.Contains(t, app["message"], "MCP server")

	response, err = http.Get(server.controlURL + "/health")
	require.NoError(t, err)
	var health map[string]any
	require.NoError(t, decode(response, &health))
	assert.Equal(t, "OK", health["status"])
}
