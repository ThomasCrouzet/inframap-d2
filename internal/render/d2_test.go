package render

import (
	"strings"
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestD2RendererBasic(t *testing.T) {
	infra := model.NewInfrastructure()

	infra.Servers["gateway"] = &model.Server{
		Hostname: "gateway",
		Label:    "gateway",
		PublicIP: "203.0.113.10",
		Type:     model.ServerTypeProduction,
		OS:       "linux",
		Online:   true,
		Services: []*model.Service{
			{
				Name:  "galerie",
				Image: "galerie:latest",
				Type:  model.ServiceTypeApp,
				Ports: []model.PortMapping{{HostPort: 3000, ContainerPort: 3000}},
			},
			{
				Name:  "netdata",
				Type:  model.ServiceTypeSystem,
				Ports: []model.PortMapping{{HostPort: 19999, ContainerPort: 19999}},
			},
		},
	}

	infra.Servers["atlas"] = &model.Server{
		Hostname: "atlas",
		Label:    "atlas",
		Type:     model.ServerTypeLab,
		Online:   true,
		Services: []*model.Service{
			{
				Name:  "uptime-kuma",
				Image: "louislam/uptime-kuma:1",
				Type:  model.ServiceTypeContainer,
				Ports: []model.PortMapping{{HostPort: 3001, ContainerPort: 3001}},
			},
		},
	}

	infra.Devices["user-phone"] = &model.Device{
		Hostname: "user-phone",
		OS:       "iOS",
		Online:   true,
	}

	infra.ServerGroups["production"] = &model.ServerGroup{
		Name:    "production",
		Label:   "Production",
		Servers: []string{"gateway"},
	}
	infra.ServerGroups["lab"] = &model.ServerGroup{
		Name:    "lab",
		Label:   "Lab Servers",
		Servers: []string{"atlas"},
	}

	infra.TailnetName = "user@example"

	cfg := &config.Config{
		Direction: "right",
		Theme:     "default",
		Display: config.Display{
			ShowDevices: true,
		},
	}

	output := RenderD2(infra, cfg)

	// Check basic structure
	assert.Contains(t, output, "direction: right")
	assert.Contains(t, output, `tailnet: "Tailscale — user@example"`)
	assert.Contains(t, output, `production: "Production"`)
	assert.Contains(t, output, `lab: "Lab Servers"`)
	assert.Contains(t, output, `gateway: "gateway — 203.0.113.10"`)
	assert.Contains(t, output, `galerie: "galerie :3000"`)
	assert.Contains(t, output, `uptime-kuma: "uptime-kuma :3001"`)
	assert.Contains(t, output, `devices: "Other Devices"`)
	assert.Contains(t, output, `user-phone`)
	assert.Contains(t, output, "cloudflare")
	assert.Contains(t, output, "internet")
}

func TestD2RendererDatabaseShape(t *testing.T) {
	infra := model.NewInfrastructure()

	infra.Servers["gateway"] = &model.Server{
		Hostname: "gateway",
		Type:     model.ServerTypeProduction,
		Services: []*model.Service{
			{
				Name:  "db",
				Image: "postgres:15-alpine",
				Type:  model.ServiceTypeDatabase,
				Ports: []model.PortMapping{{HostPort: 5432, ContainerPort: 5432}},
			},
		},
	}

	infra.ServerGroups["production"] = &model.ServerGroup{
		Name:    "production",
		Label:   "Production",
		Servers: []string{"gateway"},
	}

	cfg := &config.Config{
		Direction: "right",
		Theme:     "default",
	}

	output := RenderD2(infra, cfg)
	assert.Contains(t, output, "shape: cylinder")
}

func TestD2RendererNoDevicesWhenDisabled(t *testing.T) {
	infra := model.NewInfrastructure()
	infra.Devices["phone"] = &model.Device{Hostname: "phone"}

	cfg := &config.Config{
		Direction: "right",
		Theme:     "default",
		Display: config.Display{
			ShowDevices: false,
		},
	}

	output := RenderD2(infra, cfg)
	assert.NotContains(t, output, "devices")
}

func TestD2RendererDependsOn(t *testing.T) {
	infra := model.NewInfrastructure()

	infra.Servers["srv"] = &model.Server{
		Hostname: "srv",
		Type:     model.ServerTypeLab,
		Services: []*model.Service{
			{
				Name:      "web",
				Type:      model.ServiceTypeContainer,
				DependsOn: []string{"db"},
			},
			{
				Name: "db",
				Type: model.ServiceTypeDatabase,
			},
		},
	}
	infra.ServerGroups["lab"] = &model.ServerGroup{
		Name:    "lab",
		Label:   "Lab",
		Servers: []string{"srv"},
	}

	cfg := &config.Config{Direction: "right", Theme: "default"}
	output := RenderD2(infra, cfg)

	assert.True(t, strings.Contains(output, "tailnet.lab.srv.web -> tailnet.lab.srv.db"))
}
