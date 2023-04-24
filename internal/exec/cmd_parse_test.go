package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.szostok.io/botkube-plugins/internal/ptr"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected Command
	}{
		{
			input: "x run helm list -A",
			expected: Command{
				ToExecute:     "x run helm list -A",
				IsRawRequired: false,
				SelectIndex:   nil,
			},
		},
		{
			input: "x run helm list -A @no-interactivity",
			expected: Command{
				ToExecute:     "x run helm list -A",
				IsRawRequired: true,
				SelectIndex:   nil,
			},
		},
		{
			input: "x run kubectl get pods @idx:123",
			expected: Command{
				ToExecute:     "x run kubectl get pods",
				IsRawRequired: false,
				SelectIndex:   ptr.FromType(123),
			},
		},
		{
			input: "x run kubectl get pods @idx:abc",
			expected: Command{
				ToExecute:     "x run kubectl get pods @idx:abc",
				IsRawRequired: false,
				SelectIndex:   nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			// when
			gotCmd := Parse(tc.input)

			assert.Equal(t, tc.expected.ToExecute, gotCmd.ToExecute)
			assert.Equal(t, tc.expected.IsRawRequired, gotCmd.IsRawRequired)
			assert.EqualValues(t, tc.expected.SelectIndex, gotCmd.SelectIndex)
		})
	}
}
