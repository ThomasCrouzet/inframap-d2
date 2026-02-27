package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/ThomasCrouzet/inframap-d2/internal/util"
	"github.com/compose-spec/compose-go/v2/cli"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	yamlv3 "gopkg.in/yaml.v3"
)

func init() {
	Register(func() RegisteredCollector { return &ComposeCollector{} })
}

// ComposeCollector parses docker-compose files and templates.
type ComposeCollector struct {
	Files    []config.ComposeFile
	ScanDirs []config.ScanDir
}

func (cc *ComposeCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "compose",
		DisplayName: "Docker Compose",
		Description: "Parses docker-compose files and Jinja2 templates for services",
		ConfigKey:   "compose",
		DetectHint:  "docker-compose.yml",
	}
}

func (cc *ComposeCollector) Enabled(sources map[string]any) bool {
	section, ok := sources["compose"].(map[string]any)
	if !ok {
		return false
	}
	// Enabled if files or scan_dirs are configured
	if files, ok := section["files"]; ok {
		if list, ok := files.([]any); ok && len(list) > 0 {
			return true
		}
	}
	if dirs, ok := section["scan_dirs"]; ok {
		if list, ok := dirs.([]any); ok && len(list) > 0 {
			return true
		}
	}
	return false
}

func (cc *ComposeCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	// Parse files
	if filesRaw, ok := section["files"]; ok {
		if list, ok := filesRaw.([]any); ok {
			for _, item := range list {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				cf := config.ComposeFile{}
				if v, ok := m["path"].(string); ok {
					cf.Path = v
				}
				if v, ok := m["server"].(string); ok {
					cf.Server = v
				}
				if v, ok := m["template"].(bool); ok {
					cf.Template = v
				}
				cc.Files = append(cc.Files, cf)
			}
		}
	}
	// Parse scan_dirs
	if dirsRaw, ok := section["scan_dirs"]; ok {
		if list, ok := dirsRaw.([]any); ok {
			for _, item := range list {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				sd := config.ScanDir{}
				if v, ok := m["path"].(string); ok {
					sd.Path = v
				}
				if v, ok := m["server"].(string); ok {
					sd.Server = v
				}
				cc.ScanDirs = append(cc.ScanDirs, sd)
			}
		}
	}
	return nil
}

func (cc *ComposeCollector) Validate() []ValidationError {
	var errs []ValidationError
	for i, f := range cc.Files {
		path := util.ExpandPath(f.Path)
		if _, err := os.Stat(path); err != nil {
			errs = append(errs, ValidationError{
				Field:      fmt.Sprintf("sources.compose.files[%d]", i),
				Message:    fmt.Sprintf("file not found: %s", f.Path),
				Suggestion: "check the path or remove this entry",
			})
		}
	}
	for i, d := range cc.ScanDirs {
		path := util.ExpandPath(d.Path)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			errs = append(errs, ValidationError{
				Field:      fmt.Sprintf("sources.compose.scan_dirs[%d]", i),
				Message:    fmt.Sprintf("directory not found: %s", d.Path),
				Suggestion: "check the path or remove this entry",
			})
		}
	}
	return errs
}

func (cc *ComposeCollector) Collect(infra *model.Infrastructure) error {
	// Process explicit files
	for _, f := range cc.Files {
		path := util.ExpandPath(f.Path)
		if err := cc.parseComposeFile(infra, path, f.Server, f.Template); err != nil {
			return fmt.Errorf("parsing compose file %s: %w", f.Path, err)
		}
	}

	// Scan directories
	for _, dir := range cc.ScanDirs {
		path := util.ExpandPath(dir.Path)
		if err := cc.scanDirectory(infra, path, dir.Server); err != nil {
			return fmt.Errorf("scanning directory %s: %w", dir.Path, err)
		}
	}

	return nil
}

func (cc *ComposeCollector) scanDirectory(infra *model.Infrastructure, dir, server string) error {
	patterns := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if info.IsDir() {
			// Skip hidden directories and common non-relevant dirs
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		for _, pattern := range patterns {
			if info.Name() == pattern {
				if err := cc.parseComposeFile(infra, path, server, false); err != nil {
					// Log but don't fail on individual parse errors
					fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
				}
				return nil
			}
		}
		return nil
	})
}

func (cc *ComposeCollector) parseComposeFile(infra *model.Infrastructure, path, server string, isTemplate bool) error {
	if isTemplate {
		return cc.parseTemplate(infra, path, server)
	}
	return cc.parseStandard(infra, path, server)
}

