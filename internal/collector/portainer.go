package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

func init() {
	Register(func() RegisteredCollector { return &PortainerCollector{} })
}

// PortainerCollector collects containers from a Portainer instance.
type PortainerCollector struct {
	URL      string
	APIKey   string
	Endpoint int
	Server   string // hostname to assign containers to
	// TestFile for testing (bypasses HTTP calls)
	TestFile string
}

func (pc *PortainerCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "portainer",
		DisplayName: "Portainer",
		Description: "Collects containers from Portainer via its API",
		ConfigKey:   "portainer",
		DetectHint:  "",
	}
}

func (pc *PortainerCollector) Enabled(sources map[string]any) bool {
	section, ok := sources["portainer"].(map[string]any)
	if !ok {
		return false
	}
	url, _ := section["url"].(string)
	return url != ""
}

func (pc *PortainerCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	if v, ok := section["url"].(string); ok {
		pc.URL = v
	}
	if v, ok := section["api_key"].(string); ok {
		pc.APIKey = v
	}
	if pc.APIKey == "" {
		pc.APIKey = os.Getenv("INFRAMAP_PORTAINER_API_KEY")
	}
	if v, ok := section["endpoint"].(int); ok {
		pc.Endpoint = v
	}
	// Handle float64 from YAML unmarshaling
	if v, ok := section["endpoint"].(float64); ok {
		pc.Endpoint = int(v)
	}
	if v, ok := section["server"].(string); ok {
		pc.Server = v
	}
	if v, ok := section["test_file"].(string); ok {
		pc.TestFile = v
	}
	if pc.Endpoint == 0 {
		pc.Endpoint = 1
	}
	return nil
}

func (pc *PortainerCollector) Validate() []ValidationError {
	var errs []ValidationError
	if pc.URL == "" {
		errs = append(errs, ValidationError{
			Field:      "sources.portainer.url",
			Message:    "url is required",
			Suggestion: "set the URL of your Portainer instance, e.g. https://portainer.local:9443",
		})
	}
	if pc.APIKey == "" {
		errs = append(errs, ValidationError{
			Field:      "sources.portainer.api_key",
			Message:    "api_key is required",
			Suggestion: "create an API key in Portainer: User Settings → Access tokens",
		})
	}
	return errs
}

type portainerContainer struct {
	ID     string                 `json:"Id"`
	Names  []string               `json:"Names"`
	Image  string                 `json:"Image"`
	State  string                 `json:"State"`
	Ports  []portainerPort        `json:"Ports"`
	Labels map[string]string      `json:"Labels"`
}

type portainerPort struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

func (pc *PortainerCollector) Collect(infra *model.Infrastructure) error {
	containers, err := pc.getContainers()
	if err != nil {
		return fmt.Errorf("getting containers: %w", err)
	}

	serverName := pc.Server
	if serverName == "" {
		serverName = "portainer"
	}

	// Ensure server exists
	server, exists := infra.Servers[serverName]
	if !exists {
		server = &model.Server{
			Hostname: serverName,
			Label:    serverName,
			Type:     model.ServerTypeLab,
			Online:   true,
		}
		infra.Servers[serverName] = server
	}

	for _, c := range containers {
		if c.State != "running" {
			continue
		}

		name := containerName(c.Names)

		svc := &model.Service{
			Name:  name,
			Image: c.Image,
			Type:  detectServiceType(c.Image, name),
		}

		// Ports
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				svc.Ports = append(svc.Ports, model.PortMapping{
					HostPort:      p.PublicPort,
					ContainerPort: p.PrivatePort,
					Protocol:      p.Type,
				})
			}
		}

		// Use compose project as category if available
		if project, ok := c.Labels["com.docker.compose.project"]; ok {
			svc.Category = project
		}

		server.AddService(svc)
	}

	return nil
}

func (pc *PortainerCollector) getContainers() ([]portainerContainer, error) {
	if pc.TestFile != "" {
		data, err := os.ReadFile(pc.TestFile)
		if err != nil {
			return nil, err
		}
		var containers []portainerContainer
		if err := json.Unmarshal(data, &containers); err != nil {
			return nil, err
		}
		return containers, nil
	}

	url := fmt.Sprintf("%s/api/endpoints/%d/docker/containers/json?all=false", pc.URL, pc.Endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", pc.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("portainer API returned %d: %s", resp.StatusCode, string(body))
	}

	var containers []portainerContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}
	return containers, nil
}

// containerName extracts a clean name from Docker container names (removes leading /).
func containerName(names []string) string {
	if len(names) == 0 {
		return "unknown"
	}
	return strings.TrimPrefix(names[0], "/")
}
