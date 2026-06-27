package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

// databaseType is the engine discriminator used to route to an implementation.
type databaseType string

const (
	databaseTypePostgres databaseType = "postgres"
	databaseTypeMySQL    databaseType = "mysql"
	databaseTypeMariaDB  databaseType = "mariadb"
)

// databasesFile mirrors the on-disk YAML shape: a named map of databases.
type databasesFile struct {
	Databases map[string]rawDatabase `yaml:"databases"`
}

// rawDatabase is one database entry as written in the YAML, before defaults are
// applied and secrets are resolved.
type rawDatabase struct {
	Type             databaseType `yaml:"type"`
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

// database is a resolved, ready-to-dial definition: secrets are plaintext and
// per-engine defaults are filled in. This is what the engine implementations
// consume; the raw YAML types never cross that boundary.
type database struct {
	name             string
	dbType           databaseType
	connectionString string // empty when details-based
	details          *details
	tls              *tlsSettings
	ssh              *sshSettings
}

type details struct {
	host     string
	port     int
	database string
	username string
	password string
}

type tlsSettings struct {
	mode               string
	rootCertFilePath   string
	clientCertFilePath string
	clientKeyFilePath  string
}

type sshSettings struct {
	host           string
	port           int
	username       string
	sshKeyFilePath string
	passphrase     string
}

// resolve turns a rawDatabase into the immutable database value, applying
// per-engine defaults and resolving secrets.
func (raw rawDatabase) resolve(name string) (database, error) {
	db := database{
		name:             name,
		dbType:           raw.Type,
		connectionString: raw.ConnectionString,
	}

	if raw.Details != nil {
		resolved := details{
			host:     raw.Details.Host,
			port:     raw.Details.Port,
			database: raw.Details.Database,
			username: raw.Details.Username,
		}
		password, err := raw.Details.Password.resolve()
		if err != nil {
			return database{}, fmt.Errorf("database %q password: %w", name, err)
		}
		resolved.password = password
		applyDetailDefaults(&resolved, raw.Type)
		db.details = &resolved
	}

	if raw.TLS != nil {
		db.tls = &tlsSettings{
			mode:               raw.TLS.Mode,
			rootCertFilePath:   raw.TLS.RootCertFilePath,
			clientCertFilePath: raw.TLS.ClientCertFilePath,
			clientKeyFilePath:  raw.TLS.ClientKeyFilePath,
		}
	}

	if raw.SSH != nil {
		resolved := sshSettings{
			host:           raw.SSH.Host,
			port:           raw.SSH.Port,
			username:       raw.SSH.Username,
			sshKeyFilePath: raw.SSH.SSHKeyFilePath,
		}
		passphrase, err := raw.SSH.Passphrase.resolve()
		if err != nil {
			return database{}, fmt.Errorf("database %q ssh passphrase: %w", name, err)
		}
		resolved.passphrase = passphrase
		if resolved.port == 0 {
			resolved.port = 22
		}
		if resolved.username == "" {
			resolved.username = os.Getenv("USER")
		}
		db.ssh = &resolved
	}

	return db, nil
}

// applyDetailDefaults fills in the per-engine default port, user and database.
func applyDetailDefaults(d *details, dbType databaseType) {
	switch dbType {
	case databaseTypePostgres:
		if d.port == 0 {
			d.port = 5432
		}
		if d.database == "" {
			d.database = "postgres"
		}
		if d.username == "" {
			d.username = "postgres"
		}
	case databaseTypeMySQL, databaseTypeMariaDB:
		if d.port == 0 {
			d.port = 3306
		}
		if d.username == "" {
			d.username = "root"
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
