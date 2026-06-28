package harness

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Connect opens an MCP client session over Streamable HTTP to the server's /mcp
// endpoint, using the given HTTP client (e.g. one that injects a bearer token).
func Connect(t *testing.T, appURL string, client *http.Client) (*mcp.ClientSession, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	c := mcp.NewClient(&mcp.Implementation{Name: "e2e", Version: "0"}, nil)
	return c.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:   appURL + "/mcp",
		HTTPClient: client,
	}, nil)
}
