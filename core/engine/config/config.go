// Package config loads and validates the databases YAML and resolves it into
// the immutable, secret-free value types the rest of engine consumes. It is the
// shared data vocabulary: drivers, transport and the engine facade all speak in
// terms of Database, Limits and friends, but only this package knows the YAML
// shape, the JSON Schema and where secrets come from.
package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/databases.schema.json
var schemaJSON []byte

// Limits bound a single query: a server-side statement timeout plus caps on the
// number of rows and the encoded byte size of the result. The first cap reached
// truncates the result and sets Result.Truncated.
type Limits struct {
	StatementTimeout time.Duration
	MaxRows          int
	MaxBytes         int
}

// WithDefaults returns the limits with any zero field filled with its default.
func (l Limits) WithDefaults() Limits {
	if l.StatementTimeout <= 0 {
		l.StatementTimeout = 10 * time.Second
	}
	if l.MaxRows <= 0 {
		l.MaxRows = 2048
	}
	if l.MaxBytes <= 0 {
		l.MaxBytes = 64 * 1024
	}
	return l
}

// Page is a single paginated window over a query. A zero Limit means "no window"
// (drivers apply only the service caps). Drivers own the SQL dialect that realizes
// the window, so a non-standard engine can page its own way.
type Page struct {
	Offset int
	Limit  int
}

// DatabaseType is the engine discriminator used to route to an implementation.
type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeMySQL    DatabaseType = "mysql"
	DatabaseTypeMariaDB  DatabaseType = "mariadb"
	DatabaseTypeBigQuery DatabaseType = "bigquery"
)

// Database is a resolved, ready-to-dial definition: secrets are plaintext and
// per-engine defaults are filled in. This is what the engine implementations
// consume; the raw YAML types never cross that boundary.
type Database struct {
	Name             string
	Type             DatabaseType
	ConnectionString string // empty when details-based
	Details          *Details
	BigQuery         *BigQueryDetails // set only for the bigquery engine
	TLS              *TLSSettings
	SSH              *SSHSettings
}

type Details struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// BigQueryDetails is the resolved connection for a BigQuery database. It has no
// host/port/password: BigQuery is an HTTP job API authenticated with a Google
// service-account key file, so this struct is kept separate from the OLTP
// Details. Details is nil for a bigquery database and BigQuery is nil for every
// other engine.
type BigQueryDetails struct {
	// ProjectID is the GCP project the connection scopes to; its datasets play
	// the role of schemas. Globally unique and immutable, so it doubles as the
	// physical-database identity.
	ProjectID string
	// CredentialsFilePath is the path to the service-account JSON key file, read
	// directly by the Google client. Like SSHSettings.SSHKeyFilePath, the file is
	// the secret, so this is a plain path rather than a resolved secret value.
	CredentialsFilePath string
	// Location optionally pins the BigQuery processing location (e.g. "US", "EU",
	// "asia-south1") for datasets outside the client's auto-detected default.
	Location string
	// Endpoint overrides the BigQuery API endpoint. It is a test-only seam for
	// pointing the driver at a local emulator; it is never read from YAML, so
	// production configs cannot redirect the client to an arbitrary endpoint.
	Endpoint string
}

type TLSSettings struct {
	Mode               string
	RootCertFilePath   string
	ClientCertFilePath string
	ClientKeyFilePath  string
}

type SSHSettings struct {
	Host           string
	Port           int
	Username       string
	SSHKeyFilePath string
	Passphrase     string
}

// databasesFile mirrors the on-disk YAML shape: a named map of databases.
type databasesFile struct {
	Databases map[string]rawDatabase `yaml:"databases"`
}

// rawDatabase is one database entry as written in the YAML, before defaults are
// applied and secrets are resolved. A bigquery entry carries its details in the
// same `details` block, decoded into rawBigQueryDetails rather than rawDetails.
type rawDatabase struct {
	Type             DatabaseType `yaml:"type"`
	Details          *rawDetails  `yaml:"details"`
	ConnectionString string       `yaml:"connectionString"`
	TLS              *rawTLS      `yaml:"tls"`
	SSH              *rawSSH      `yaml:"ssh"`
}

type rawDetails struct {
	Host     string     `yaml:"host"`
	Port     int        `yaml:"port"`
	Database string     `yaml:"database"`
	Username string     `yaml:"username"`
	Password *rawSecret `yaml:"password"`

	// BigQuery-only fields. They share the `details` block with the OLTP fields
	// above; the JSON Schema keeps the two shapes from mixing in one entry.
	ProjectID           string `yaml:"projectId"`
	CredentialsFilePath string `yaml:"credentialsFilePath"`
	Location            string `yaml:"location"`
}

type rawTLS struct {
	Mode               string `yaml:"mode"`
	RootCertFilePath   string `yaml:"rootCertFilePath"`
	ClientCertFilePath string `yaml:"clientCertFilePath"`
	ClientKeyFilePath  string `yaml:"clientKeyFilePath"`
}

type rawSSH struct {
	Host           string     `yaml:"host"`
	Port           int        `yaml:"port"`
	Username       string     `yaml:"username"`
	SSHKeyFilePath string     `yaml:"sshKeyFilePath"`
	Passphrase     *rawSecret `yaml:"passphrase"`
}

