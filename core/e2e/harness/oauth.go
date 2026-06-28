package harness

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRedirectURI is the redirect_uri used by the e2e OAuth helpers.
const TestRedirectURI = "http://127.0.0.1:9999/callback"

// Register performs RFC 7591 dynamic client registration, returning client_id.
func Register(t *testing.T, appURL string) string {
	t.Helper()
	body := `{"redirect_uris":["` + TestRedirectURI + `"]}`
	response, err := http.Post(appURL+"/register", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	var out struct {
		ClientID string `json:"client_id"`
	}
	require.NoError(t, Decode(response, &out))
	require.NotEmpty(t, out.ClientID)
	return out.ClientID
}

// PKCEPair returns a random verifier and its S256 challenge.
func PKCEPair(t *testing.T) (verifier, challenge string) {
	t.Helper()
	b := make([]byte, 32)
	_, err := rand.Read(b)
	require.NoError(t, err)
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge
}

// Authorize hits /authorize without following redirects, returning the raw
// response so callers can inspect the Location (code) or the already-paired page.
func Authorize(t *testing.T, appURL, clientID, challenge, state string) *http.Response {
	t.Helper()
	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", clientID)
	query.Set("redirect_uri", TestRedirectURI)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	if state != "" {
		query.Set("state", state)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	response, err := client.Get(appURL + "/oauth/authorize?" + query.Encode())
	require.NoError(t, err)
	return response
}

// AuthorizeCode runs Authorize and extracts the code from the 302 Location.
func AuthorizeCode(t *testing.T, appURL, clientID, challenge string) string {
	t.Helper()
	response := Authorize(t, appURL, clientID, challenge, "")
	defer response.Body.Close()
	require.Equal(t, http.StatusFound, response.StatusCode)
	location, err := url.Parse(response.Header.Get("Location"))
	require.NoError(t, err)
	code := location.Query().Get("code")
	require.NotEmpty(t, code)
	return code
}

// PostForm POSTs urlencoded values, returning status and parsed JSON body.
func PostForm(t *testing.T, endpoint string, form url.Values) (int, map[string]any) {
	t.Helper()
	response, err := http.PostForm(endpoint, form)
	require.NoError(t, err)
	defer response.Body.Close()
	var body map[string]any
	_ = Decode(response, &body)
	return response.StatusCode, body
}

// ExchangeCode redeems an authorization code for tokens.
func ExchangeCode(t *testing.T, appURL, clientID, code, verifier string) map[string]any {
	t.Helper()
	status, body := PostForm(t, appURL+"/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {TestRedirectURI},
	})
	require.Equal(t, http.StatusOK, status, "token exchange failed: %v", body)
	return body
}

// FullHandshake registers, authorizes and exchanges, returning the token body.
func FullHandshake(t *testing.T, appURL string) (clientID string, tokens map[string]any) {
	t.Helper()
	clientID = Register(t, appURL)
	verifier, challenge := PKCEPair(t)
	code := AuthorizeCode(t, appURL, clientID, challenge)
	return clientID, ExchangeCode(t, appURL, clientID, code, verifier)
}
