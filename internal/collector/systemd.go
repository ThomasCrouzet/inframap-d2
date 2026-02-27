package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

func init() {
	Register(func() RegisteredCollector { return &SystemdCollector{} })
}

// SystemdCollector collects running systemd services.
type SystemdCollector struct {
	Servers []systemdServer
}

type systemdServer struct {
	Host     string
	SSH      string   // user@host for remote execution
	Filter   []string // include only these service names
	Exclude  []string // exclude these service names
	TestFile string   // path to test JSON data
}

func (sc *SystemdCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "systemd",
		DisplayName: "systemd Services",
		Description: "Collects running systemd services from local or remote servers",
		ConfigKey:   "systemd",
		DetectHint:  "systemctl",
	}
}

func (sc *SystemdCollector) Enabled(sources map[string]any) bool {
	section, ok := sources["systemd"].(map[string]any)
	if !ok {
		return false
	}
	servers, ok := section["servers"].([]any)
	return ok && len(servers) > 0
}

func (sc *SystemdCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	serversRaw, ok := section["servers"].([]any)
	if !ok {
		return nil
	}
	for _, item := range serversRaw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		srv := systemdServer{}
		if v, ok := m["host"].(string); ok {
			srv.Host = v
		}
		if v, ok := m["ssh"].(string); ok {
			srv.SSH = v
		}
		if v, ok := m["filter"].([]any); ok {
			for _, f := range v {
				if s, ok := f.(string); ok {
					srv.Filter = append(srv.Filter, s)
				}
			}
		}
		if v, ok := m["exclude"].([]any); ok {
			for _, e := range v {
				if s, ok := e.(string); ok {
					srv.Exclude = append(srv.Exclude, s)
				}
			}
		}
		if v, ok := m["test_file"].(string); ok {
			srv.TestFile = v
		}
		sc.Servers = append(sc.Servers, srv)
	}
	return nil
}

func (sc *SystemdCollector) Validate() []ValidationError {
	var errs []ValidationError
	for i, srv := range sc.Servers {
		if srv.Host == "" {
			errs = append(errs, ValidationError{
				Field:      fmt.Sprintf("sources.systemd.servers[%d].host", i),
				Message:    "host is required",
				Suggestion: "set the hostname for this server",
			})
		}
	}
	return errs
}

type systemdUnit struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

func (sc *SystemdCollector) Collect(infra *model.Infrastructure) error {
	for _, srv := range sc.Servers {
		units, err := sc.getUnits(srv)
		if err != nil {
			return fmt.Errorf("getting units for %s: %w", srv.Host, err)
		}

		// Ensure server exists
		server, exists := infra.Servers[srv.Host]
		if !exists {
			server = &model.Server{
				Hostname: srv.Host,
				Label:    srv.Host,
				Type:     model.ServerTypeLab,
				Online:   true,
			}
			infra.Servers[srv.Host] = server
		}

		for _, unit := range units {
			name := strings.TrimSuffix(unit.Unit, ".service")

			// Apply filters
			if len(srv.Filter) > 0 && !matchesAny(name, srv.Filter) {
				continue
			}
			if matchesAny(name, srv.Exclude) {
				continue
			}

			svcType := model.ServiceTypeSystem
			if detectServiceType("", name) == model.ServiceTypeDatabase {
				svcType = model.ServiceTypeDatabase
			}

			svc := &model.Service{
				Name: name,
				Type: svcType,
			}

			server.AddService(svc)
		}
	}

	return nil
}

func (sc *SystemdCollector) getUnits(srv systemdServer) ([]systemdUnit, error) {
	if srv.TestFile != "" {
		data, err := os.ReadFile(srv.TestFile)
		if err != nil {
			return nil, err
		}
		var units []systemdUnit
		if err := json.Unmarshal(data, &units); err != nil {
			return nil, err
		}
		return units, nil
	}

	args := []string{"list-units", "--type=service", "--state=running", "--output=json"}

	var cmd *exec.Cmd
	if srv.SSH != "" {
		sshArgs := []string{srv.SSH, "systemctl"}
		sshArgs = append(sshArgs, args...)
		cmd = exec.Command("ssh", sshArgs...)
	} else {
		cmd = exec.Command("systemctl", args...)
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("systemctl: %w", err)
	}

	var units []systemdUnit
	if err := json.Unmarshal(out, &units); err != nil {
		return nil, fmt.Errorf("parsing systemctl output: %w", err)
	}
	return units, nil
}

func matchesAny(name string, patterns []string) bool {
	lower := strings.ToLower(name)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
