package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortainerCollector(t *testing.T) {
	pc := &PortainerCollector{
		Server:   "myhost",
		TestFile: "../../testdata/portainer/containers.json",
	}

	infra := model.NewInfrastructure()
	err := pc.Collect(infra)
	require.NoError(t, err)

	assert.Contains(t, infra.Servers, "myhost")
	server := infra.Servers["myhost"]

	// Should have 4 running containers (stopped-app excluded)
	assert.Len(t, server.Services, 4)

	svcNames := make(map[string]bool)
	for _, svc := range server.Services {
		svcNames[svc.Name] = true
	}
	assert.True(t, svcNames["traefik"])
	assert.True(t, svcNames["jellyfin"])
	assert.True(t, svcNames["radarr"])
	assert.True(t, svcNames["postgres"])
	assert.False(t, svcNames["stopped-app"])
}

func TestPortainerContainerTypes(t *testing.T) {
	pc := &PortainerCollector{
		Server:   "typetest",
		TestFile: "../../testdata/portainer/containers.json",
	}

	infra := model.NewInfrastructure()
	err := pc.Collect(infra)
	require.NoError(t, err)

	server := infra.Servers["typetest"]
	for _, svc := range server.Services {
		if svc.Name == "postgres" {
			assert.Equal(t, model.ServiceTypeDatabase, svc.Type)
		}
	}
}

func TestPortainerCategories(t *testing.T) {
	pc := &PortainerCollector{
		Server:   "cattest",
		TestFile: "../../testdata/portainer/containers.json",
	}

	infra := model.NewInfrastructure()
	err := pc.Collect(infra)
	require.NoError(t, err)

	server := infra.Servers["cattest"]
	for _, svc := range server.Services {
		switch svc.Name {
		case "traefik":
			assert.Equal(t, "proxy", svc.Category)
		case "jellyfin", "radarr":
			assert.Equal(t, "media", svc.Category)
		case "postgres":
			assert.Equal(t, "databases", svc.Category)
		}
	}
}

func TestPortainerPorts(t *testing.T) {
	pc := &PortainerCollector{
		Server:   "porttest",
		TestFile: "../../testdata/portainer/containers.json",
	}

	infra := model.NewInfrastructure()
	err := pc.Collect(infra)
	require.NoError(t, err)

	server := infra.Servers["porttest"]
	for _, svc := range server.Services {
		if svc.Name == "traefik" {
			assert.Len(t, svc.Ports, 2) // 80 and 443
		}
	}
}

func TestPortainerMetadata(t *testing.T) {
	pc := &PortainerCollector{}
	meta := pc.Metadata()
	assert.Equal(t, "portainer", meta.Name)
	assert.Equal(t, "portainer", meta.ConfigKey)
}

func TestPortainerEnabled(t *testing.T) {
	pc := &PortainerCollector{}
	assert.False(t, pc.Enabled(map[string]any{}))
	assert.True(t, pc.Enabled(map[string]any{
		"portainer": map[string]any{
			"url": "https://portainer.local:9443",
		},
	}))
}
