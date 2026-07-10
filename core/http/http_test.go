package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestLoggerEmitsJSON drives a request through the middleware and asserts
// it emits one structured event whose level tracks the response status.
func TestRequestLoggerEmitsJSON(t *testing.T) {
	cases := []struct {
		status int
		level  string
	}{
		{http.StatusOK, "info"},
		{http.StatusNotFound, "warn"},
		{http.StatusInternalServerError, "error"},
	}
	for _, testCase := range cases {
		t.Run(http.StatusText(testCase.status), func(t *testing.T) {
			var buffer bytes.Buffer
			original := log.Logger
			t.Cleanup(func() { log.Logger = original })
			log.Logger = zerolog.New(&buffer)

			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.Use(requestLogger())
			engine.GET("/thing", func(context *gin.Context) {
				context.Status(testCase.status)
			})

			request := httptest.NewRequest(http.MethodGet, "/thing?q=1", nil)
			engine.ServeHTTP(httptest.NewRecorder(), request)

			var record map[string]any
			require.NoError(t, json.Unmarshal(buffer.Bytes(), &record))
			assert.Equal(t, testCase.level, record["level"])
			assert.Equal(t, float64(testCase.status), record["status"])
			assert.Equal(t, http.MethodGet, record["method"])
			assert.Equal(t, "/thing?q=1", record["path"])
			assert.Equal(t, "request", record["message"])
		})
	}
}
