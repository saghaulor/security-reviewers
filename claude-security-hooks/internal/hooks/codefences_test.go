package hooks

import (
	"testing"
)

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "raw JSON",
			input:    `{"a":1}`,
			expected: `{"a":1}`,
		},
		{
			name:     "json fence",
			input:    "```json\n{\"a\":1}\n```",
			expected: `{"a":1}`,
		},
		{
			name:     "bare fence",
			input:    "```\n{\"a\":1}\n```",
			expected: `{"a":1}`,
		},
		{
			name:     "trailing whitespace",
			input:    "```{\"a\":1}```\n",
			expected: `{"a":1}`,
		},
		{
			name:     "leading whitespace",
			input:    "  ```json\n{\"a\":1}\n```",
			expected: `{"a":1}`,
		},
		{
			name:     "internal backticks",
			input:    "```\nhello ``` world\n```",
			expected: "hello ``` world",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCodeFences(tt.input)
			if got != tt.expected {
				t.Errorf("stripCodeFences(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
