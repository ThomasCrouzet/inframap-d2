package collector

import (
	"testing"

	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeCollectorStandard(t *testing.T) {
	infra := model.NewInfrastructure()

	cc := &ComposeCollector{
		Files: []config.ComposeFile{
			{
				Path:   "../../testdata/compose/docker-compose.yml",
				Server: "testserver",
			},
		},
	}

	err := cc.Collect(infra)
	require.NoError(t, err)

	server, ok := infra.Servers["testserver"]
	require.True(t, ok)
	assert.Len(t, server.Services, 2)

	// Check uptime-kuma
	var uptimeKuma *model.Service
	for _, svc := range server.Services {
		if svc.Name == "uptime-kuma" {
			uptimeKuma = svc
			break
		}
	}
	require.NotNil(t, uptimeKuma)
	assert.Equal(t, "louislam/uptime-kuma:1", uptimeKuma.Image)
	assert.Equal(t, model.ServiceTypeContainer, uptimeKuma.Type)
}

func TestComposeCollectorTemplate(t *testing.T) {
	infra := model.NewInfrastructure()

	cc := &ComposeCollector{
		Files: []config.ComposeFile{
			{
				Path:     "../../testdata/compose/template.yml.j2",
				Server:   "atlas",
				Template: true,
			},
		},
	}

	err := cc.Collect(infra)
	require.NoError(t, err)

	server, ok := infra.Servers["atlas"]
	require.True(t, ok)

	var stirling *model.Service
	for _, svc := range server.Services {
		if svc.Name == "stirling-pdf" {
			stirling = svc
			break
		}
	}
	require.NotNil(t, stirling, "stirling-pdf service should be parsed from template")
	assert.Equal(t, "stirlingtools/stirling-pdf:latest", stirling.Image)
}

func TestComposeCollectorScanDir(t *testing.T) {
	infra := model.NewInfrastructure()

	cc := &ComposeCollector{
		ScanDirs: []config.ScanDir{
			{
				Path:   "../../testdata/compose",
				Server: "scanserver",
			},
		},
	}

	err := cc.Collect(infra)
	require.NoError(t, err)

	server, ok := infra.Servers["scanserver"]
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(server.Services), 2, "should find services from scanned compose files")
}

func TestDetectServiceType(t *testing.T) {
	tests := []struct {
		image    string
		name     string
		expected model.ServiceType
	}{
		{"postgres:15-alpine", "db", model.ServiceTypeDatabase},
		{"mysql:8", "database", model.ServiceTypeDatabase},
		{"redis:7", "cache", model.ServiceTypeDatabase},
		{"louislam/uptime-kuma:1", "uptime-kuma", model.ServiceTypeContainer},
		{"nginx:latest", "web", model.ServiceTypeContainer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectServiceType(tt.image, tt.name)
			assert.Equal(t, tt.expected, got)
		})
	}
}
