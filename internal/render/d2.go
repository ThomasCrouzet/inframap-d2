package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/ThomasCrouzet/inframap-d2/internal/util"
)

// D2Renderer generates D2 diagram text.
type D2Renderer struct {
	DetailLevel string // minimal, standard, detailed
}

func (r *D2Renderer) detail() string {
	if r.DetailLevel == "" {
		return "standard"
	}
	return r.DetailLevel
}

func (r *D2Renderer) Render(infra *model.Infrastructure, cfg *config.Config) string {
	r.DetailLevel = cfg.Render.DetailLevel
	theme := GetTheme(cfg.Theme)
	var b strings.Builder

	direction := cfg.Direction
	if direction == "" {
		direction = "right"
	}

	fmt.Fprintf(&b, "direction: %s\n\n", direction)

	// Tailnet wrapper
	tailnetLabel := "Tailscale VPN"
	if infra.TailnetName != "" {
		tailnetLabel = fmt.Sprintf("Tailscale — %s", infra.TailnetName)
	}

	fmt.Fprintf(&b, "tailnet: %s {\n", util.Quote(tailnetLabel))

	// Render server groups in order
	groupOrder := []model.ServerType{
		model.ServerTypeProduction,
		model.ServerTypeLab,
		model.ServerTypeCluster,
		model.ServerTypeHypervisor,
		model.ServerTypeLocal,
	}

	for _, stype := range groupOrder {
		servers := serversOfType(infra, stype)
		if len(servers) == 0 {
			continue
		}

		group := infra.ServerGroups[string(stype)]
		groupLabel := string(stype)
		if group != nil {
			groupLabel = group.Label
		}

		color := theme.ColorForServerType(stype)
		fmt.Fprintf(&b, "  %s: %s {\n", util.SanitizeID(string(stype)), util.Quote(groupLabel))
		fmt.Fprintf(&b, "    style.fill: %q\n", color.Fill)
		fmt.Fprintf(&b, "    style.stroke: %q\n", color.Stroke)
		b.WriteString("\n")

		for _, server := range servers {
			r.renderServer(&b, server, theme, cfg, "    ")
		}

		b.WriteString("  }\n\n")
	}

	// Render devices
	if cfg.Display.ShowDevices && len(infra.Devices) > 0 && r.detail() != "minimal" {
		r.renderDevices(&b, infra, theme)
	}

	b.WriteString("}\n\n")

	// External connections
	if r.detail() != "minimal" {
		r.renderExternalConnections(&b, infra)
	}

	return b.String()
}

func (r *D2Renderer) renderServer(b *strings.Builder, server *model.Server, theme *Theme, cfg *config.Config, indent string) {
	id := util.SanitizeID(server.Hostname)
	label := server.Hostname
	if server.PublicIP != "" && r.detail() != "minimal" {
		label = fmt.Sprintf("%s — %s", server.Hostname, server.PublicIP)
	}

	fmt.Fprintf(b, "%s%s: %s {\n", indent, id, util.Quote(label))

	// Add icon for OS (standard and detailed only)
	if r.detail() != "minimal" {
		if icon := LookupOSIcon(server.OS); icon != "" {
			fmt.Fprintf(b, "%s  icon: %s\n", indent, icon)
		}
	}

	if server.TailscaleIP != "" && r.detail() != "minimal" {
		fmt.Fprintf(b, "%s  tooltip: %q\n", indent, fmt.Sprintf("Tailscale: %s", server.TailscaleIP))
	}

	if r.detail() != "minimal" {
		// Filter services based on detail level
		services := r.filterServices(server.Services)

		// Grid layout for servers with many services
		if len(services) > 8 {
			fmt.Fprintf(b,"%s  grid-columns: 4\n", indent)
		}

		// Group services by category if local and grouping enabled
		if server.Type == model.ServerTypeLocal && cfg.Display.GroupBy == "category" {
			r.renderGroupedServices(b, server, services, theme, indent+"  ")
		} else {
			r.renderFlatServices(b, services, theme, indent+"  ")
		}
	}

	fmt.Fprintf(b,"%s}\n", indent)
}

// filterServices returns the services to render based on detail level.
func (r *D2Renderer) filterServices(services []*model.Service) []*model.Service {
	if r.detail() == "detailed" {
		return services
	}

	// Standard mode: collapse system services into a summary node
	var filtered []*model.Service
	var systemCount int
	for _, svc := range services {
		if svc.Type == model.ServiceTypeSystem {
			systemCount++
		} else {
			filtered = append(filtered, svc)
		}
	}

	// Add a placeholder for system services if any exist
	if systemCount > 0 {
		filtered = append(filtered, &model.Service{
			Name: fmt.Sprintf("system-summary-%d", systemCount),
			Type: model.ServiceTypeSystem,
		})
	}

	return filtered
}

