package store

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed schemas/clients.schema.json schemas/tokens.schema.json
var schemaFS embed.FS

// schemas maps a state file name to its resolved JSON Schema. State files not
// present here (e.g. signing.key, which is not JSON) are read without
// validation.
var schemas = map[string]*jsonschema.Resolved{
	"clients.json": mustResolve("schemas/clients.schema.json"),
	"tokens.json":  mustResolve("schemas/tokens.schema.json"),
}

// mustResolve loads and resolves an embedded schema at package init. A broken
// schema is a programming error, so it panics rather than failing silently.
func mustResolve(path string) *jsonschema.Resolved {
	raw, err := schemaFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("store: reading embedded schema %s: %v", path, err))
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		panic(fmt.Sprintf("store: parsing schema %s: %v", path, err))
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		panic(fmt.Sprintf("store: resolving schema %s: %v", path, err))
	}
	return resolved
}

// validateAgainstSchema validates raw JSON for the named state file against its
// schema. Files without a registered schema pass through unchecked.
func validateAgainstSchema(name string, raw []byte) error {
	resolved := schemas[name]
	if resolved == nil {
		return nil
	}
	var instance any
	if err := json.Unmarshal(raw, &instance); err != nil {
		return fmt.Errorf("%s: invalid JSON: %w", name, err)
	}
	if err := resolved.Validate(instance); err != nil {
		return fmt.Errorf("%s: schema validation failed: %w", name, err)
	}
	return nil
}
