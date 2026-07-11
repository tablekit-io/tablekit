package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeYAML writes a databases YAML to a temp file and returns its path.
func writeYAML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "databases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestLoadValidAndDefaults(t *testing.T) {
	path := writeYAML(t, `
databases:
  pg:
    type: postgres
    details:
      host: pg.internal
      username: app_ro
  my:
    type: mysql
    details:
      host: my.internal
      username: reader
`)
	databases, err := Load(path)
	require.NoError(t, err)

	pg := databases["pg"]
	require.NotNil(t, pg.Details)
	assert.Equal(t, 5432, pg.Details.Port)
	assert.Equal(t, "postgres", pg.Details.Database)
	assert.Equal(t, "app_ro", pg.Details.Username)

	my := databases["my"]
	require.NotNil(t, my.Details)
	assert.Equal(t, 3306, my.Details.Port)
	assert.Equal(t, "reader", my.Details.Username)
}

func TestLoadSecretResolution(t *testing.T) {
	t.Setenv("TEST_DB_PASSWORD", "from-env")
	secretFile := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("  from-file\n"), 0o600))

	path := writeYAML(t, `
databases:
  scalar:
    type: postgres
    details: { host: h, username: u, password: scalar-secret }
  fromenv:
    type: postgres
    details: { host: h, username: u, password: { from: env, env: TEST_DB_PASSWORD } }
  fromfile:
    type: postgres
    details: { host: h, username: u, password: { from: file, path: `+secretFile+` } }
  fromliteral:
    type: postgres
    details: { host: h, username: u, password: { from: literal, value: lit-secret } }
`)
	databases, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "scalar-secret", databases["scalar"].Details.Password)
	assert.Equal(t, "from-env", databases["fromenv"].Details.Password)
	assert.Equal(t, "from-file", databases["fromfile"].Details.Password) // trimmed
	assert.Equal(t, "lit-secret", databases["fromliteral"].Details.Password)
}

func TestLoadMissingFileTolerated(t *testing.T) {
	databases, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err)
	assert.Empty(t, databases)
}

func TestLoadRejectsUnknownType(t *testing.T) {
	path := writeYAML(t, `
databases:
  x:
    type: sqlite
    details: { host: h, username: u }
`)
	_, err := Load(path)
	assert.Error(t, err)
}

func TestLoadRejectsDetailsAndConnectionStringTogether(t *testing.T) {
	path := writeYAML(t, `
databases:
  x:
    type: postgres
    details: { host: h, username: u }
    connectionString: postgres://h/db
`)
	_, err := Load(path)
	assert.Error(t, err, "details and connectionString are mutually exclusive (XOR)")
}

func TestLoadBigQuery(t *testing.T) {
	path := writeYAML(t, `
databases:
  warehouse:
    type: bigquery
    details:
      projectId: my-gcp-project
      credentialsFilePath: /keys/sa.json
      location: EU
`)
	databases, err := Load(path)
	require.NoError(t, err)

	warehouse := databases["warehouse"]
	assert.Equal(t, DatabaseTypeBigQuery, warehouse.Type)
	assert.Nil(t, warehouse.Details, "bigquery has no OLTP details")
	require.NotNil(t, warehouse.BigQuery)
	assert.Equal(t, "my-gcp-project", warehouse.BigQuery.ProjectID)
	assert.Equal(t, "/keys/sa.json", warehouse.BigQuery.CredentialsFilePath)
	assert.Equal(t, "EU", warehouse.BigQuery.Location)
	assert.Empty(t, warehouse.BigQuery.Endpoint, "endpoint is never set from YAML")
}

func TestLoadBigQueryRejectsMissingFields(t *testing.T) {
	missingCredentials := writeYAML(t, `
databases:
  x:
    type: bigquery
    details: { projectId: p }
`)
	_, err := Load(missingCredentials)
	assert.Error(t, err, "credentialsFilePath is required")

	missingProject := writeYAML(t, `
databases:
  x:
    type: bigquery
    details: { credentialsFilePath: /keys/sa.json }
`)
	_, err = Load(missingProject)
	assert.Error(t, err, "projectId is required")
}

func TestLoadBigQueryRejectsSQLOnlyFields(t *testing.T) {
	// tls, ssh and connectionString are not part of the bigquery variant, so
	// additionalProperties:false rejects them.
	for _, body := range []string{
		`
databases:
  x:
    type: bigquery
    details: { projectId: p, credentialsFilePath: /k }
    tls: { mode: require }
`,
		`
databases:
  x:
    type: bigquery
    details: { projectId: p, credentialsFilePath: /k }
    ssh: { host: bastion, sshKeyFilePath: /k }
`,
		`
databases:
  x:
    type: bigquery
    connectionString: bigquery://p
`,
	} {
		_, err := Load(writeYAML(t, body))
		assert.Error(t, err)
	}
}

func TestLoadSSHDefaults(t *testing.T) {
	t.Setenv("USER", "deployer")
	path := writeYAML(t, `
databases:
  tunneled:
    type: postgres
    details: { host: db.internal, username: u }
    ssh:
      host: bastion.internal
      sshKeyFilePath: /tmp/key
`)
	databases, err := Load(path)
	require.NoError(t, err)
	ssh := databases["tunneled"].SSH
	require.NotNil(t, ssh)
	assert.Equal(t, 22, ssh.Port)
	assert.Equal(t, "deployer", ssh.Username)
}