// rawSecret is either a bare YAML scalar (shorthand for a literal value) or an
// object naming where to read the secret from (env/file/literal).
type rawSecret struct {
	From  string `yaml:"from"`
	Env   string `yaml:"env"`
	Path  string `yaml:"path"`
	Value string `yaml:"value"`

	scalar   string
	isScalar bool
}

// UnmarshalYAML accepts both the scalar shorthand and the object form.
func (s *rawSecret) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		s.isScalar = true
		s.scalar = node.Value
		return nil
	}
	type alias rawSecret
	var a alias
	if err := node.Decode(&a); err != nil {
		return err
	}
	*s = rawSecret(a)
	return nil
}

// resolve reads the secret to its plaintext value. A nil secret resolves to "".
func (s *rawSecret) resolve() (string, error) {
	if s == nil {
		return "", nil
	}
	if s.isScalar {
		return s.scalar, nil
	}
	switch s.From {
	case "literal":
		return s.Value, nil
	case "env":
		return os.Getenv(s.Env), nil
	case "file":
		contents, err := os.ReadFile(s.Path)
		if err != nil {
			return "", fmt.Errorf("read secret file %q: %w", s.Path, err)
		}
		return strings.TrimSpace(string(contents)), nil
	default:
		return "", fmt.Errorf("unknown secret source %q", s.From)
	}
}

// Load reads and validates the databases YAML at path against the embedded JSON
// Schema, resolves secrets and per-engine defaults, and returns the databases by
// name. A missing file is tolerated and yields an empty map.
func Load(path string) (map[string]Database, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]Database{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read databases file %q: %w", path, err)
	}

	var instance any
	if err := yaml.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("parse databases file %q: %w", path, err)
	}
	if err := validate(instance); err != nil {
		return nil, fmt.Errorf("databases config %q is invalid: %w", path, err)
	}

	var file databasesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode databases file %q: %w", path, err)
	}

	databases := make(map[string]Database, len(file.Databases))
	for name, raw := range file.Databases {
		resolved, err := raw.resolve(name)
		if err != nil {
			return nil, err
		}
		databases[name] = resolved
	}
	return databases, nil
}

// resolve turns a rawDatabase into the immutable Database value, applying
// per-engine defaults and resolving secrets.
func (raw rawDatabase) resolve(name string) (Database, error) {
	db := Database{
		Name:             name,
		Type:             raw.Type,
		ConnectionString: raw.ConnectionString,
	}

	if raw.Type == DatabaseTypeBigQuery {
		if raw.Details != nil {
			db.BigQuery = &BigQueryDetails{
				ProjectID:           raw.Details.ProjectID,
				CredentialsFilePath: raw.Details.CredentialsFilePath,
				Location:            raw.Details.Location,
			}
		}
	} else if raw.Details != nil {
		resolved := Details{
			Host:     raw.Details.Host,
			Port:     raw.Details.Port,
			Database: raw.Details.Database,
			Username: raw.Details.Username,
		}
		password, err := raw.Details.Password.resolve()
		if err != nil {
			return Database{}, fmt.Errorf("database %q password: %w", name, err)
		}
		resolved.Password = password
		applyDetailDefaults(&resolved, raw.Type)
		db.Details = &resolved
	}

	if raw.TLS != nil {
		db.TLS = &TLSSettings{
			Mode:               raw.TLS.Mode,
			RootCertFilePath:   raw.TLS.RootCertFilePath,
			ClientCertFilePath: raw.TLS.ClientCertFilePath,
			ClientKeyFilePath:  raw.TLS.ClientKeyFilePath,
		}
	}

	if raw.SSH != nil {
		resolved := SSHSettings{
			Host:           raw.SSH.Host,
			Port:           raw.SSH.Port,
			Username:       raw.SSH.Username,
			SSHKeyFilePath: raw.SSH.SSHKeyFilePath,
		}
		passphrase, err := raw.SSH.Passphrase.resolve()
		if err != nil {
			return Database{}, fmt.Errorf("database %q ssh passphrase: %w", name, err)
		}
		resolved.Passphrase = passphrase
		if resolved.Port == 0 {
			resolved.Port = 22
		}
		if resolved.Username == "" {
			resolved.Username = os.Getenv("USER")
		}
		db.SSH = &resolved
	}

	return db, nil
}

// applyDetailDefaults fills in the per-engine default port, user and database.
func applyDetailDefaults(d *Details, dbType DatabaseType) {
	switch dbType {
	case DatabaseTypePostgres:
		if d.Port == 0 {
			d.Port = 5432
		}
		if d.Database == "" {
			d.Database = "postgres"
		}
		if d.Username == "" {
			d.Username = "postgres"
		}
	case DatabaseTypeMySQL, DatabaseTypeMariaDB:
		if d.Port == 0 {
			d.Port = 3306
		}
		if d.Username == "" {
			d.Username = "root"
		}
	}
}

var (
	schemaOnce     sync.Once
	resolvedSchema *jsonschema.Resolved
	schemaErr      error
)

// validate checks a parsed config instance against the embedded JSON Schema.
func validate(instance any) error {
	schemaOnce.Do(func() {
		var schema jsonschema.Schema
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			schemaErr = err
			return
		}
		resolvedSchema, schemaErr = schema.Resolve(nil)
	})
	if schemaErr != nil {
		return fmt.Errorf("load embedded schema: %w", schemaErr)
	}
	return resolvedSchema.Validate(instance)
}
