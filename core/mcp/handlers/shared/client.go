package shared

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClientID extracts the authenticated client's id (a UUID) from the request's
// bearer-token info, which the HTTP-layer verifier places in TokenInfo.Extra
// under "client_id". Every request that reaches a tool handler has passed the
// bearer gate, so it carries a client id; a missing/unparseable one is an error.
func ClientID(req *mcp.CallToolRequest) (uuid.UUID, error) {
	extra := req.GetExtra()
	if extra == nil || extra.TokenInfo == nil {
		return uuid.Nil, fmt.Errorf("request has no authenticated client")
	}
	raw, _ := extra.TokenInfo.Extra["client_id"].(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("request client id %q is not a uuid: %w", raw, err)
	}
	return id, nil
}