func (r *D2Renderer) renderFlatServices(b *strings.Builder, services []*model.Service, theme *Theme, indent string) {
	sorted := sortedServices(services)
	for _, svc := range sorted {
		r.renderService(b, svc, theme, indent)
	}
}

func (r *D2Renderer) renderGroupedServices(b *strings.Builder, server *model.Server, services []*model.Service, theme *Theme, indent string) {
	groups := make(map[string][]*model.Service)
	for _, svc := range services {
		cat := svc.Category
		if cat == "" {
			cat = "services"
		}
		groups[cat] = append(groups[cat], svc)
	}

	// Sort group names
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	if len(groupNames) <= 1 {
		// Don't create sub-groups for a single category
		r.renderFlatServices(b, services, theme, indent)
		return
	}

	for _, name := range groupNames {
		svcs := groups[name]
		id := util.SanitizeID(name)
		label := strings.ToUpper(name[:1]) + name[1:]

		fmt.Fprintf(b,"%s%s: %s {\n", indent, id, util.Quote(label))

		for _, svc := range sortedServices(svcs) {
			r.renderService(b, svc, theme, indent+"  ")
		}

		fmt.Fprintf(b,"%s}\n", indent)
	}
}

func (r *D2Renderer) renderService(b *strings.Builder, svc *model.Service, theme *Theme, indent string) {
	// Handle system service summary node
	if svc.Type == model.ServiceTypeSystem && strings.HasPrefix(svc.Name, "system-summary-") {
		count := svc.Name[len("system-summary-"):]
		id := "system-services"
		label := fmt.Sprintf("System (%s)", count)
		color := theme.ColorForElement("system")
		fmt.Fprintf(b,"%s%s: %s {\n", indent, id, util.Quote(label))
		fmt.Fprintf(b,"%s  style.fill: %q\n", indent, color.Fill)
		fmt.Fprintf(b,"%s  style.stroke: %q\n", indent, color.Stroke)
		fmt.Fprintf(b,"%s}\n", indent)
		return
	}

	id := util.SanitizeID(svc.Name)
	label := r.serviceLabel(svc)

	fmt.Fprintf(b,"%s%s: %s", indent, id, util.Quote(label))

	// Inline properties
	props := r.serviceProperties(svc, theme)
	if len(props) > 0 {
		b.WriteString(" {\n")
		for _, prop := range props {
			fmt.Fprintf(b,"%s  %s\n", indent, prop)
		}
		fmt.Fprintf(b,"%s}\n", indent)
	} else {
		b.WriteString("\n")
	}
}

// serviceLabel builds a human-readable label for a service.
func (r *D2Renderer) serviceLabel(svc *model.Service) string {
	// Smart label: use image-derived name if the service name is generic
	displayName := smartServiceName(svc.Name, svc.Image)

	if r.detail() == "detailed" {
		// Show all ports
		if len(svc.Ports) > 0 {
			var portStrs []string
			for _, p := range svc.Ports {
				if p.HostPort > 0 {
					portStrs = append(portStrs, fmt.Sprintf(":%d", p.HostPort))
				}
			}
			if len(portStrs) > 0 {
				return fmt.Sprintf("%s %s", displayName, strings.Join(portStrs, " "))
			}
		}
		return displayName
	}

	// Standard: show first port only, skip :0
	if len(svc.Ports) > 0 && svc.Ports[0].HostPort > 0 {
		return fmt.Sprintf("%s :%d", displayName, svc.Ports[0].HostPort)
	}

	return displayName
}

// smartServiceName returns a better display name if the service name is generic.
func smartServiceName(name, image string) string {
	// Map generic container names to their image-derived names
	imageNames := map[string]string{
		"postgres":  "PostgreSQL",
		"mysql":     "MySQL",
		"mariadb":   "MariaDB",
		"redis":     "Redis",
		"mongo":     "MongoDB",
		"memcached": "Memcached",
		"influxdb":  "InfluxDB",
		"nginx":     "Nginx",
		"traefik":   "Traefik",
		"caddy":     "Caddy",
	}

	// Only apply smart naming for generic names like "db", "cache", "proxy", "web"
	genericNames := map[string]bool{
		"db": true, "database": true, "cache": true, "proxy": true,
		"web": true, "server": true, "app": true, "api": true,
	}

	if genericNames[strings.ToLower(name)] && image != "" {
		imgLower := strings.ToLower(image)
		for key, display := range imageNames {
			if strings.Contains(imgLower, key) {
				return display
			}
		}
	}

	return name
}

