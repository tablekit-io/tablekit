package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	// Clear every var Load reads so we observe the built-in defaults.
	for _, k := range []string{
		"APP_PORT", "CONTROL_PORT", "PUBLIC_BASE_URL", "DATA_DIR",
		"ACCESS_TTL", "REFRESH_TTL",
	} {
		t.Setenv(k, "")
	}

	cfg := Load()

	assert.Equal(t, "8080", cfg.AppPort)
	assert.Equal(t, "8081", cfg.ControlPort)
	assert.Equal(t, "http://localhost:8080", cfg.PublicBaseURL)
	assert.Equal(t, "./data", cfg.DataDir)
	assert.Equal(t, 15*time.Minute, cfg.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, cfg.RefreshTTL)
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("APP_PORT", "9000")
	t.Setenv("CONTROL_PORT", "9001")
	t.Setenv("DATA_DIR", "/var/lib/tablekit")
	t.Setenv("ACCESS_TTL", "30s")
	t.Setenv("REFRESH_TTL", "48h")
	// Trailing slash must be trimmed so endpoint URLs don't double up.
	t.Setenv("PUBLIC_BASE_URL", "https://mcp.example.com/")

	cfg := Load()

	assert.Equal(t, "9000", cfg.AppPort)
	assert.Equal(t, "9001", cfg.ControlPort)
	assert.Equal(t, "/var/lib/tablekit", cfg.DataDir)
	assert.Equal(t, 30*time.Second, cfg.AccessTTL)
	assert.Equal(t, 48*time.Hour, cfg.RefreshTTL)
	assert.Equal(t, "https://mcp.example.com", cfg.PublicBaseURL)
}

func TestLoadBadDurationFallsBack(t *testing.T) {
	t.Setenv("ACCESS_TTL", "not-a-duration")
	t.Setenv("REFRESH_TTL", "")

	cfg := Load()

	// Unparseable duration → built-in default, not a zero value.
	assert.Equal(t, 15*time.Minute, cfg.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, cfg.RefreshTTL)
}
