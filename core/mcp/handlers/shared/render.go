package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v7"
)

// MustTemplate compiles a tool's embedded output template (its output.tmpl, the
// canonical structured result rendered as faithful text) and panics on a parse
// error. Tools call it once at package-init, so a malformed template surfaces
// the moment the server is constructed — in tests too — never at request time.
// This mirrors InputSchema's fail-at-startup guarantee for schema.json.
func MustTemplate(raw []byte) *pongo2.Template {
	tmpl, err := pongo2.FromBytes(raw)
	if err != nil {
		panic(fmt.Errorf("shared.MustTemplate: parse template: %w", err))
	}
	return tmpl
}

// RenderText renders a tool's canonical structured output to the text content
// block clients show the model. The structured output is the single source of
// truth: out is marshaled to JSON and decoded back into a generic map so the
// template addresses every field by its json tag name (the exact canonical
// shape), then the template executes against it.
//
// Numbers are decoded with UseNumber so no int64 is flattened to a lossy
// float64 on the round-trip — the text rendering stays faithful to the wire
// value. Trailing whitespace (from template control lines) is trimmed.
func RenderText[Out any](tmpl *pongo2.Template, out Out) (string, error) {
	raw, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("shared.RenderText: marshal output: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var context map[string]any
	if err := decoder.Decode(&context); err != nil {
		return "", fmt.Errorf("shared.RenderText: decode output: %w", err)
	}
	rendered, err := tmpl.Execute(pongo2.Context(context))
	if err != nil {
		return "", fmt.Errorf("shared.RenderText: execute template: %w", err)
	}
	return strings.TrimRight(rendered, "\n \t"), nil
}

func init() {
	pongo2.RegisterFilter("mdtable", filterMarkdownTable)
}

// tableColumn is one resolved column: the header label and the key used to pull
// each row's cell value.
type tableColumn struct {
	label string
	key   string
}

// filterMarkdownTable renders a list of row maps as a GitHub-flavored markdown
// table — the generic tabular renderer every tool template shares via
// `{{ rows|mdtable:columns }}`. It normalizes the two column shapes the tools
// use: objects `{"name": ...}` (query_database's ColumnInfo) and plain strings
// (read_results' column names). Cells are rendered losslessly through
// markdownCell. The result is a safe value so its structural `|` and `<br>`
// survive regardless of the template's autoescape setting.
func filterMarkdownTable(rows *pongo2.Value, columns *pongo2.Value) (*pongo2.Value, error) {
	resolved := resolveColumns(columns)

	var builder strings.Builder
	builder.WriteString("|")
	for _, column := range resolved {
		builder.WriteString(" " + markdownCell(column.label) + " |")
	}
	builder.WriteString("\n|")
	for range resolved {
		builder.WriteString(" --- |")
	}

	rows.Iterate(func(_, _ int, row, _ *pongo2.Value) bool {
		values, _ := row.Interface().(map[string]any)
		builder.WriteString("\n|")
		for _, column := range resolved {
			builder.WriteString(" " + markdownCell(values[column.key]) + " |")
		}
		return true
	}, func() {})

	return pongo2.AsSafeValue(builder.String()), nil
}

// resolveColumns turns a template's `columns` value into header/key pairs,
// accepting either object columns (`{"name": ...}`) or bare string columns.
func resolveColumns(columns *pongo2.Value) []tableColumn {
	var resolved []tableColumn
	columns.Iterate(func(_, _ int, column, _ *pongo2.Value) bool {
		switch value := column.Interface().(type) {
		case map[string]any:
			name, _ := value["name"].(string)
			resolved = append(resolved, tableColumn{label: name, key: name})
		case string:
			resolved = append(resolved, tableColumn{label: value, key: value})
		default:
			name := column.String()
			resolved = append(resolved, tableColumn{label: name, key: name})
		}
		return true
	}, func() {})
	return resolved
}

// markdownCell renders one value as a markdown table cell without losing
// information: nil becomes empty, pipes are escaped so they can't break the
// column layout, and newlines become <br> so a multi-line value stays on its
// row. Numbers arrive as json.Number (see RenderText's UseNumber), so their
// exact textual form is preserved rather than reformatted through a float.
func markdownCell(value any) string {
	var text string
	switch typed := value.(type) {
	case nil:
		return ""
	case json.Number:
		text = typed.String()
	case string:
		text = typed
	case bool:
		text = strconv.FormatBool(typed)
	default:
		text = fmt.Sprintf("%v", typed)
	}
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", "<br>")
	return text
}
