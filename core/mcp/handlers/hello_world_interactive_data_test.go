package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInteractiveDataReturnsRandomSlices(t *testing.T) {
	tests := []struct {
		name string
		in   dataInput
		want int
	}{
		{"default", dataInput{}, 5},
		{"custom", dataInput{Slices: 3}, 3},
		{"clamped low", dataInput{Slices: 0}, 5},
		{"clamped high", dataInput{Slices: 99}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, out, err := helloInteractiveData(context.Background(), nil, tt.in)
			require.NoError(t, err)
			require.Len(t, out.Data, tt.want)
			for _, s := range out.Data {
				assert.NotEmpty(t, s.Label)
				assert.GreaterOrEqual(t, s.Value, float64(10))
				assert.LessOrEqual(t, s.Value, float64(100))
			}
		})
	}
}
