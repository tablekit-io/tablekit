// Package logging installs zerolog as the single structured-logging backend for
// the whole process. It emits JSON to stdout and, crucially, redirects the other
// loggers this service pulls in — the standard library log package, log/slog, and
// goose's own logger — into the same zerolog stream, so nothing reaches stdout in
// a raw, unstructured format.
//
// Wiring is global (matching the codebase's existing all-global logging style):
// Init configures zerolog's package-level logger and the redirects, and call
// sites use github.com/rs/zerolog/log directly. Init must run before anything
// logs — in particular before services.New, which opens the database and triggers
// goose during startup — so main calls it first thing.
package logging

import (
	standardlog "log"
	"log/slog"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// defaultLevel is used when LOG_LEVEL is unset or unparseable, so a missing or
// misconfigured level never silences logging or crashes startup.
const defaultLevel = zerolog.InfoLevel

// Init configures the global zerolog logger for JSON output on stdout and points
// the standard library log and log/slog defaults at it, then applies the level
// from the LOG_LEVEL environment variable. It is safe to call once at startup.
func Init() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(levelFromEnv())

	// Route the standard library log package (used by the databases watcher and
	// cli/root's log.Fatal) through zerolog. Clearing the flags drops log's own
	// timestamp/prefix so zerolog supplies the only one, and each write becomes a
	// single JSON message-level event.
	standardlog.SetFlags(0)
	standardlog.SetOutput(log.Logger)

	// Route log/slog (used by the MCP audit middleware) through the same zerolog
	// logger so its records share one JSON format.
	slog.SetDefault(slog.New(zerolog.NewSlogHandler(log.Logger)))
}

// levelFromEnv parses LOG_LEVEL — one of trace, debug, info, warn, error —
// falling back to defaultLevel for an empty or unrecognized value.
func levelFromEnv() zerolog.Level {
	raw := os.Getenv("LOG_LEVEL")
	if raw == "" {
		return defaultLevel
	}
	level, err := zerolog.ParseLevel(raw)
	if err != nil {
		return defaultLevel
	}
	return level
}

// GooseLogger returns an adapter implementing goose's Logger interface
// (Printf/Fatalf) that forwards to the global zerolog logger, so goose migration
// output (e.g. "goose: no migrations to run") is emitted as JSON rather than
// goose's raw stdlib format. Pass it to goose.SetLogger before running migrations.
func GooseLogger() *gooseLogger {
	return &gooseLogger{}
}

// gooseLogger adapts goose's Logger interface to zerolog. goose calls Printf for
// informational progress and Fatalf for fatal errors.
type gooseLogger struct{}

func (gooseLogger) Printf(format string, values ...any) {
	log.Info().Msgf(trimNewline(format), values...)
}

func (gooseLogger) Fatalf(format string, values ...any) {
	log.Fatal().Msgf(trimNewline(format), values...)
}

// trimNewline drops a single trailing newline goose appends to some formats, so
// the JSON message field does not carry a stray "\n".
func trimNewline(format string) string {
	if len(format) > 0 && format[len(format)-1] == '\n' {
		return format[:len(format)-1]
	}
	return format
}
