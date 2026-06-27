// Package config loads runtime configuration from the environment with
// sensible localhost defaults, so the server boots zero-config in dev.
package config

import (
	"os"
	"strings"
	"time"
)

// Config holds all runtime knobs. It is built once at startup and passed
// (by pointer) into the OAuth and MCP layers — a minimal stand-in for the
// dbctx "services" dependency-injection container.
type Config struct {
	// AppPort serves MCP (/mcp) and OAuth (/oauth/*, /register, /.well-known/*).
	AppPort string
	// ControlPort serves /, /health and is reserved for future ops endpoints.
	ControlPort string
	// PublicBaseURL is the externally reachable origin of the app port. It is
	// advertised as the OAuth issuer and used to build endpoint URLs, so it
	// must match what clients actually dial. No trailing slash.
	PublicBaseURL string
	// DataDir holds the gitignored JSON state (clients, tokens, signing key).
	DataDir string
	// SigningKey, if set, is a base64-encoded HS256 secret supplied externally;
	// it takes precedence over the generated DATA_DIR/signing.key. Shorter keys
	// are zero-padded to 32 bytes.
	SigningKey string
	// AccessTTL is how long an access token is valid.
	AccessTTL time.Duration
	// RefreshTTL is how long a refresh token is valid.
	RefreshTTL time.Duration
}

// Load reads configuration from the environment, applying defaults.
func Load() *Config {
	return &Config{
		AppPort:       envOrDefault("APP_PORT", "8080"),
		ControlPort:   envOrDefault("CONTROL_PORT", "8081"),
		PublicBaseURL: strings.TrimRight(envOrDefault("PUBLIC_BASE_URL", "http://localhost:8080"), "/"),
		DataDir:       envOrDefault("DATA_DIR", "./data"),
		SigningKey:    os.Getenv("SIGNING_KEY"),
		AccessTTL:     durationOrDefault("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:    durationOrDefault("REFRESH_TTL", 7*24*time.Hour),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
