package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	// Clear every var Load reads so we observe the built-in defaults.
	for _, k := range []string{
		"APP_PORT", "CONTROL_PORT", "PUBLIC_BASE_URL",
		"ACCESS_TTL", "REFRESH_TTL",
	} {
		t.Setenv(k, "")
	}

	configService := Load()

	assert.Equal(t, "8080", configService.AppPort)
	assert.Equal(t, "8081", configService.ControlPort)
	assert.Equal(t, "http://localhost:8080", configService.PublicBaseURL)
	assert.Equal(t, 15*time.Minute, configService.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, configService.RefreshTTL)
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("APP_PORT", "9000")
	t.Setenv("CONTROL_PORT", "9001")
	t.Setenv("ACCESS_TTL", "30s")
	t.Setenv("REFRESH_TTL", "48h")
	// Trailing slash must be trimmed so endpoint URLs don't double up.
	t.Setenv("PUBLIC_BASE_URL", "https://mcp.example.com/")

	configService := Load()

	assert.Equal(t, "9000", configService.AppPort)
	assert.Equal(t, "9001", configService.ControlPort)
	assert.Equal(t, 30*time.Second, configService.AccessTTL)
	assert.Equal(t, 48*time.Hour, configService.RefreshTTL)
	assert.Equal(t, "https://mcp.example.com", configService.PublicBaseURL)
}

func TestIsDevelopment(t *testing.T) {
	// Only the exact value "development" enables development behavior; everything
	// else (including empty and near-misses) is default-deny.
	cases := map[string]bool{
		"development": true,
		"":            false,
		"production":  false,
		"dev":         false,
		"Development": false,
		"development ": false,
	}
	for value, want := range cases {
		t.Run("TABLEKIT_ENV="+value, func(t *testing.T) {
			t.Setenv("TABLEKIT_ENV", value)
			assert.Equal(t, want, Load().IsDevelopment())
		})
	}
}

func TestResolveDatabasesFile(t *testing.T) {
	write := func(t *testing.T, path string) {
		require.NoError(t, os.WriteFile(path, []byte("databases: {}\n"), 0o600))
	}

	t.Run("only .yaml present is used", func(t *testing.T) {
		dir := t.TempDir()
		yamlPath := filepath.Join(dir, "databases.yaml")
		write(t, yamlPath)

		// Configured hint carries the .yml extension; resolution ignores it.
		resolved, err := ResolveDatabasesFile(filepath.Join(dir, "databases.yml"))
		require.NoError(t, err)
		assert.Equal(t, yamlPath, resolved)
	})

	t.Run("only .yml present is used", func(t *testing.T) {
		dir := t.TempDir()
		ymlPath := filepath.Join(dir, "databases.yml")
		write(t, ymlPath)

		// Configured hint carries the .yaml extension; resolution ignores it.
		resolved, err := ResolveDatabasesFile(filepath.Join(dir, "databases.yaml"))
		require.NoError(t, err)
		assert.Equal(t, ymlPath, resolved)
	})

	t.Run("both present is a fatal error", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "databases.yaml"))
		write(t, filepath.Join(dir, "databases.yml"))

		_, err := ResolveDatabasesFile(filepath.Join(dir, "databases.yaml"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "databases.yaml")
		assert.Contains(t, err.Error(), "databases.yml")
	})

	t.Run("neither present returns the configured path unchanged", func(t *testing.T) {
		dir := t.TempDir()
		configured := filepath.Join(dir, "databases.yaml")

		resolved, err := ResolveDatabasesFile(configured)
		require.NoError(t, err)
		assert.Equal(t, configured, resolved)
	})

	t.Run("configured hint without extension is taken literally", func(t *testing.T) {
		dir := t.TempDir()
		// A sibling .yml exists, but the configured path has no YAML extension,
		// so it must NOT be picked up — the path is used verbatim.
		write(t, filepath.Join(dir, "databases.yml"))
		configured := filepath.Join(dir, "databases")

		resolved, err := ResolveDatabasesFile(configured)
		require.NoError(t, err)
		assert.Equal(t, configured, resolved)
	})

	t.Run("non-YAML extension is taken literally", func(t *testing.T) {
		dir := t.TempDir()
		// Even with a sibling .yaml present, a non-YAML extension gets no
		// cross-extension resolution and no ambiguity check.
		write(t, filepath.Join(dir, "databases.yaml"))
		configured := filepath.Join(dir, "databases.conf")

		resolved, err := ResolveDatabasesFile(configured)
		require.NoError(t, err)
		assert.Equal(t, configured, resolved)
	})
}

func TestLoadBadDurationFallsBack(t *testing.T) {
	t.Setenv("ACCESS_TTL", "not-a-duration")
	t.Setenv("REFRESH_TTL", "")

	configService := Load()

	// Unparseable duration → built-in default, not a zero value.
	assert.Equal(t, 15*time.Minute, configService.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, configService.RefreshTTL)
}
