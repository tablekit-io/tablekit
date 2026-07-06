package shared

import "core/engine"

// Pagination sizing shared by the stored-query tools. DefaultLimit is the window
// query_database previews and read_results uses when no limit is given; MaxLimit
// caps a single read_results window.
const (
	DefaultLimit = 128
	MaxLimit     = 2048
)

// Chart/export sizing: the render and export paths fetch the whole result, so
// they raise the row and byte caps well above the query_database page size.
const (
	ChartMaxRows  = 100_000
	ChartMaxBytes = 16 << 20 // 16 MiB
)

// EnginePage builds engine.PageOptions; a zero maxBytes lets the engine apply
// its default.
func EnginePage(offset, limit, maxBytes int) engine.PageOptions {
	return engine.PageOptions{Offset: offset, Limit: limit, MaxBytes: maxBytes}
}

// ToColumnInfos wraps column names as the structured ColumnInfo list tools return.
func ToColumnInfos(columns []string) []ColumnInfo {
	infos := make([]ColumnInfo, len(columns))
	for i, name := range columns {
		infos[i] = ColumnInfo{Name: name}
	}
	return infos
}

// MoreSuffix renders a short " (more rows available)" note for tool summaries.
func MoreSuffix(hasMore bool) string {
	if hasMore {
		return " (more rows available)"
	}
	return ""
}

// ProjectColumns narrows a result to the requested columns, preserving the
// requested order and silently dropping names the result doesn't have. An empty
// request returns the columns and rows unchanged. It does not mutate the input.
func ProjectColumns(columns []string, rows []map[string]any, requested []string) (projectedColumns []string, projectedRows []map[string]any) {
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
