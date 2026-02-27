package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeService(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{"radarr", "ghcr.io/hotio/radarr:latest", "media"},
		{"sonarr", "linuxserver/sonarr", "media"},
		{"transmission", "linuxserver/transmission:latest", "downloads"},
		{"qbittorrent", "qbittorrentofficial/qbittorrent-nox", "downloads"},
		{"traefik", "traefik:v3", "infrastructure"},
		{"nginx-proxy-manager", "jc21/nginx-proxy-manager", "infrastructure"},
		{"uptime-kuma", "louislam/uptime-kuma:1", "monitoring"},
		{"netdata", "", "monitoring"},
		{"vikunja", "vikunja/vikunja", "productivity"},
		{"gitea", "gitea/gitea:latest", "dev"},
		{"vaultwarden", "vaultwarden/server", "security"},
		{"my-custom-app", "myapp:latest", ""},
		{"stirling-pdf", "frooodle/s-pdf:latest", "tools"},
		{"homepage", "ghcr.io/gethomepage/homepage", "tools"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeService(tt.name, tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}
