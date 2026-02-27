package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailscaleCollector(t *testing.T) {
	infra := model.NewInfrastructure()

	tc := &TailscaleCollector{
		JsonFile:       "../../testdata/tailscale/status.json",
		IncludeOffline: false,
	}

	err := tc.Collect(infra)
	require.NoError(t, err)

	// Tailnet name
	assert.Equal(t, "user@example", infra.TailnetName)

	// Self (minicore) should be added as server (no server tags, but it's self)
	// Actually minicore has no tags so it becomes a device
	// gateway, atlas, nexus are tagged as server
	assert.Contains(t, infra.Servers, "gateway")
	assert.Contains(t, infra.Servers, "atlas")
	assert.Contains(t, infra.Servers, "nexus")

	// Check gateway enrichment
	gateway := infra.Servers["gateway"]
	assert.Equal(t, "100.64.0.2", gateway.TailscaleIP)
	assert.Equal(t, "linux", gateway.OS)
	assert.True(t, gateway.Online)

	// Devices: minicore (self, no server tag), user-phone, homeassistant
	assert.Contains(t, infra.Devices, "minicore")
	assert.Contains(t, infra.Devices, "user-phone")
	assert.Contains(t, infra.Devices, "homeassistant")

	// Offline laptop should not be included (IncludeOffline=false)
	assert.NotContains(t, infra.Servers, "offline-laptop")
	assert.NotContains(t, infra.Devices, "offline-laptop")
}

func TestTailscaleCollectorIncludeOffline(t *testing.T) {
	infra := model.NewInfrastructure()

	tc := &TailscaleCollector{
		JsonFile:       "../../testdata/tailscale/status.json",
		IncludeOffline: true,
	}

	err := tc.Collect(infra)
	require.NoError(t, err)

	// Offline laptop should be included now
	assert.Contains(t, infra.Devices, "offline-laptop")
}

func TestTailscaleCollectorEnrichesExistingServers(t *testing.T) {
	infra := model.NewInfrastructure()

	// Pre-populate from Ansible
	infra.Servers["gateway"] = &model.Server{
		Hostname: "gateway",
		Label:    "gateway",
		PublicIP: "203.0.113.10",
		Type:     model.ServerTypeProduction,
	}

	tc := &TailscaleCollector{
		JsonFile: "../../testdata/tailscale/status.json",
	}

	err := tc.Collect(infra)
	require.NoError(t, err)

	gateway := infra.Servers["gateway"]
	assert.Equal(t, "100.64.0.2", gateway.TailscaleIP)
	assert.Equal(t, "linux", gateway.OS)
	assert.Equal(t, "203.0.113.10", gateway.PublicIP) // preserved from Ansible
	assert.Equal(t, model.ServerTypeProduction, gateway.Type) // preserved
}
