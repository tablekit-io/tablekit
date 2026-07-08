package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---- RNG + rendering helpers ------------------------------------------------

func (g *gen) uuid() string {
	var u [16]byte
	g.rng.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

func (g *gen) intn(n int) int      { return g.rng.Intn(n) }
func (g *gen) chance(p float64) bool { return g.rng.Float64() < p }

// pick returns a uniformly random element.
func pick[T any](g *gen, xs []T) T { return xs[g.rng.Intn(len(xs))] }

// weightedIndex returns an index chosen by the given weights.
func (g *gen) weightedIndex(weights []float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := g.rng.Float64() * total
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(weights) - 1
}

// s renders a SQL string literal (single quotes doubled).
func s(v string) string { return "'" + strings.ReplaceAll(v, "'", "''") + "'" }

// ns renders a nullable string: empty -> NULL.
func ns(v string) string {
	if v == "" {
		return "NULL"
	}
	return s(v)
}

// money renders a numeric with two decimals.
func money(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }

func b2(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func jsonbArray(items []string) string {
	if len(items) == 0 {
		return "'[]'::jsonb"
	}
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = `"` + strings.ReplaceAll(it, `"`, `\"`) + `"`
	}
	return s("["+strings.Join(quoted, ", ")+"]") + "::jsonb"
}

// rel renders a load-relative timestamp: now() - interval '<age>'. A non-positive
// age clamps to now().
func rel(age time.Duration) string {
	if age < 0 {
		age = 0
	}
	total := int64(age / time.Second)
	days := total / 86400
	rem := total % 86400
	h := rem / 3600
	m := (rem % 3600) / 60
	sec := rem % 60
	return fmt.Sprintf("now() - interval '%d days %d hours %d mins %d secs'", days, h, m, sec)
}

// slug turns a name into a URL slug.
func slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
