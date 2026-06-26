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
	srv := startServer(t)
	_, tokens := fullHandshake(t, srv.appURL)

	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
	assert.Equal(t, "Bearer", tokens["token_type"])
	assert.Equal(t, "mcp", tokens["scope"])
}

func TestRefreshRotationAndReplay(t *testing.T) {
	srv := startServer(t)
	clientID, tokens := fullHandshake(t, srv.appURL)
	oldRefresh := tokens["refresh_token"].(string)

	// Rotation: old refresh yields a new pair.
	status, body := postForm(t, srv.appURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {oldRefresh},
	})
	require.Equal(t, http.StatusOK, status, body)
	newRefresh := body["refresh_token"].(string)
	assert.NotEqual(t, oldRefresh, newRefresh)

	// Replay: reusing the old refresh is rejected as theft.
	status, body = postForm(t, srv.appURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {oldRefresh},
	})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", body["error"])

	// And the whole chain is now revoked: the rotated refresh fails too.
	status, body = postForm(t, srv.appURL+"/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {newRefresh},
	})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", body["error"])
}

func TestMetadataEndpoints(t *testing.T) {
	srv := startServer(t)

	resp, err := http.Get(srv.appURL + "/.well-known/oauth-authorization-server")
	require.NoError(t, err)
	var as map[string]any
	require.NoError(t, decode(resp, &as))
	assert.Equal(t, srv.appURL, as["issuer"])
	assert.Equal(t, srv.appURL+"/oauth/token", as["token_endpoint"])

	resp, err = http.Get(srv.appURL + "/.well-known/oauth-protected-resource")
	require.NoError(t, err)
	var pr map[string]any
	require.NoError(t, decode(resp, &pr))
	assert.Equal(t, srv.appURL, pr["resource"])
}

func TestMCPRequiresBearer(t *testing.T) {
	srv := startServer(t)
	resp, err := http.Post(srv.appURL+"/mcp", "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("WWW-Authenticate"), "resource_metadata")
}

func TestWelcomeRoutes(t *testing.T) {
	srv := startServer(t)

	resp, err := http.Get(srv.appURL + "/")
	require.NoError(t, err)
	var app map[string]any
	require.NoError(t, decode(resp, &app))
	assert.Contains(t, app["message"], "MCP server")

	resp, err = http.Get(srv.controlURL + "/health")
	require.NoError(t, err)
	var health map[string]any
	require.NoError(t, decode(resp, &health))
	assert.Equal(t, "OK", health["status"])
}
