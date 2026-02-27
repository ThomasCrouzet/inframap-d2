package model

import "strings"

// categoryPatterns maps service/image substrings to category names.
var categoryPatterns = map[string]string{
	// Media
	"plex":       "media",
	"jellyfin":   "media",
	"jellyseerr": "media",
	"radarr":     "media",
	"sonarr":     "media",
	"prowlarr":   "media",
	"bazarr":     "media",
	"overseerr":  "media",
	"tautulli":   "media",
	"emby":       "media",
	"kodi":       "media",

	// Downloads
	"transmission": "downloads",
	"qbittorrent":  "downloads",
	"sabnzbd":      "downloads",
	"gluetun":      "downloads",
	"nzbget":       "downloads",
	"deluge":       "downloads",
	"aria2":        "downloads",

	// Infrastructure
	"traefik":             "infrastructure",
	"nginx":               "infrastructure",
	"nginx-proxy-manager": "infrastructure",
	"caddy":               "infrastructure",
	"portainer":           "infrastructure",
	"docker":              "infrastructure",
	"watchtower":          "infrastructure",

	// Monitoring
	"netdata":    "monitoring",
	"grafana":    "monitoring",
	"prometheus": "monitoring",
	"uptime-kuma": "monitoring",
	"cockpit":    "monitoring",

	// Tools
	"stirling-pdf": "tools",
	"it-tools":     "tools",
	"homepage":     "tools",
	"homarr":       "tools",
	"dashy":        "tools",

	// Productivity
	"vikunja":            "productivity",
	"n8n":                "productivity",
	"super-productivity": "productivity",

	// Dev
	"gitea":     "dev",
	"gitlab":    "dev",
	"forgejo":   "dev",
	"semaphore": "dev",

	// Home
	"home-assistant": "home",
	"homeassistant":  "home",

	// Security
	"vaultwarden": "security",
	"bitwarden":   "security",
	"authelia":    "security",

	// Communication
	"ntfy": "communication",
}

// CategorizeService determines the category for a service based on its name and image.
func CategorizeService(name, image string) string {
	lower := strings.ToLower(name + " " + image)

	// Try exact name match first
	if cat, ok := categoryPatterns[strings.ToLower(name)]; ok {
		return cat
	}

	// Try substring match
	for pattern, cat := range categoryPatterns {
		if strings.Contains(lower, pattern) {
			return cat
		}
	}

	return ""
}
