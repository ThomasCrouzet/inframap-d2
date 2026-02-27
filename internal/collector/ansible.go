package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(func() RegisteredCollector { return &AnsibleCollector{} })
}

// AnsibleCollector parses Ansible YAML inventory and group_vars.
type AnsibleCollector struct {
	InventoryPath string
	GroupVarsPath string
	PrimaryGroup  string
}

func (ac *AnsibleCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "ansible",
		DisplayName: "Ansible Inventory",
		Description: "Parses Ansible YAML inventory and group_vars for servers and system services",
		ConfigKey:   "ansible",
		DetectHint:  "hosts.yml",
	}
}

func (ac *AnsibleCollector) Enabled(sources map[string]any) bool {
	section, ok := sources["ansible"].(map[string]any)
	if !ok {
		return false
	}
	inv, _ := section["inventory"].(string)
	return inv != ""
}

func (ac *AnsibleCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	if v, ok := section["inventory"].(string); ok {
		ac.InventoryPath = v
	}
	if v, ok := section["group_vars"].(string); ok {
		ac.GroupVarsPath = v
	}
	if v, ok := section["primary_group"].(string); ok {
		ac.PrimaryGroup = v
	}
	return nil
}

func (ac *AnsibleCollector) Validate() []ValidationError {
	var errs []ValidationError
	if ac.InventoryPath != "" {
		if _, err := os.Stat(ac.InventoryPath); err != nil {
			errs = append(errs, ValidationError{
				Field:      "sources.ansible.inventory",
				Message:    fmt.Sprintf("file not found: %s", ac.InventoryPath),
				Suggestion: "check the path or run 'inframap-d2 init' to reconfigure",
			})
		}
	}
	if ac.GroupVarsPath != "" {
		if info, err := os.Stat(ac.GroupVarsPath); err != nil || !info.IsDir() {
			errs = append(errs, ValidationError{
				Field:      "sources.ansible.group_vars",
				Message:    fmt.Sprintf("directory not found: %s", ac.GroupVarsPath),
				Suggestion: "check the path to your group_vars directory",
			})
		}
	}
	return errs
}

// inventoryData is a parsed Ansible inventory.
type inventoryData = map[string]interface{}

// hostEntry represents a single host's variables.
type hostEntry struct {
	AnsibleHost       string `yaml:"ansible_host"`
	AnsibleUser       string `yaml:"ansible_user"`
	ServerType        string `yaml:"server_type"`
	Hostname          string `yaml:"hostname"`
	TailscaleHostname string `yaml:"tailscale_hostname"`
}

func (ac *AnsibleCollector) Collect(infra *model.Infrastructure) error {
	if ac.InventoryPath == "" {
		return nil
	}

	if err := ac.parseInventory(infra); err != nil {
		return fmt.Errorf("parsing ansible inventory: %w", err)
	}

	if ac.GroupVarsPath != "" {
		if err := ac.parseGroupVars(infra); err != nil {
			return fmt.Errorf("parsing group_vars: %w", err)
		}
	}

	return nil
}

func (ac *AnsibleCollector) parseInventory(infra *model.Infrastructure) error {
	data, err := os.ReadFile(ac.InventoryPath)
	if err != nil {
		return err
	}

	var inv inventoryData
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return fmt.Errorf("unmarshal inventory: %w", err)
	}

	// Build a map of bootstrap hosts (to get public IPs)
	bootstrapIPs := make(map[string]string) // tailscale_hostname → public IP
	if bootstrapGroup, ok := inv["bootstrap"]; ok {
		hosts := extractHosts(bootstrapGroup)
		for _, h := range hosts {
			if h.TailscaleHostname != "" && h.AnsibleHost != "" {
				bootstrapIPs[h.TailscaleHostname] = h.AnsibleHost
			}
		}
	}

	// Parse primary group (tailnet) or all groups
	primaryGroup := ac.PrimaryGroup
	if primaryGroup == "" {
		primaryGroup = "tailnet"
	}

	if group, ok := inv[primaryGroup]; ok {
		hosts := extractHosts(group)
		for name, h := range hosts {
			hostname := strings.ToLower(name)
			if h.Hostname != "" {
				hostname = strings.ToLower(h.Hostname)
			}

			stype := model.ServerType(h.ServerType)
			if stype == "" {
				stype = model.ServerTypeLab
			}

			server := &model.Server{
				Hostname: hostname,
				Label:    hostname,
				Type:     stype,
				Online:   true,
			}

			if ip, ok := bootstrapIPs[hostname]; ok {
				server.PublicIP = ip
			}

			server.AnsibleGroups = ac.findGroups(inv, name)
			infra.Servers[hostname] = server
		}
	}

	// Build server groups
	for groupName, groupData := range inv {
		if groupName == "all" {
			continue
		}
		hosts := extractHosts(groupData)
		if len(hosts) == 0 {
			continue
		}
		sg := &model.ServerGroup{
			Name:  groupName,
			Label: groupName,
		}
		for name := range hosts {
			sg.Servers = append(sg.Servers, name)
		}
		infra.ServerGroups[groupName] = sg
	}

	return nil
}

