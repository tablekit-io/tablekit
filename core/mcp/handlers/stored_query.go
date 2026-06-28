package handlers

import (
	"core/engine"
)

// Chart/export sizing: the render and export paths fetch the whole result, so
// they raise the row and byte caps well above the run_sql/run_query page size.
const (
	chartMaxRows  = 100_000
	chartMaxBytes = 16 << 20 // 16 MiB
)

// enginePage builds engine.PageOptions; a zero maxBytes lets the engine apply
// its default.
func enginePage(offset, limit, maxBytes int) engine.PageOptions {
	return engine.PageOptions{Offset: offset, Limit: limit, MaxBytes: maxBytes}
}

// toColumnInfos wraps column names as the structured columnInfo list tools return.
func toColumnInfos(columns []string) []columnInfo {
	infos := make([]columnInfo, len(columns))
	for i, name := range columns {
		infos[i] = columnInfo{Name: name}
	}
	return infos
}

// moreSuffix renders a short " (more rows available)" note for tool summaries.
func moreSuffix(hasMore bool) string {
	if hasMore {
		return " (more rows available)"
	}
	return ""
}

// projectColumns narrows a result to the requested columns, preserving the
// requested order and silently dropping names the result doesn't have. An empty
// request returns the columns and rows unchanged. It does not mutate the input.
func projectColumns(columns []string, rows []map[string]any, requested []string) (projectedColumns []string, projectedRows []map[string]any) {
	if len(requested) == 0 {
		return columns, rows
	}
	present := make(map[string]bool, len(columns))
	for _, name := range columns {
		present[name] = true
	}
	kept := make([]string, 0, len(requested))
	for _, name := range requested {
		if present[name] {
			kept = append(kept, name)
		}
	}
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		narrowed := make(map[string]any, len(kept))
		for _, name := range kept {
			if value, ok := row[name]; ok {
				narrowed[name] = value
			}
		}
		out[i] = narrowed
	}
	return kept, out
}
