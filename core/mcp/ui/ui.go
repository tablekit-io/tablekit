// Package ui serves the compiled MCP Apps UI templates that @tablekit/widgets
// builds. The vite build emits one content-hashed single-file HTML per template
// plus a manifest.json mapping template name -> {file, hash, bytes}; that
// widgets tree is bind-mounted into this package directory (see docker-compose)
// and embedded into the binary here. Because the URI carries the content hash,
// any widget change yields a new ui://tablekit/<name>-<hash> URI and auto-busts
// the host's per-URI resource cache.
package ui

import (
	"embed"
	"encoding/json"
)

// The build output lives next to this file. Only a committed .gitkeep keeps the
// directory present so this compiles before the first widget build; the manifest
// and content-hashed HTML the vite build emits are gitignored. The all: prefix
// is required because the directory form of //go:embed excludes dotfiles, so a
// tree holding only .gitkeep would otherwise match nothing. readManifest is
// fail-soft, so a binary built before the first widget build still boots.
//
//go:embed all:widgets
var widgets embed.FS

// MIMEType is the MCP Apps content type: a self-contained HTML document the host
// renders in a sandboxed iframe.
const MIMEType = "text/html;profile=mcp-app"

// manifestEntry is one template's build record, as written by the widgets build.
type manifestEntry struct {
	File  string `json:"file"`
	Hash  string `json:"hash"`
	Bytes int64  `json:"bytes"`
}

// UIResource is a registerable MCP Apps UI resource: HTML served under a ui://
// URI keyed by content hash.
type UIResource struct {
	Name     string
	URI      string
	MIMEType string
	HTML     string
}

// readManifest parses widgets/manifest.json from the embedded FS. Fail-soft: a
// tree without a (valid) manifest yields an empty map rather than an error, so a
// binary built before the first widget build still boots — it simply advertises
// no widgets until rebuilt.
func readManifest() map[string]manifestEntry {
	raw, err := widgets.ReadFile("widgets/manifest.json")
	if err != nil {
		return map[string]manifestEntry{}
	}
	var m map[string]manifestEntry
	if err := json.Unmarshal(raw, &m); err != nil {
		return map[string]manifestEntry{}
	}
	return m
}

// WidgetURI returns the content-hashed ui:// URI for a template, or "" if it
// isn't built. Tools link their widget via this URI in _meta.ui.resourceUri.
func WidgetURI(name string) string {
	entry, ok := readManifest()[name]
	if !ok {
		return ""
	}
	return "ui://tablekit/" + name + "-" + entry.Hash
}

// Resources returns every built template as a UIResource, ready to register on
// the MCP server. Reads each HTML file once; call it once at server build, not
// per request.
func Resources() []UIResource {
	manifest := readManifest()
	resources := make([]UIResource, 0, len(manifest))
	for name, entry := range manifest {
		html, err := widgets.ReadFile("widgets/" + entry.File)
		if err != nil {
			continue
		}
		resources = append(resources, UIResource{
			Name:     name,
			URI:      "ui://tablekit/" + name + "-" + entry.Hash,
			MIMEType: MIMEType,
			HTML:     string(html),
		})
	}
	return resources
}
