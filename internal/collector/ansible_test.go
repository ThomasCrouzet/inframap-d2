package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnsibleCollector(t *testing.T) {
	infra := model.NewInfrastructure()

	ac := &AnsibleCollector{
		InventoryPath: "../../testdata/ansible/hosts.yml",
		GroupVarsPath: "../../testdata/ansible/group_vars",
		PrimaryGroup:  "tailnet",
	}

	err := ac.Collect(infra)
	require.NoError(t, err)

	// Should have 3 servers
	assert.Len(t, infra.Servers, 3)

	// Check gateway
	gateway, ok := infra.Servers["gateway"]
	require.True(t, ok, "gateway server should exist")
	assert.Equal(t, "gateway", gateway.Hostname)
	assert.Equal(t, model.ServerTypeProduction, gateway.Type)
	assert.Equal(t, "203.0.113.10", gateway.PublicIP)

	// Check atlas
	atlas, ok := infra.Servers["atlas"]
	require.True(t, ok, "atlas server should exist")
	assert.Equal(t, model.ServerTypeLab, atlas.Type)
	assert.Equal(t, "203.0.113.20", atlas.PublicIP)

	// Check nexus
	nexus, ok := infra.Servers["nexus"]
	require.True(t, ok, "nexus server should exist")
	assert.Equal(t, model.ServerTypeLab, nexus.Type)
	assert.Equal(t, "203.0.113.30", nexus.PublicIP)

	// Check system services (netdata, cockpit) from all.yml
	for _, server := range infra.Servers {
		var hasNetdata, hasCockpit bool
		for _, svc := range server.Services {
			if svc.Name == "netdata" {
				hasNetdata = true
				assert.Equal(t, model.ServiceTypeSystem, svc.Type)
				assert.Equal(t, 19999, svc.Ports[0].HostPort)
			}
			if svc.Name == "cockpit" {
				hasCockpit = true
				assert.Equal(t, model.ServiceTypeSystem, svc.Type)
				assert.Equal(t, 9090, svc.Ports[0].HostPort)
			}
		}
		assert.True(t, hasNetdata, "server %s should have netdata", server.Hostname)
		assert.True(t, hasCockpit, "server %s should have cockpit", server.Hostname)
	}

	// Check server groups
	assert.Contains(t, infra.ServerGroups, "tailnet")
	assert.Contains(t, infra.ServerGroups, "bootstrap")
}

func TestAnsibleCollectorNoFile(t *testing.T) {
	infra := model.NewInfrastructure()

	ac := &AnsibleCollector{
		InventoryPath: "",
	}

	err := ac.Collect(infra)
	assert.NoError(t, err)
	assert.Empty(t, infra.Servers)
}
