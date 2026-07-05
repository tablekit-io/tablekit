// Package config loads runtime configuration from the environment with
// sensible localhost defaults, so the server boots zero-config in dev.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds all runtime knobs. It is built once at startup and bundled into
// the Services value (alongside the store, engine and issuer), which is threaded
// through the OAuth and MCP layers.
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
	// DatabasesFile is the YAML file declaring the databases run_query can query.
	// Defaults to DATA_DIR/databases.yaml; a missing file means no databases.
	DatabasesFile string
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
	dataDir := envOrDefault("DATA_DIR", "./data")
	return &Config{
		AppPort:       envOrDefault("APP_PORT", "8080"),
		ControlPort:   envOrDefault("CONTROL_PORT", "8081"),
		PublicBaseURL: strings.TrimRight(envOrDefault("PUBLIC_BASE_URL", "http://localhost:8080"), "/"),
		DataDir:       dataDir,
		DatabasesFile: envOrDefault("DATABASES_FILE", filepath.Join(dataDir, "databases.yaml")),
		SigningKey:    os.Getenv("SIGNING_KEY"),
		AccessTTL:     durationOrDefault("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:    durationOrDefault("REFRESH_TTL", 7*24*time.Hour),
	}
}

// ResolveDatabasesFile resolves the databases config path by base name, ignoring
// the extension: given a configured path (from DATABASES_FILE or the default), it
// accepts either <base>.yaml or <base>.yml, so a mount/env `.yml` vs `.yaml`
// mismatch still finds the file. If BOTH extensions exist the config is ambiguous
// and it returns an error (the caller fails startup). If neither exists it returns
// the configured path unchanged, so the loader treats it as "no databases".
func ResolveDatabasesFile(configured string) (string, error) {
	base := strings.TrimSuffix(strings.TrimSuffix(configured, ".yaml"), ".yml")
	yamlPath := base + ".yaml"
	ymlPath := base + ".yml"

	yamlExists := fileExists(yamlPath)
	ymlExists := fileExists(ymlPath)

	switch {
	case yamlExists && ymlExists:
		return "", fmt.Errorf("ambiguous databases config: both %q and %q exist; keep only one", yamlPath, ymlPath)
	case yamlExists:
		return yamlPath, nil
	case ymlExists:
		return ymlPath, nil
	default:
		return configured, nil
	}
}

// fileExists reports whether path names an existing regular file (not a directory).
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
