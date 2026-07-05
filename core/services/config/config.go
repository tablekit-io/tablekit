// Package config loads runtime configuration from the environment with
// sensible localhost defaults, so the server boots zero-config in dev.
package config

import (
	"fmt"
	"net/url"
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
	// DatabasesFile is the YAML file declaring the databases run_query can query.
	// Defaults to ./databases.yaml (relative to the working directory); a missing
	// file means no databases.
	DatabasesFile string
	// DatabaseURL is a full postgres:// DSN for tablekit's own state database. It
	// is used only when the structured DB_* variables below are not all set (see
	// DatabaseDSN).
	DatabaseURL string
	// DBHost/DBPort/DBUser/DBPassword/DBName/DBSSLMode are the structured
	// connection parameters for tablekit's own state database. When DBHost,
	// DBUser and DBPassword are all set they take precedence over DatabaseURL.
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	// SigningKey is the required base64-encoded HS256 secret (SIGNING_KEY). It is
	// supplied externally — there is no generated fallback — so startup fails if
	// it is empty. Shorter keys are zero-padded to 32 bytes.
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
		DatabasesFile: envOrDefault("DATABASES_FILE", "databases.yaml"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        envOrDefault("DB_PORT", "5432"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        envOrDefault("DB_NAME", "tablekit"),
		DBSSLMode:     envOrDefault("DB_SSLMODE", "disable"),
		SigningKey:    os.Getenv("SIGNING_KEY"),
		AccessTTL:     durationOrDefault("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:    durationOrDefault("REFRESH_TTL", 7*24*time.Hour),
	}
}

// DatabaseDSN returns the connection string for tablekit's own state database,
// resolving the two configuration styles by precedence:
//
//  1. Structured DB_* variables, when DBHost, DBUser and DBPassword are all set:
//     assembled into a postgres:// URL (with DBPort/DBName/DBSSLMode defaults
//     already applied by Load). The password is URL-encoded, so secrets with
//     reserved characters need no manual escaping.
//  2. DatabaseURL (DATABASE_URL), when set: used verbatim.
//  3. A localhost default, so a plain `go run` against a local Postgres works
//     with zero configuration.
func (c *Config) DatabaseDSN() string {
	if c.DBHost != "" && c.DBUser != "" && c.DBPassword != "" {
		dsn := url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(c.DBUser, c.DBPassword),
			Host:     c.DBHost + ":" + c.DBPort,
			Path:     "/" + c.DBName,
			RawQuery: "sslmode=" + c.DBSSLMode,
		}
		return dsn.String()
	}
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return "postgres://postgres@localhost:5432/tablekit?sslmode=disable"
}

// ResolveDatabasesFile resolves the databases config path across the two YAML
// extensions: when the configured path (from DATABASES_FILE or the default) ends
// in `.yaml` or `.yml`, it accepts either <base>.yaml or <base>.yml, so a
// mount/env `.yml` vs `.yaml` mismatch still finds the file. If BOTH extensions
// exist the config is ambiguous and it returns an error (the caller fails
// startup). If neither exists it returns the configured path unchanged, so the
// loader treats it as "no databases". A configured path with any other extension
// (or none) is taken literally — no cross-extension resolution, no ambiguity check.
func ResolveDatabasesFile(configured string) (string, error) {
	extension := filepath.Ext(configured)
	if extension != ".yaml" && extension != ".yml" {
		return configured, nil
	}

	base := strings.TrimSuffix(configured, extension)
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
