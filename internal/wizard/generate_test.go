package wizard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateConfigMinimal(t *testing.T) {
	answers := WizardAnswers{
		EnableTailscale: true,
		ShowDevices:     true,
		Direction:       "right",
	}

	out, err := GenerateConfig(answers)
	require.NoError(t, err)

	assert.Contains(t, out, "direction: right")
	assert.Contains(t, out, "tailscale:")
	assert.Contains(t, out, "enabled: true")
	assert.Contains(t, out, "show_devices: true")
	assert.NotContains(t, out, "ansible:")
	assert.NotContains(t, out, "compose:")
}

func TestGenerateConfigFull(t *testing.T) {
	answers := WizardAnswers{
		EnableAnsible:    true,
		EnableCompose:    true,
		EnableTailscale:  true,
		AnsibleInventory: "./inventory/hosts.yml",
		AnsibleGroupVars: "./inventory/group_vars",
		AnsiblePrimary:   "tailnet",
		ComposeFiles: []ComposeFileEntry{
			{Path: "./docker/compose.yml", Server: "myserver"},
		},
		ComposeScanDirs: []ComposeScanEntry{
			{Path: "~/docker", Server: "homelab"},
		},
		IncludeOffline: false,
		Direction:      "down",
		ShowDevices:    true,
		GroupBy:        "category",
		DetailLevel:    "detailed",
	}

	out, err := GenerateConfig(answers)
	require.NoError(t, err)

	assert.Contains(t, out, "direction: down")
	assert.Contains(t, out, "inventory: ./inventory/hosts.yml")
	assert.Contains(t, out, "group_vars: ./inventory/group_vars")
	assert.Contains(t, out, "primary_group: tailnet")
	assert.Contains(t, out, "path: ./docker/compose.yml")
	assert.Contains(t, out, "server: myserver")
	assert.Contains(t, out, "path: ~/docker")
	assert.Contains(t, out, "server: homelab")
	assert.Contains(t, out, "detail_level: detailed")
}

func TestGenerateConfigTemplate(t *testing.T) {
	answers := WizardAnswers{
		EnableCompose: true,
		ComposeFiles: []ComposeFileEntry{
			{Path: "./templates/compose.yml.j2", Server: "srv", Template: true},
		},
		ShowDevices: false,
	}

	out, err := GenerateConfig(answers)
	require.NoError(t, err)

	assert.Contains(t, out, "template: true")
	assert.Contains(t, out, "show_devices: false")
}

func TestGenerateConfigDefaults(t *testing.T) {
	answers := WizardAnswers{}
	out, err := GenerateConfig(answers)
	require.NoError(t, err)

	// Should have default values
	assert.Contains(t, out, "direction: right")
	assert.Contains(t, out, "group_by: category")
	assert.Contains(t, out, "detail_level: standard")

	// Count non-empty lines to make sure output is reasonable
	lines := strings.Split(out, "\n")
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	assert.Greater(t, nonEmpty, 3)
}
