//go:build windows

package task

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeWindowsArg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple argument",
			input:    "test",
			expected: "test",
		},
		{
			name:     "argument with space",
			input:    "test arg",
			expected: "\"test arg\"",
		},
		{
			name:     "empty argument",
			input:    "",
			expected: "\"\"",
		},
		{
			name:     "argument with quotes",
			input:    "test \"quoted\" arg",
			expected: "\"test \\\"quoted\\\" arg\"",
		},
		{
			name:     "argument with backslash",
			input:    "C:\\Program Files\\kopia",
			expected: "\"C:\\Program Files\\kopia\"",
		},
		{
			name:     "argument with trailing backslash",
			input:    "C:\\path\\",
			expected: "\"C:\\path\\\\\"",
		},
		{
			name:     "argument with quote and backslash",
			input:    "C:\\path\\\"file\"",
			expected: "\"C:\\path\\\\\\\"file\\\"\"",
		},
		{
			name:     "argument with tab",
			input:    "test\targ",
			expected: "\"test\targ\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeWindowsArg(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
