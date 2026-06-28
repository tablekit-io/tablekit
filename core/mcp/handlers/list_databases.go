package handlers

import (
	"context"
	"fmt"

	"core/engine"
	"core/helpers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// listDatabasesInput is empty: the tool takes no arguments.
type listDatabasesInput struct{}

// listDatabasesOutput is the set of configured databases, name and type only.
type listDatabasesOutput struct {
	Databases []engine.DatabaseInfo `json:"databases" jsonschema:"the configured databases, each with its name and engine type"`
}

// listDatabases returns the databases available to run_sql. It never exposes
// hosts, credentials or any connection detail.
func (h *Handlers) listDatabases(_ context.Context, _ *mcp.CallToolRequest, _ listDatabasesInput) (*mcp.CallToolResult, listDatabasesOutput, error) {
	databases := h.Engine.List()
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("%d database(s) configured.", len(databases)),
		}},
	}, listDatabasesOutput{Databases: databases}, nil
}

// registerListDatabases adds the list_databases tool.
func (h *Handlers) registerListDatabases(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_databases",
		Description: "Lists the databases available to run_sql, each with its name and engine type (postgres/mysql/mariadb). Returns no credentials or connection details.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			IdempotentHint:  true,
			DestructiveHint: helpers.Pointer(false),
			OpenWorldHint:   helpers.Pointer(false),
		},
	}, h.listDatabases)
}
