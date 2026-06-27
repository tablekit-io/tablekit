package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelloWorld(t *testing.T) {
	h := &Handlers{}
	tests := []struct {
		name string
		in   helloInput
		want string
	}{
		{"with name", helloInput{Name: "omran"}, "Hello, omran!"},
		{"empty name", helloInput{}, "Hello, world!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, out, err := h.helloWorld(context.Background(), nil, tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, out.Greeting)
			require.Len(t, result.Content, 1)
			text, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Equal(t, tt.want, text.Text)
		})
	}
}
