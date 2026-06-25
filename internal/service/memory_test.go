package service

import (
	"strings"
	"testing"
)

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:     "Short text, no truncation",
			text:     "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "Exact length, no truncation",
			text:     "Hello World",
			maxLen:   11,
			expected: "Hello World",
		},
		{
			name:     "Truncation needed",
			text:     "Hello World, how are you?",
			maxLen:   15,
			expected: "Hello World,...",
		},
		{
			name:     "Truncation length too small",
			text:     "Hello World",
			maxLen:   2,
			expected: "He",
		},
		{
			name:     "Whitespace trimming",
			text:     "   Trim me   ",
			maxLen:   10,
			expected: "Trim me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractRememberTitle(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Single line text",
			text:     "This is a single line memory item.",
			expected: "This is a single line memory item.",
		},
		{
			name:     "Multi-line text, uses first line",
			text:     "First line of memory\nSecond line of memory\nThird line",
			expected: "First line of memory",
		},
		{
			name:     "Empty text fallback",
			text:     "   \n  \n",
			expected: "Untitled memory",
		},
		{
			name:     "Extremely long first line gets truncated",
			text:     strings.Repeat("A", 150) + "\nSecond line",
			expected: strings.Repeat("A", 117) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRememberTitle(tt.text)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
