package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html
var templatesFS embed.FS

// AlreadyPairedTemplate renders the "already paired" page. Parsed once at init;
// html/template applies context-aware escaping to the interpolated redirect URL.
var AlreadyPairedTemplate = template.Must(
	template.ParseFS(templatesFS, "already_paired.html"),
)
