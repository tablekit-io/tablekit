package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"core/services/requests"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// loggingMiddleware records every MCP request in the mcp_requests audit log. It
// is receiving middleware, so it wraps the whole dispatch and sees the method
// name, the request, and the handler's result/error — including the app-only
// bridge tools (fetch_chart_data, get_export_url) the agent never sees. Logging
// is best-effort: a failure to write the audit row is logged and swallowed so it
// never breaks the real MCP request, whose (result, err) is always returned
// unchanged.
func loggingMiddleware(logRecorder requests.RequestLog) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			result, err := next(ctx, method, req)

			entry := requests.Entry{
				Method:     method,
				ToolName:   toolNameOf(req),
				ClientID:   clientIDOf(req),
				Params:     marshalOrNil(req),
				DurationMS: int(time.Since(start).Milliseconds()),
			}
			if err != nil {
				entry.Error = marshalError(err)
			} else {
				entry.Result = marshalOrNil(result)
			}
			if logErr := logRecorder.Log(ctx, entry); logErr != nil {
				log.Error().Str("method", method).Err(logErr).Msg("mcp request audit log failed")
			}
			return result, err
		}
	}
}

// toolNameOf returns the invoked tool name for a tools/call request, or "" for
// any other method.
func toolNameOf(req mcp.Request) string {
	if call, ok := req.(*mcp.CallToolRequest); ok && call.Params != nil {
		return call.Params.Name
	}
	return ""
}

// clientIDOf returns the OAuth client id the request's bearer token was issued
// to (set by the HTTP-layer verifier in TokenInfo.Extra), or "" when absent.
func clientIDOf(req mcp.Request) string {
	extra := req.GetExtra()
	if extra == nil || extra.TokenInfo == nil {
		return ""
	}
	if clientID, ok := extra.TokenInfo.Extra["client_id"].(string); ok {
		return clientID
	}
	return ""
}

// marshalOrNil JSON-encodes a request/result, returning nil (stored as SQL NULL)
// for a nil value or a marshal failure — the audit row is never worth failing on.
func marshalOrNil(value any) []byte {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return encoded
}

// marshalError renders a handler error as structured JSON: a JSON-RPC error's
// {code, message, data} when the error is one, otherwise {"message": ...}.
func marshalError(err error) []byte {
	var wire *jsonrpc.Error
	if errors.As(err, &wire) {
		if encoded, mErr := json.Marshal(wire); mErr == nil {
			return encoded
		}
	}
	encoded, mErr := json.Marshal(map[string]string{"message": err.Error()})
	if mErr != nil {
		return nil
	}
	return encoded
}
