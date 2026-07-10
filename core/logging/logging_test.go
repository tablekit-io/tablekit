package logging

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelFromEnv(t *testing.T) {
	cases := map[string]zerolog.Level{
		"trace": zerolog.TraceLevel,
		"debug": zerolog.DebugLevel,
		"info":  zerolog.InfoLevel,
		"warn":  zerolog.WarnLevel,
		"error": zerolog.ErrorLevel,
	}
	for value, expected := range cases {
		t.Run(value, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", value)
			assert.Equal(t, expected, levelFromEnv())
		})
	}
}

func TestLevelFromEnvFallsBackToInfo(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "")
		assert.Equal(t, defaultLevel, levelFromEnv())
	})
	t.Run("unparseable", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "bogus")
		assert.Equal(t, defaultLevel, levelFromEnv())
	})
}

// TestGooseLoggerEmitsJSON verifies the goose adapter forwards Printf through
// zerolog as a JSON info event with the trailing newline trimmed.
func TestGooseLoggerEmitsJSON(t *testing.T) {
	var buffer bytes.Buffer
	original := log.Logger
	t.Cleanup(func() { log.Logger = original })
	log.Logger = zerolog.New(&buffer)

	GooseLogger().Printf("goose: no migrations to run. current version: %d\n", 1)

	var record map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &record))
	assert.Equal(t, "info", record["level"])
	assert.Equal(t, "goose: no migrations to run. current version: 1", record["message"])
}
