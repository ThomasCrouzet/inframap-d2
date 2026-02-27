package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePortMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected PortMapping
	}{
		{
			"8080",
			PortMapping{HostPort: 8080, ContainerPort: 8080, Protocol: "tcp"},
		},
		{
			"8080:80",
			PortMapping{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
		{
			"127.0.0.1:8080:80",
			PortMapping{HostIP: "127.0.0.1", HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
		{
			"8080:80/udp",
			PortMapping{HostPort: 8080, ContainerPort: 80, Protocol: "udp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParsePortMapping(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestPortMappingString(t *testing.T) {
	tests := []struct {
		pm       PortMapping
		expected string
	}{
		{PortMapping{HostPort: 8080, ContainerPort: 8080, Protocol: "tcp"}, "8080"},
		{PortMapping{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"}, "8080→80"},
		{PortMapping{HostPort: 8080, ContainerPort: 80, Protocol: "udp"}, "8080→80/udp"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pm.String())
		})
	}
}
