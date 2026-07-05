// Command gen regenerates the go-jet table/model code for tablekit's own state
// schema. It introspects a LIVE, already-migrated database (go-jet cannot read
// migration SQL), so point it at a Postgres that has the current migrations
// applied — by default the dev `postgres` service.
//
// Run manually after changing a migration:
//
//	# apply the migration to the dev postgres first (restart core), then:
//	cd core && go run ./db/gen        # or JET_DSN=... go run ./db/gen
//
// The generated code under db/gen/tablekit/public is committed.
package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/postgres"
	"github.com/go-jet/jet/v2/generator/template"
	postgresdialect "github.com/go-jet/jet/v2/postgres"

	"core/db/dbjson"
)

// defaultDSN targets the dev `postgres` service on its host-published port.
const defaultDSN = "postgres://postgres:pw@localhost:5433/tablekit?sslmode=disable"

func main() {
	dsn := os.Getenv("JET_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	connection, err := parseDSN(dsn)
	if err != nil {
		log.Fatalf("parse JET_DSN: %v", err)
	}

	// Override the two jsonb columns to precise Go types instead of raw bytes:
	// redirect_uris has a fixed shape ([]string); config.value is a polymorphic
	// key/value store, so it stays generic JSON (json.RawMessage), unmarshaled
	// per key by the caller.
	typed := template.Default(postgresdialect.Dialect).
		UseSchema(func(schema metadata.Schema) template.Schema {
			return template.DefaultSchema(schema).
				UseModel(template.DefaultModel().
					UseTable(func(table metadata.Table) template.TableModel {
						return template.DefaultTableModel(table).
							UseField(func(column metadata.Column) template.TableModelField {
								field := template.DefaultTableModelField(column)
								switch {
								case table.Name == "oauth_clients" && column.Name == "redirect_uris":
									field.Type = template.NewType(dbjson.JSON[[]string]{})
								case table.Name == "config" && column.Name == "value":
									field.Type = template.NewType(json.RawMessage{})
								}
								return field
							})
					}),
				)
		})

	if err := postgres.Generate("./db/gen", connection, typed); err != nil {
		log.Fatalf("generate: %v", err)
	}
	log.Println("generated db/gen/tablekit/public")
}

// parseDSN turns a postgres:// URL into the discrete connection fields the
// generator wants.
func parseDSN(dsn string) (postgres.DBConnection, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return postgres.DBConnection{}, err
	}
	port := 5432
	if raw := parsed.Port(); raw != "" {
		if port, err = strconv.Atoi(raw); err != nil {
			return postgres.DBConnection{}, err
		}
	}
	password, _ := parsed.User.Password()
	sslMode := parsed.Query().Get("sslmode")
	if sslMode == "" {
		sslMode = "disable"
	}
	return postgres.DBConnection{
		Host:       parsed.Hostname(),
		Port:       port,
		User:       parsed.User.Username(),
		Password:   password,
		DBName:     strings.TrimPrefix(parsed.Path, "/"),
		SchemaName: "public",
		SslMode:    sslMode,
	}, nil
}
