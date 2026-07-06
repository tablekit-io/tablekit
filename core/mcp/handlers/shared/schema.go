package shared

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

// InputSchema parses a tool's hand-written JSON input schema (its schema.json,
// the source of truth for what the tool advertises) and guarantees it stays
// decode-compatible with In — the Go struct the SDK unmarshals arguments into.
// It panics at startup on a parse error or any struct↔schema drift, so an
// incompatible schema can never ship: the mismatch surfaces the moment the
// server is constructed (including in tests) rather than silently dropping a
// field at request time.
func InputSchema[In any](raw []byte) *jsonschema.Schema {
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		panic(fmt.Errorf("shared.InputSchema[%T]: parse schema: %w", *new(In), err))
	}
	if err := assertCompatible[In](&schema); err != nil {
		panic(fmt.Errorf("shared.InputSchema[%T]: %w", *new(In), err))
	}
	return &schema
}

// assertCompatible reflects In the same way mcp.AddTool would and checks the
// hand-written schema describes exactly the same shape: identical property-name
// sets at every object level, matching JSON types, and matching array-ness.
// Constraint-only keywords the struct can't express (enum, minimum, format,
// default, description, additionalProperties) are ignored — the schema may
// refine, it just may not diverge.
func assertCompatible[In any](have *jsonschema.Schema) error {
	want, err := jsonschema.For[In](nil)
	if err != nil {
		return fmt.Errorf("reflect %T: %w", *new(In), err)
	}
	return compatible("(root)", want, have)
}

func compatible(path string, want, have *jsonschema.Schema) error {
	if wantType, haveType := schemaType(want), schemaType(have); wantType != "" && haveType != "" && wantType != haveType {
		return fmt.Errorf("%s: schema declares type %q but struct is %q", path, haveType, wantType)
	}

	for name := range have.Properties {
		if _, ok := want.Properties[name]; !ok {
			return fmt.Errorf("%s.%s: property in schema but not in struct (args would be dropped on decode)", path, name)
		}
	}
	for name := range want.Properties {
		if _, ok := have.Properties[name]; !ok {
			return fmt.Errorf("%s.%s: property in struct but not advertised in schema", path, name)
		}
	}
	for name, haveProp := range have.Properties {
		if err := compatible(path+"."+name, want.Properties[name], haveProp); err != nil {
			return err
		}
	}

	if (want.Items == nil) != (have.Items == nil) {
		return fmt.Errorf("%s: array-items mismatch (struct items=%t, schema items=%t)", path, want.Items != nil, have.Items != nil)
	}
	if want.Items != nil && have.Items != nil {
		if err := compatible(path+"[]", want.Items, have.Items); err != nil {
			return err
		}
	}
	return nil
}

// schemaType returns a schema's single JSON type, tolerating the Types slice
// form jsonschema-go uses for nullable/multi-type schemas ("" when ambiguous).
func schemaType(s *jsonschema.Schema) string {
	if s.Type != "" {
		return s.Type
	}
	if len(s.Types) == 1 {
		return s.Types[0]
	}
	return ""
}
