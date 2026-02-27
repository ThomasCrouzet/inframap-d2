package render

import "strings"

const (
	terrastruct = "https://icons.terrastruct.com"
	selfhst     = "https://cdn.jsdelivr.net/gh/selfhst/icons/svg"
)

// iconRegistry maps service names/images to icon URLs.
var iconRegistry = map[string]string{
	// Databases
	"postgres":   terrastruct + "/dev/postgresql.svg",
	"postgresql": terrastruct + "/dev/postgresql.svg",
	"mysql":      terrastruct + "/dev/mysql.svg",
	"mariadb":    selfhst + "/mariadb.svg",
	"redis":      terrastruct + "/dev/redis.svg",
	"mongodb":    selfhst + "/mongodb.svg",
	"mongo":      selfhst + "/mongodb.svg",
	"couchdb":    selfhst + "/couchdb.svg",

	// Web/Proxy
	"nginx":               terrastruct + "/dev/nginx.svg",
	"nginx-proxy-manager": selfhst + "/nginx-proxy-manager.svg",
	"npm":                 selfhst + "/nginx-proxy-manager.svg",
	"traefik":             selfhst + "/traefik.svg",
	"caddy":               selfhst + "/caddy.svg",
	"cloudflare":          selfhst + "/cloudflare.svg",

	// Monitoring
	"netdata":    selfhst + "/netdata.svg",
	"grafana":    selfhst + "/grafana.svg",
	"prometheus": selfhst + "/prometheus.svg",
	"uptime-kuma": selfhst + "/uptime-kuma.svg",

	// Docker / Containers
	"docker":    terrastruct + "/dev/docker.svg",
	"portainer": selfhst + "/portainer.svg",

	// Languages / Frameworks
	"nodejs": terrastruct + "/dev/nodejs.svg",
	"node":   terrastruct + "/dev/nodejs.svg",
	"go":     selfhst + "/golang.svg",
	"python": terrastruct + "/dev/python.svg",

	// Infrastructure
	"tailscale":  selfhst + "/tailscale.svg",
	"cockpit":    selfhst + "/cockpit.svg",
	"kubernetes": terrastruct + "/dev/kubernetes.svg",
	"k8s":        terrastruct + "/dev/kubernetes.svg",
	"proxmox":    selfhst + "/proxmox.svg",
	"terraform":  terrastruct + "/dev/terraform.svg",

	// OS
	"linux":   terrastruct + "/dev/linux.svg",
	"debian":  terrastruct + "/dev/debian.svg",
	"macos":   terrastruct + "/dev/apple.svg",
	"ios":     terrastruct + "/dev/apple.svg",
	"android": terrastruct + "/dev/android.svg",
	"windows": terrastruct + "/dev/windows.svg",

	// Media
	"plex":       selfhst + "/plex.svg",
	"jellyfin":   selfhst + "/jellyfin.svg",
	"jellyseerr": selfhst + "/jellyseerr.svg",
	"radarr":     selfhst + "/radarr.svg",
	"sonarr":     selfhst + "/sonarr.svg",
	"prowlarr":   selfhst + "/prowlarr.svg",
	"bazarr":     selfhst + "/bazarr.svg",
	"overseerr":  selfhst + "/overseerr.svg",
	"tautulli":   selfhst + "/tautulli.svg",

	// Download
	"transmission": selfhst + "/transmission.svg",
	"qbittorrent":  selfhst + "/qbittorrent.svg",
	"sabnzbd":      selfhst + "/sabnzbd.svg",
	"gluetun":      selfhst + "/gluetun.svg",

	// Tools
	"vaultwarden":       selfhst + "/vaultwarden.svg",
	"bitwarden":         selfhst + "/bitwarden.svg",
	"homepage":          selfhst + "/homepage.svg",
	"homarr":            selfhst + "/homarr.svg",
	"home-assistant":    selfhst + "/home-assistant.svg",
	"homeassistant":     selfhst + "/home-assistant.svg",
	"stirling-pdf":      selfhst + "/stirling-pdf.svg",
	"it-tools":          selfhst + "/it-tools.svg",

	// Self-hosted services
	"n8n":                selfhst + "/n8n.svg",
	"gitea":              selfhst + "/gitea.svg",
	"vikunja":            selfhst + "/vikunja.svg",
	"ntfy":               selfhst + "/ntfy.svg",
	"semaphore":          selfhst + "/semaphore.svg",
	"kiwix":              selfhst + "/kiwix.svg",
	"audiobookshelf":     selfhst + "/audiobookshelf.svg",
	"recyclarr":          selfhst + "/recyclarr.svg",
	"super-productivity": selfhst + "/super-productivity.svg",
	"obsidian":           selfhst + "/obsidian.svg",
}

// LookupIcon returns the icon URL for a service name or image.
func LookupIcon(name, image string) string {
	// Try exact name match
	if url, ok := iconRegistry[strings.ToLower(name)]; ok {
		return url
	}

	// Try image-based matching
	imgLower := strings.ToLower(image)
	for key, url := range iconRegistry {
		if strings.Contains(imgLower, key) {
			return url
		}
	}

	// Try name parts
	nameLower := strings.ToLower(name)
	for key, url := range iconRegistry {
		if strings.Contains(nameLower, key) {
			return url
		}
	}

	return ""
}

// LookupOSIcon returns the icon URL for an OS string.
func LookupOSIcon(os string) string {
	osLower := strings.ToLower(os)
	for key, url := range iconRegistry {
		if strings.Contains(osLower, key) {
			return url
		}
	}
	return ""
}
