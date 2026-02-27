package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripJinja2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"{{ docker_services_tailscale_ip }}:7200:8080",
			"PLACEHOLDER:7200:8080",
		},
		{
			"{{ var1 }} and {{ var2 }}",
			"PLACEHOLDER and PLACEHOLDER",
		},
		{
			"no jinja here",
			"no jinja here",
		},
		{
			"port: {{ netdata_port }}",
			"port: PLACEHOLDER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StripJinja2(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
