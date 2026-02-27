package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxmoxCollector(t *testing.T) {
	pc := &ProxmoxCollector{
		TestNodes:     "../../testdata/proxmox/nodes.json",
		TestResources: "../../testdata/proxmox/resources.json",
	}

	infra := model.NewInfrastructure()
	err := pc.Collect(infra)
	require.NoError(t, err)

	// Should have 2 hypervisor servers
	assert.Contains(t, infra.Servers, "pve1")
	assert.Contains(t, infra.Servers, "pve2")

	pve1 := infra.Servers["pve1"]
	assert.Equal(t, model.ServerTypeHypervisor, pve1.Type)
	assert.True(t, pve1.Online)

	// pve1 should have 3 running VMs/LXCs: ubuntu-server, docker-host, pihole
	assert.Len(t, pve1.Services, 3)

	svcNames := make(map[string]bool)
	for _, svc := range pve1.Services {
		svcNames[svc.Name] = true
	}
	assert.True(t, svcNames["ubuntu-server"])
	assert.True(t, svcNames["docker-host"])
	assert.True(t, svcNames["pihole"])

	// Check pihole is LXC type
	for _, svc := range pve1.Services {
		if svc.Name == "pihole" {
			assert.Equal(t, model.ServiceTypeLXC, svc.Type)
		}
		if svc.Name == "ubuntu-server" {
			assert.Equal(t, model.ServiceTypeVM, svc.Type)
		}
	}

	// pve2 should have 1 running VM (truenas), windows-desktop is stopped
	pve2 := infra.Servers["pve2"]
	assert.Len(t, pve2.Services, 1)
	assert.Equal(t, "truenas", pve2.Services[0].Name)
}

func TestProxmoxMetadata(t *testing.T) {
	pc := &ProxmoxCollector{}
	meta := pc.Metadata()
	assert.Equal(t, "proxmox", meta.Name)
	assert.Equal(t, "proxmox", meta.ConfigKey)
}

func TestProxmoxEnabled(t *testing.T) {
	pc := &ProxmoxCollector{}

	// Not enabled without config
	assert.False(t, pc.Enabled(map[string]any{}))

	// Enabled with api_url
	assert.True(t, pc.Enabled(map[string]any{
		"proxmox": map[string]any{
			"api_url": "https://pve.local:8006",
		},
	}))
}
