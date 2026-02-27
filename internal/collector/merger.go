package collector

import (
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

// Merge correlates data across collectors.
func Merge(infra *model.Infrastructure) {
	// Auto-categorize services
	categorizeServices(infra)

	// Build server groups by type
	buildTypeGroups(infra)
}

func categorizeServices(infra *model.Infrastructure) {
	for _, server := range infra.Servers {
		for _, svc := range server.Services {
			if svc.Category == "" {
				svc.Category = model.CategorizeService(svc.Name, svc.Image)
			}
		}
	}
}

func buildTypeGroups(infra *model.Infrastructure) {
	groups := map[model.ServerType]*model.ServerGroup{
		model.ServerTypeProduction: {Name: "production", Label: "Production"},
		model.ServerTypeLab:        {Name: "lab", Label: "Lab Servers"},
		model.ServerTypeLocal:      {Name: "local", Label: "Local"},
		model.ServerTypeCluster:    {Name: "cluster", Label: "Kubernetes"},
		model.ServerTypeHypervisor: {Name: "hypervisor", Label: "Hypervisors"},
	}

	for hostname, server := range infra.Servers {
		if g, ok := groups[server.Type]; ok {
			g.Servers = append(g.Servers, hostname)
		}
	}

	for stype, group := range groups {
		if len(group.Servers) > 0 {
			infra.ServerGroups[string(stype)] = group
		}
	}
}
