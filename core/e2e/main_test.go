// Package e2e drives the real compiled `tablekit` binary over HTTP. TestMain
// builds the binary once; each test spawns its own server on random free ports
// with a random, auto-cleaned DATA_DIR (t.TempDir), so tests are isolated.
//
// Note: this package shells out to `go build`, so it needs the Go toolchain and
// localhost sockets available (no external network).
package e2e

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// binPath is the compiled test binary, built once in TestMain.
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tablekit-e2e-bin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "mktemp:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binPath = filepath.Join(dir, "tablekit")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = moduleRoot()
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "go build:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// moduleRoot returns core/ — the parent dir of this test file (core/e2e).
func moduleRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file))
}

type server struct {
	appURL     string
	controlURL string
	dataDir    string
}

// startServer spawns the binary on random free ports with a fresh data dir and
// waits until it is healthy. The process is killed and the data dir removed on
// test cleanup.
func startServer(t *testing.T) server {
	return startServerEnv(t)
}

// startServerEnv is startServer with extra "KEY=value" environment entries.
func startServerEnv(t *testing.T, extraEnv ...string) server {
	t.Helper()
	appPort, controlPort := freePort(t), freePort(t)
	dataDir := t.TempDir()
	appURL := fmt.Sprintf("http://127.0.0.1:%d", appPort)
	controlURL := fmt.Sprintf("http://127.0.0.1:%d", controlPort)

	cmd := exec.Command(binPath, "serve")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("APP_PORT=%d", appPort),
		fmt.Sprintf("CONTROL_PORT=%d", controlPort),
		"DATA_DIR="+dataDir,
		"PUBLIC_BASE_URL="+appURL,
	)
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	waitHealthy(t, controlURL)
	return server{appURL: appURL, controlURL: controlURL, dataDir: dataDir}
}

// freePort asks the OS for an unused TCP port, then releases it.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func waitHealthy(t *testing.T, controlURL string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(controlURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server did not become healthy at %s", controlURL)
}

// ---- OAuth helpers ------------------------------------------------------

const testRedirectURI = "http://127.0.0.1:9999/callback"

// register performs RFC 7591 dynamic client registration, returning client_id.
func register(t *testing.T, appURL string) string {
	t.Helper()
	body := `{"redirect_uris":["` + testRedirectURI + `"]}`
	resp, err := http.Post(appURL+"/register", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var out struct {
		ClientID string `json:"client_id"`
	}
	require.NoError(t, decode(resp, &out))
	require.NotEmpty(t, out.ClientID)
	return out.ClientID
}

// pkcePair returns a random verifier and its S256 challenge.
func pkcePair(t *testing.T) (verifier, challenge string) {
	t.Helper()
	b := make([]byte, 32)
	_, err := rand.Read(b)
	require.NoError(t, err)
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge
}

// authorize hits /authorize without following redirects, returning the raw
// response so callers can inspect the Location (code) or the already-paired page.
func authorize(t *testing.T, appURL, clientID, challenge, state string) *http.Response {
	t.Helper()
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", testRedirectURI)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if state != "" {
		q.Set("state", state)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(appURL + "/oauth/authorize?" + q.Encode())
	require.NoError(t, err)
	return resp
}

// authorizeCode runs authorize and extracts the code from the 302 Location.
func authorizeCode(t *testing.T, appURL, clientID, challenge string) string {
	t.Helper()
	resp := authorize(t, appURL, clientID, challenge, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusFound, resp.StatusCode)
	loc, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	code := loc.Query().Get("code")
	require.NotEmpty(t, code)
	return code
}

// postForm POSTs urlencoded values, returning status and parsed JSON body.
func postForm(t *testing.T, endpoint string, form url.Values) (int, map[string]any) {
	t.Helper()
	resp, err := http.PostForm(endpoint, form)
	require.NoError(t, err)
	defer resp.Body.Close()
	var body map[string]any
	_ = decode(resp, &body)
	return resp.StatusCode, body
}

// exchangeCode redeems an authorization code for tokens.
func exchangeCode(t *testing.T, appURL, clientID, code, verifier string) map[string]any {
	t.Helper()
	status, body := postForm(t, appURL+"/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {testRedirectURI},
	})
	require.Equal(t, http.StatusOK, status, "token exchange failed: %v", body)
	return body
}

// fullHandshake registers, authorizes and exchanges, returning the token body.
func fullHandshake(t *testing.T, appURL string) (clientID string, tokens map[string]any) {
	t.Helper()
	clientID = register(t, appURL)
	verifier, challenge := pkcePair(t)
	code := authorizeCode(t, appURL, clientID, challenge)
	return clientID, exchangeCode(t, appURL, clientID, code, verifier)
}

// bearerClient returns an HTTP client that injects a bearer token on every request.
func bearerClient(token string) *http.Client {
	return &http.Client{Transport: bearerRT{token: token, base: http.DefaultTransport}}
}

type bearerRT struct {
	token string
	base  http.RoundTripper
}

func (b bearerRT) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+b.token)
	return b.base.RoundTrip(req)
}

// runCLI runs the binary with a subcommand against the given data dir.
func runCLI(t *testing.T, dataDir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), binPath, args...)
	cmd.Env = append(os.Environ(), "DATA_DIR="+dataDir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "cli %v failed: %s", args, out)
}

func decode(resp *http.Response, v any) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
