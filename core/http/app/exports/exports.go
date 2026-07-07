// Package exports serves the signed download links get_export_url hands out. The
// endpoint is unauthenticated by design: access is proven solely by the
// short-lived export token in the path (minted by the JWT issuer), so it must be
// mounted OUTSIDE the /mcp bearer guard. On each request it verifies the token,
// loads the stored query it authorizes, re-runs it against the live database, and
// streams the result as CSV (download) or JSON (inline).
package exports

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"core/engine"
	"core/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers serves the export endpoint, wired to the issuer (token verification),
// the stored-query repository, and the engine (re-running the query).
type Handlers struct {
	services *services.Services
}

// NewHandlers builds the export handlers from the shared services.
func NewHandlers(appServices *services.Services) *Handlers {
	return &Handlers{services: appServices}
}

// Register mounts GET /exports/:format/:token on engine.
func (h *Handlers) Register(engine *gin.Engine) {
	engine.GET("/exports/:format/:token", h.handleExport)
}

// handleExport verifies the token, re-runs the stored query, and serializes the
// result in the requested format.
func (h *Handlers) handleExport(c *gin.Context) {
	format := c.Param("format")
	if format != "csv" && format != "json" {
		c.String(http.StatusNotFound, "unknown export format %q", format)
		return
	}

	claims, err := h.services.Issuer.VerifyExport(c.Param("token"))
	if err != nil {
		c.String(http.StatusUnauthorized, "this export link is invalid or has expired")
		return
	}

	queryID, err := uuid.Parse(claims.QK)
	if err != nil {
		c.String(http.StatusNotFound, "the query for this export no longer exists")
		return
	}
	descriptor, err := h.services.Queries.Get(c.Request.Context(), queryID)
	if err != nil {
		c.String(http.StatusInternalServerError, "could not load the query")
		return
	}
	if descriptor == nil {
		c.String(http.StatusNotFound, "the query for this export no longer exists")
		return
	}

	name, err := h.services.Databases.Verify(c.Request.Context(), descriptor.DatabaseID)
	if err != nil {
		c.String(http.StatusConflict, "%v", err)
		return
	}

	result, _, err := h.services.Engine.RunReadOnlyPage(
		c.Request.Context(), name, descriptor.SQL,
		engine.PageOptions{Limit: chartMaxRows, MaxBytes: chartMaxBytes},
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "could not run the query: %v", err)
		return
	}
	if result.Truncated {
		c.String(http.StatusRequestEntityTooLarge, "the result is too large to export")
		return
	}

	switch format {
	case "csv":
		writeCSV(c, descriptor.ID.String(), result)
	case "json":
		writeJSON(c, result)
	}
}

// chart/export caps mirror the MCP chart tools: full result up to these bounds.
const (
	chartMaxRows  = 100_000
	chartMaxBytes = 16 << 20 // 16 MiB
)

// writeCSV streams the result as an RFC 4180 CSV attachment (forces a download).
func writeCSV(c *gin.Context, queryID string, result *engine.Result) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", "export-"+queryID+".csv"))
	c.Status(http.StatusOK)

	writer := csv.NewWriter(c.Writer)
	_ = writer.Write(result.Columns)
	record := make([]string, len(result.Columns))
	for _, row := range result.Rows {
		for i, column := range result.Columns {
			record[i] = cellToString(row[column])
		}
		_ = writer.Write(record)
	}
	writer.Flush()
}

// writeJSON sends the rows as pretty-printed JSON inline (the browser renders it).
func writeJSON(c *gin.Context, result *engine.Result) {
	c.IndentedJSON(http.StatusOK, result.Rows)
}

// cellToString renders a normalized cell value for CSV. Values are already
// JSON-friendly scalars (strings, numbers, bools) or nil; nil becomes empty.
func cellToString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
