package oauth

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templatesFS embed.FS

// alreadyPairedTmpl renders the "already paired" page. Parsed once at init;
// html/template applies context-aware escaping to the interpolated redirect URL.
var alreadyPairedTmpl = template.Must(
	template.ParseFS(templatesFS, "templates/already_paired.html"),
)