func (r *D2Renderer) serviceProperties(svc *model.Service, theme *Theme) []string {
	var props []string

	// Shape for databases
	if svc.Type == model.ServiceTypeDatabase {
		props = append(props, "shape: cylinder")
		color := theme.ColorForElement("database")
		props = append(props, fmt.Sprintf("style.fill: %q", color.Fill))
		props = append(props, fmt.Sprintf("style.stroke: %q", color.Stroke))
	}

	// Shape for VMs and LXC containers
	if svc.Type == model.ServiceTypeVM {
		props = append(props, "shape: rectangle")
	}
	if svc.Type == model.ServiceTypeLXC {
		props = append(props, "shape: hexagon")
	}

	// Shape for system services
	if svc.Type == model.ServiceTypeSystem {
		color := theme.ColorForElement("system")
		props = append(props, fmt.Sprintf("style.fill: %q", color.Fill))
		props = append(props, fmt.Sprintf("style.stroke: %q", color.Stroke))
	}

	// Icon (standard and detailed)
	if r.detail() != "minimal" {
		if icon := LookupIcon(svc.Name, svc.Image); icon != "" {
			props = append(props, fmt.Sprintf("icon: %s", icon))
		}
	}

	return props
}

func (r *D2Renderer) renderDevices(b *strings.Builder, infra *model.Infrastructure, theme *Theme) {
	color := theme.ColorForElement("devices")
	b.WriteString("  devices: \"Other Devices\" {\n")
	fmt.Fprintf(b,"    style.fill: %q\n", color.Fill)
	fmt.Fprintf(b,"    style.stroke: %q\n", color.Stroke)
	b.WriteString("\n")

	devices := sortedDevices(infra.Devices)
	for _, dev := range devices {
		id := util.SanitizeID(dev.Hostname)
		label := dev.Hostname
		if dev.OS != "" && r.detail() == "detailed" {
			label = fmt.Sprintf("%s (%s)", dev.Hostname, dev.OS)
		}

		fmt.Fprintf(b,"    %s: %s", id, util.Quote(label))

		if icon := LookupOSIcon(dev.OS); icon != "" {
			fmt.Fprintf(b," {\n      icon: %s\n    }\n", icon)
		} else {
			b.WriteString("\n")
		}
	}

	b.WriteString("  }\n\n")
}

func (r *D2Renderer) renderExternalConnections(b *strings.Builder, infra *model.Infrastructure) {
	// Check if there's a production server that implies cloudflare
	hasProduction := false
	for _, server := range infra.Servers {
		if server.Type == model.ServerTypeProduction {
			hasProduction = true
			break
		}
	}

	if hasProduction {
		cloudColor := GetTheme("default").ColorForElement("cloud")
		fmt.Fprintf(b,"cloudflare: \"Cloudflare\" {\n  shape: cloud\n  style.fill: %q\n  style.stroke: %q\n}\n", cloudColor.Fill, cloudColor.Stroke)
		b.WriteString("internet: \"Internet\" {\n  shape: cloud\n}\n\n")

		// Connect internet → cloudflare → production servers/services
		connWritten := false
		for _, server := range infra.Servers {
			if server.Type != model.ServerTypeProduction {
				continue
			}
			sid := util.SanitizeID(server.Hostname)

			// Find first non-system service to connect to
			target := fmt.Sprintf("tailnet.production.%s", sid)
			for _, svc := range server.Services {
				if svc.Type == model.ServiceTypeSystem {
					continue
				}
				svcID := util.SanitizeID(svc.Name)
				target = fmt.Sprintf("tailnet.production.%s.%s", sid, svcID)
				break
			}

			if !connWritten {
				b.WriteString("internet -> cloudflare { style.stroke-dash: 3 }\n")
				connWritten = true
			}
			fmt.Fprintf(b,"cloudflare -> %s\n", target)
		}
	}

	// Render internal connections (depends_on)
	if r.detail() != "minimal" {
		for _, server := range infra.Servers {
			sid := util.SanitizeID(server.Hostname)
			groupID := util.SanitizeID(string(server.Type))
			for _, svc := range server.Services {
				svcID := util.SanitizeID(svc.Name)
				for _, dep := range svc.DependsOn {
					depID := util.SanitizeID(dep)
					if r.detail() == "detailed" {
						fmt.Fprintf(b,"tailnet.%s.%s.%s -> tailnet.%s.%s.%s: \"depends_on\" { style.stroke-dash: 3 }\n",
							groupID, sid, svcID, groupID, sid, depID)
					} else {
						fmt.Fprintf(b,"tailnet.%s.%s.%s -> tailnet.%s.%s.%s { style.stroke-dash: 3 }\n",
							groupID, sid, svcID, groupID, sid, depID)
					}
				}
			}
		}
	}
}

func serversOfType(infra *model.Infrastructure, stype model.ServerType) []*model.Server {
	var servers []*model.Server
	for _, s := range infra.Servers {
		if s.Type == stype {
			servers = append(servers, s)
		}
	}
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})
	return servers
}

func sortedServices(services []*model.Service) []*model.Service {
	sorted := make([]*model.Service, len(services))
	copy(sorted, services)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

func sortedDevices(devices map[string]*model.Device) []*model.Device {
	sorted := make([]*model.Device, 0, len(devices))
	for _, d := range devices {
		sorted = append(sorted, d)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Hostname < sorted[j].Hostname
	})
	return sorted
}
