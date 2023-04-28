package ptr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.szostok.io/botkube-plugins/internal/ptr"
)

func TestFromType(t *testing.T) {
	type exampleStruct struct {
		Name string
	}
	tests := []struct {
		name  string
		given any
	}{
		{
			name:  "Test with number",
			given: 1,
		},
		{
			name:  "Test with string",
			given: "test",
		},
		{
			name:  "Test with bool",
			given: true,
		},
		{
			name: "Test with struct",
			given: exampleStruct{
				Name: "test",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := ptr.FromType(tc.given)

			// then
			assert.NotNil(t, got)
			assert.EqualValues(t, tc.given, *got)
		})
	}
}

func TestToValue(t *testing.T) {
	t.Run("Test with number", func(t *testing.T) {
		given := ptr.FromType(1)
		got := ptr.ToValue(given)
		assert.EqualValues(t, *given, got)
	})
	t.Run("Test with string", func(t *testing.T) {
		given := ptr.FromType("test")
		got := ptr.ToValue(given)
		assert.EqualValues(t, *given, got)
	})
	t.Run("Test with bool", func(t *testing.T) {
		given := ptr.FromType(true)
		got := ptr.ToValue(given)
		assert.EqualValues(t, *given, got)
	})
	t.Run("Test with nil", func(t *testing.T) {
		given := ptr.ToValue[bool](nil)
		assert.False(t, given)
	})
}
