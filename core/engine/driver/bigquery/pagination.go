package bigquery

import (
	"math"
	"strconv"
	"strings"

	"core/engine/config"
	"core/engine/encoding"
)

// pageQuery applies the page window to a user query in BigQuery Standard SQL. It
// wraps the query in a subquery with LIMIT/OFFSET, over-fetching one row
// (limit+1) so trimToLimit can tell whether more rows remain without a second
// job. A zero/negative Limit means "no window": the query is returned unwrapped
// (only the byte cap applies). Trailing whitespace and semicolons are stripped so
// the query is a valid subquery body.
//
// BigQuery Standard SQL only accepts OFFSET together with LIMIT, so the wrapper
// always emits both — matching the postgres driver's shape.
func pageQuery(query string, page config.Page) string {
	if page.Limit <= 0 {
		return query
	}
	trimmed := strings.TrimRight(strings.TrimSpace(query), "; \t\r\n")
	return "SELECT * FROM (" + trimmed + ") AS _tablekit_page" +
		" LIMIT " + strconv.Itoa(page.Limit+1) +
		" OFFSET " + strconv.Itoa(page.Offset)
}

// accumulatorRows is the row cap for the accumulator: the page window plus the
// one over-fetched row. An unbounded page (Limit <= 0) is capped only by bytes.
func accumulatorRows(page config.Page) int {
	if page.Limit <= 0 {
		return math.MaxInt
	}
	return page.Limit + 1
}

// trimToLimit drops the over-fetched extra row (if present) and reports whether
// it was there — i.e. whether more rows exist beyond this window. A result
// already cut short by the byte cap keeps its Truncated flag untouched. An
// unbounded page (limit <= 0) never trims.
func trimToLimit(result *encoding.Result, limit int) bool {
	if limit > 0 && result.RowCount > limit {
		result.Rows = result.Rows[:limit]
		result.RowCount = limit
		return true
	}
	return false
}
