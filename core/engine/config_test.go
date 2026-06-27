package engine

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
	svc, err := Load(path, Limits{})
	require.NoError(t, err)

	pg := svc.databases["pg"]
	require.NotNil(t, pg.details)
	assert.Equal(t, 5432, pg.details.port)
	assert.Equal(t, "postgres", pg.details.database)
	assert.Equal(t, "app_ro", pg.details.username)

	my := svc.databases["my"]
	require.NotNil(t, my.details)
	assert.Equal(t, 3306, my.details.port)
	assert.Equal(t, "reader", my.details.username)

	infos := svc.List()
	require.Len(t, infos, 2)
	assert.Equal(t, "my", infos[0].Name)
	assert.Equal(t, "mysql", infos[0].Type)
	assert.Equal(t, "pg", infos[1].Name)
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
	svc, err := Load(path, Limits{})
	require.NoError(t, err)

	assert.Equal(t, "scalar-secret", svc.databases["scalar"].details.password)
	assert.Equal(t, "from-env", svc.databases["fromenv"].details.password)
	assert.Equal(t, "from-file", svc.databases["fromfile"].details.password) // trimmed
	assert.Equal(t, "lit-secret", svc.databases["fromliteral"].details.password)
}

func TestLoadMissingFileTolerated(t *testing.T) {
	svc, err := Load(filepath.Join(t.TempDir(), "nope.yaml"), Limits{})
	require.NoError(t, err)
	assert.Empty(t, svc.List())
}

func TestLoadRejectsUnknownType(t *testing.T) {
	path := writeYAML(t, `
databases:
  x:
    type: sqlite
    details: { host: h, username: u }
`)
	_, err := Load(path, Limits{})
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
	_, err := Load(path, Limits{})
	assert.Error(t, err, "details and connectionString are mutually exclusive (XOR)")
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
	svc, err := Load(path, Limits{})
	require.NoError(t, err)
	ssh := svc.databases["tunneled"].ssh
	require.NotNil(t, ssh)
	assert.Equal(t, 22, ssh.port)
	assert.Equal(t, "deployer", ssh.username)
}

func TestLimitsDefaults(t *testing.T) {
	svc, err := Load(filepath.Join(t.TempDir(), "nope.yaml"), Limits{})
	require.NoError(t, err)
	assert.Equal(t, 2048, svc.limits.MaxRows)
	assert.Equal(t, 64*1024, svc.limits.MaxBytes)
	assert.NotZero(t, svc.limits.StatementTimeout)
}