func (ac *AnsibleCollector) parseGroupVars(infra *model.Infrastructure) error {
	// Parse all.yml for global vars
	allPath := filepath.Join(ac.GroupVarsPath, "all.yml")
	if data, err := os.ReadFile(allPath); err == nil {
		var allVars map[string]interface{}
		if err := yaml.Unmarshal(data, &allVars); err == nil {
			ac.extractSystemServices(infra, allVars)
		}
	}

	// Parse tailnet/vars.yml for service_health_checks
	tailnetVarsPath := filepath.Join(ac.GroupVarsPath, "tailnet", "vars.yml")
	if data, err := os.ReadFile(tailnetVarsPath); err == nil {
		var tailnetVars map[string]interface{}
		if err := yaml.Unmarshal(data, &tailnetVars); err == nil {
			ac.extractHealthChecks(infra, tailnetVars)
		}
	}

	return nil
}

func (ac *AnsibleCollector) extractSystemServices(infra *model.Infrastructure, vars map[string]interface{}) {
	// Extract netdata and cockpit ports as system services for all servers
	type sysService struct {
		name string
		key  string
	}
	services := []sysService{
		{"netdata", "netdata_port"},
		{"cockpit", "cockpit_port"},
	}

	for _, ss := range services {
		if portVal, ok := vars[ss.key]; ok {
			port := toInt(portVal)
			if port == 0 {
				continue
			}
			for _, server := range infra.Servers {
				svc := &model.Service{
					Name: ss.name,
					Type: model.ServiceTypeSystem,
					Ports: []model.PortMapping{
						{HostPort: port, ContainerPort: port, Protocol: "tcp"},
					},
				}
				server.AddService(svc)
			}
		}
	}
}

func (ac *AnsibleCollector) extractHealthChecks(infra *model.Infrastructure, vars map[string]interface{}) {
	checks, ok := vars["service_health_checks"]
	if !ok {
		return
	}

	checksMap, ok := checks.(map[string]interface{})
	if !ok {
		return
	}

	for name, checkData := range checksMap {
		checkMap, ok := checkData.(map[string]interface{})
		if !ok {
			continue
		}

		hc := &model.HealthCheck{
			Port:           toInt(checkMap["port"]),
			Path:           toString(checkMap["path"]),
			ExpectedStatus: toInt(checkMap["expected_status"]),
			Timeout:        toInt(checkMap["timeout"]),
		}

		for _, server := range infra.Servers {
			for _, svc := range server.Services {
				if svc.Name == name {
					svc.HealthCheck = hc
				}
			}
		}
	}
}

func (ac *AnsibleCollector) findGroups(inv inventoryData, hostName string) []string {
	var groups []string
	for groupName, groupData := range inv {
		if groupName == "all" {
			continue
		}
		hosts := extractHosts(groupData)
		if _, ok := hosts[hostName]; ok {
			groups = append(groups, groupName)
		}
	}
	return groups
}

// extractHosts pulls host entries from an Ansible group structure.
func extractHosts(group interface{}) map[string]hostEntry {
	result := make(map[string]hostEntry)

	groupMap, ok := group.(map[string]interface{})
	if !ok {
		return result
	}

	hostsRaw, ok := groupMap["hosts"]
	if !ok {
		return result
	}

	hostsMap, ok := hostsRaw.(map[string]interface{})
	if !ok {
		return result
	}

	for name, hostData := range hostsMap {
		var entry hostEntry
		if hostData == nil {
			result[name] = entry
			continue
		}

		hostMap, ok := hostData.(map[string]interface{})
		if !ok {
			result[name] = entry
			continue
		}

		entry.AnsibleHost = toString(hostMap["ansible_host"])
		entry.AnsibleUser = toString(hostMap["ansible_user"])
		entry.ServerType = toString(hostMap["server_type"])
		entry.Hostname = toString(hostMap["hostname"])
		entry.TailscaleHostname = toString(hostMap["tailscale_hostname"])
		result[name] = entry
	}

	return result
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	}
	return 0
}
