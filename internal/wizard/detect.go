package wizard

import (
	"os"
	"os/exec"
	"path/filepath"
)

// DetectionResult holds what was auto-detected on the system.
type DetectionResult struct {
	TailscaleAvailable bool
	AnsibleInventory   string // path if found, empty otherwise
	ComposeFiles       []string
}

// Detector abstracts filesystem and path lookups for testing.
type Detector interface {
	LookPath(name string) (string, error)
	Stat(path string) (os.FileInfo, error)
	Glob(pattern string) ([]string, error)
}

// OSDetector uses the real OS for detection.
type OSDetector struct{}

func (OSDetector) LookPath(name string) (string, error) { return exec.LookPath(name) }
func (OSDetector) Stat(path string) (os.FileInfo, error) { return os.Stat(path) }
func (OSDetector) Glob(pattern string) ([]string, error) { return filepath.Glob(pattern) }

// Detect scans the environment for known infrastructure sources.
func Detect(d Detector) DetectionResult {
	if d == nil {
		d = OSDetector{}
	}

	result := DetectionResult{}

	// Check for tailscale binary
	if _, err := d.LookPath("tailscale"); err == nil {
		result.TailscaleAvailable = true
	}

	// Look for Ansible inventory
	inventoryPaths := []string{
		"hosts.yml",
		"inventory/hosts.yml",
		"../inventory/hosts.yml",
	}
	for _, p := range inventoryPaths {
		if _, err := d.Stat(p); err == nil {
			result.AnsibleInventory = p
			break
		}
	}

	// Look for Docker Compose files
	composePatterns := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}
	for _, pattern := range composePatterns {
		if _, err := d.Stat(pattern); err == nil {
			result.ComposeFiles = append(result.ComposeFiles, pattern)
		}
	}

	// Also check ~/docker for compose files
	if home, err := os.UserHomeDir(); err == nil {
		dockerDir := filepath.Join(home, "docker")
		if info, err := d.Stat(dockerDir); err == nil && info.IsDir() {
			for _, name := range composePatterns {
				p := filepath.Join(dockerDir, name)
				if _, err := d.Stat(p); err == nil {
					result.ComposeFiles = append(result.ComposeFiles, p)
				}
			}
		}
	}

	return result
}
