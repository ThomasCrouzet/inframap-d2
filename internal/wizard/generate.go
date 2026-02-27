package wizard

import (
	"bytes"
	"text/template"
)

// WizardAnswers holds all user responses from the wizard.
type WizardAnswers struct {
	// Sources to enable
	EnableAnsible   bool
	EnableCompose   bool
	EnableTailscale bool

	// Ansible settings
	AnsibleInventory string
	AnsibleGroupVars string
	AnsiblePrimary   string

	// Compose settings
	ComposeScanDirs []ComposeScanEntry
	ComposeFiles    []ComposeFileEntry

	// Tailscale settings
	TailscaleJSON    string
	IncludeOffline   bool

	// Display settings
	Direction   string
	GroupBy     string
	ShowDevices bool
	DetailLevel string
}

// ComposeScanEntry is a directory + server for compose scanning.
type ComposeScanEntry struct {
	Path   string
	Server string
}

// ComposeFileEntry is a file + server for explicit compose files.
type ComposeFileEntry struct {
	Path     string
	Server   string
	Template bool
}

const configTemplate = `# inframap-d2 configuration
# Documentation: https://github.com/ThomasCrouzet/inframap-d2

output: infrastructure.d2
direction: {{ .Direction }}

sources:
{{- if .EnableAnsible }}
  ansible:
    inventory: {{ .AnsibleInventory }}
{{- if .AnsibleGroupVars }}
    group_vars: {{ .AnsibleGroupVars }}
{{- end }}
{{- if .AnsiblePrimary }}
    primary_group: {{ .AnsiblePrimary }}
{{- end }}
{{- end }}

{{- if .EnableCompose }}
  compose:
{{- if .ComposeFiles }}
    files:
{{- range .ComposeFiles }}
      - path: {{ .Path }}
        server: {{ .Server }}
{{- if .Template }}
        template: true
{{- end }}
{{- end }}
{{- end }}
{{- if .ComposeScanDirs }}
    scan_dirs:
{{- range .ComposeScanDirs }}
      - path: {{ .Path }}
        server: {{ .Server }}
{{- end }}
{{- end }}
{{- end }}

{{- if .EnableTailscale }}
  tailscale:
    enabled: true
{{- if .TailscaleJSON }}
    json_file: {{ .TailscaleJSON }}
{{- end }}
    include_offline: {{ if .IncludeOffline }}true{{ else }}false{{ end }}
{{- end }}

display:
  show_devices: {{ if .ShowDevices }}true{{ else }}false{{ end }}
  group_by: {{ .GroupBy }}

render:
  detail_level: {{ .DetailLevel }}
`

// GenerateConfig renders the YAML config from wizard answers.
func GenerateConfig(answers WizardAnswers) (string, error) {
	// Set defaults
	if answers.Direction == "" {
		answers.Direction = "right"
	}
	if answers.GroupBy == "" {
		answers.GroupBy = "category"
	}
	if answers.DetailLevel == "" {
		answers.DetailLevel = "standard"
	}

	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, answers); err != nil {
		return "", err
	}

	return buf.String(), nil
}
