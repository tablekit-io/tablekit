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

	configService := Load()

	assert.Equal(t, "8080", configService.AppPort)
	assert.Equal(t, "8081", configService.ControlPort)
	assert.Equal(t, "http://localhost:8080", configService.PublicBaseURL)
	assert.Equal(t, "./data", configService.DataDir)
	assert.Equal(t, 15*time.Minute, configService.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, configService.RefreshTTL)
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("APP_PORT", "9000")
	t.Setenv("CONTROL_PORT", "9001")
	t.Setenv("DATA_DIR", "/var/lib/tablekit")
	t.Setenv("ACCESS_TTL", "30s")
	t.Setenv("REFRESH_TTL", "48h")
	// Trailing slash must be trimmed so endpoint URLs don't double up.
	t.Setenv("PUBLIC_BASE_URL", "https://mcp.example.com/")

	configService := Load()

	assert.Equal(t, "9000", configService.AppPort)
	assert.Equal(t, "9001", configService.ControlPort)
	assert.Equal(t, "/var/lib/tablekit", configService.DataDir)
	assert.Equal(t, 30*time.Second, configService.AccessTTL)
	assert.Equal(t, 48*time.Hour, configService.RefreshTTL)
	assert.Equal(t, "https://mcp.example.com", configService.PublicBaseURL)
}

func TestLoadBadDurationFallsBack(t *testing.T) {
	t.Setenv("ACCESS_TTL", "not-a-duration")
	t.Setenv("REFRESH_TTL", "")

	configService := Load()

	// Unparseable duration → built-in default, not a zero value.
	assert.Equal(t, 15*time.Minute, configService.AccessTTL)
	assert.Equal(t, 7*24*time.Hour, configService.RefreshTTL)
}