func (cc *ComposeCollector) parseStandard(infra *model.Infrastructure, path, server string) error {
	ctx := context.Background()

	opts, err := cli.NewProjectOptions(
		[]string{path},
		cli.WithDotEnv,
		cli.WithInterpolation(false),
	)
	if err != nil {
		return fmt.Errorf("project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(ctx, opts)
	if err != nil {
		// Fallback: try manual YAML parse
		return cc.parseFallback(infra, path, server)
	}

	return cc.projectToServices(infra, project, path, server)
}

func (cc *ComposeCollector) parseTemplate(infra *model.Infrastructure, path, server string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Strip Jinja2 expressions
	cleaned := util.StripJinja2(string(data))

	// Write to temp file and parse
	tmpFile, err := os.CreateTemp("", "compose-*.yml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(cleaned); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	return cc.parseFallback(infra, tmpFile.Name(), server)
}

// parseFallback uses raw YAML parsing when compose-go fails.
func (cc *ComposeCollector) parseFallback(infra *model.Infrastructure, path, server string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	// Strip Jinja2 if present
	if strings.Contains(content, "{{") {
		content = util.StripJinja2(content)
	}

	var raw map[string]interface{}
	if err := yamlv3.Unmarshal([]byte(content), &raw); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}

	servicesRaw, ok := raw["services"]
	if !ok {
		return nil
	}
	servicesMap, ok := servicesRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	ensureServer(infra, server)

	for name, svcData := range servicesMap {
		svcMap, ok := svcData.(map[string]interface{})
		if !ok {
			continue
		}

		svc := &model.Service{
			Name:        name,
			Image:       toString(svcMap["image"]),
			Type:        detectServiceType(toString(svcMap["image"]), name),
			ComposeFile: path,
		}

		// Parse ports
		if portsRaw, ok := svcMap["ports"]; ok {
			svc.Ports = parsePorts(portsRaw)
		}

		// Parse networks
		if netsRaw, ok := svcMap["networks"]; ok {
			svc.Networks = parseNetworkNames(netsRaw)
		}

		// Parse depends_on
		if depsRaw, ok := svcMap["depends_on"]; ok {
			svc.DependsOn = parseDependsOn(depsRaw)
		}

		// Parse volumes
		if volsRaw, ok := svcMap["volumes"]; ok {
			svc.Volumes = parseVolumes(volsRaw)
		}

		infra.Servers[server].AddService(svc)
	}

	return nil
}

func (cc *ComposeCollector) projectToServices(infra *model.Infrastructure, project *composetypes.Project, path, server string) error {
	ensureServer(infra, server)

	for _, svc := range project.Services {
		service := &model.Service{
			Name:        svc.Name,
			Image:       svc.Image,
			Type:        detectServiceType(svc.Image, svc.Name),
			ComposeFile: path,
		}

		// Ports
		for _, p := range svc.Ports {
			hostPort, _ := strconv.Atoi(p.Published)
			service.Ports = append(service.Ports, model.PortMapping{
				HostIP:        p.HostIP,
				HostPort:      hostPort,
				ContainerPort: int(p.Target),
				Protocol:      p.Protocol,
			})
		}

		// Networks
		for netName := range svc.Networks {
			service.Networks = append(service.Networks, netName)
		}

		// DependsOn
		for depName := range svc.DependsOn {
			service.DependsOn = append(service.DependsOn, depName)
		}

		// Volumes
		for _, v := range svc.Volumes {
			service.Volumes = append(service.Volumes, model.VolumeMount{
				Source: v.Source,
				Target: v.Target,
			})
		}

		infra.Servers[server].AddService(service)
	}

	return nil
}

func ensureServer(infra *model.Infrastructure, hostname string) {
	if hostname == "" {
		return
	}
	if _, exists := infra.Servers[hostname]; !exists {
		infra.Servers[hostname] = &model.Server{
			Hostname: hostname,
			Label:    hostname,
			Type:     model.ServerTypeLocal,
			Online:   true,
		}
	}
}

func detectServiceType(image, name string) model.ServiceType {
	lower := strings.ToLower(image + " " + name)
	dbKeywords := []string{"postgres", "mysql", "mariadb", "mongo", "redis", "memcached", "influxdb", "sqlite"}
	for _, kw := range dbKeywords {
		if strings.Contains(lower, kw) {
			return model.ServiceTypeDatabase
		}
	}
	return model.ServiceTypeContainer
}

func parsePorts(raw interface{}) []model.PortMapping {
	var ports []model.PortMapping
	switch v := raw.(type) {
	case []interface{}:
		for _, p := range v {
			s := fmt.Sprintf("%v", p)
			// Clean Jinja2 placeholder from port strings
			s = strings.ReplaceAll(s, "PLACEHOLDER:", "")
			if s == "" || s == "PLACEHOLDER" {
				continue
			}
			pm := model.ParsePortMapping(s)
			if pm.HostPort > 0 {
				ports = append(ports, pm)
			}
		}
	}
	return ports
}

func parseNetworkNames(raw interface{}) []string {
	var nets []string
	switch v := raw.(type) {
	case []interface{}:
		for _, n := range v {
			nets = append(nets, fmt.Sprintf("%v", n))
		}
	case map[string]interface{}:
		for name := range v {
			nets = append(nets, name)
		}
	}
	return nets
}

func parseDependsOn(raw interface{}) []string {
	var deps []string
	switch v := raw.(type) {
	case []interface{}:
		for _, d := range v {
			deps = append(deps, fmt.Sprintf("%v", d))
		}
	case map[string]interface{}:
		for name := range v {
			deps = append(deps, name)
		}
	}
	return deps
}

func parseVolumes(raw interface{}) []model.VolumeMount {
	var vols []model.VolumeMount
	switch v := raw.(type) {
	case []interface{}:
		for _, vol := range v {
			s := fmt.Sprintf("%v", vol)
			parts := strings.SplitN(s, ":", 2)
			vm := model.VolumeMount{Source: parts[0]}
			if len(parts) > 1 {
				vm.Target = parts[1]
			}
			vols = append(vols, vm)
		}
	}
	return vols
}

