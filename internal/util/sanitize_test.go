package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gateway", "gateway"},
		{"uptime-kuma", "uptime-kuma"},
		{"my server", "my-server"},
		{"node.js", "node-js"},
		{"path/to/thing", "path-to-thing"},
		{"special@chars!", "specialchars"},
		{"", "unknown"},
		{"MixedCase", "mixedcase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeID(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestQuote(t *testing.T) {
	assert.Equal(t, `"hello"`, Quote("hello"))
	assert.Equal(t, `"say \"hi\""`, Quote(`say "hi"`))
}
