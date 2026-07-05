package mysql

import (
	"math"
	"strconv"
	"strings"

	"core/engine/config"
	"core/engine/encoding"
)

// pageQuery applies the page window to a user query in MySQL/MariaDB's dialect. It
// wraps the query in a subquery with LIMIT/OFFSET, over-fetching one row
// (limit+1) so trimToLimit can tell whether more rows remain without a second
// round-trip. A zero/negative Limit means "no window": the query is returned
// unwrapped (only the byte cap applies). Trailing whitespace and semicolons are
// stripped so the query is a valid subquery body.
//
// Note this wraps rather than injecting LIMIT into the user's own SQL: MySQL and
// MariaDB do not materialize the full inner result — the executor pipelines rows
// through the derived table and stops the scan once the outer LIMIT is satisfied
// — so the over-fetch reads at most limit+1 rows, not the whole table. (An inner
// ORDER BY without a matching index is the exception: the sort must consume all
// rows before the outer LIMIT can take the top ones.)
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
