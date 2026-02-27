package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// Run executes the interactive wizard and returns the user's answers.
func Run(detection DetectionResult) (*WizardAnswers, error) {
	answers := &WizardAnswers{
		ShowDevices: true,
		Direction:   "right",
		GroupBy:     "category",
		DetailLevel: "standard",
	}

	// Build detection summary
	var hints []string
	if detection.TailscaleAvailable {
		hints = append(hints, "Tailscale detected")
	}
	if detection.AnsibleInventory != "" {
		hints = append(hints, fmt.Sprintf("Ansible inventory found: %s", detection.AnsibleInventory))
	}
	if len(detection.ComposeFiles) > 0 {
		hints = append(hints, fmt.Sprintf("Compose files found: %s", strings.Join(detection.ComposeFiles, ", ")))
	}

	// Pre-select detected sources
	var preSelected []string
	if detection.AnsibleInventory != "" {
		preSelected = append(preSelected, "ansible")
	}
	if len(detection.ComposeFiles) > 0 {
		preSelected = append(preSelected, "compose")
	}
	if detection.TailscaleAvailable {
		preSelected = append(preSelected, "tailscale")
	}

	// Step 1: Source selection
	var selectedSources []string

	desc := "Select the data sources to include in your infrastructure diagram."
	if len(hints) > 0 {
		desc += "\n\nAuto-detected:\n  " + strings.Join(hints, "\n  ")
	}

	sourceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which sources do you want to enable?").
				Description(desc).
				Options(
					huh.NewOption("Ansible Inventory", "ansible").Selected(contains(preSelected, "ansible")),
					huh.NewOption("Docker Compose", "compose").Selected(contains(preSelected, "compose")),
					huh.NewOption("Tailscale VPN", "tailscale").Selected(contains(preSelected, "tailscale")),
				).
				Value(&selectedSources),
		),
	)

	if err := sourceForm.Run(); err != nil {
		return nil, err
	}

	answers.EnableAnsible = contains(selectedSources, "ansible")
	answers.EnableCompose = contains(selectedSources, "compose")
	answers.EnableTailscale = contains(selectedSources, "tailscale")

	// Step 2: Source-specific config
	var groups []*huh.Group

	if answers.EnableAnsible {
		defaultInv := detection.AnsibleInventory
		if defaultInv == "" {
			defaultInv = "./inventory/hosts.yml"
		}
		answers.AnsibleInventory = defaultInv

		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Ansible inventory path").
				Value(&answers.AnsibleInventory),
			huh.NewInput().
				Title("Ansible group_vars path (optional)").
				Value(&answers.AnsibleGroupVars),
			huh.NewInput().
				Title("Primary group name").
				Description("The main host group to parse").
				Placeholder("tailnet").
				Value(&answers.AnsiblePrimary),
		))
	}

	if answers.EnableCompose {
		var composePath string
		var composeServer string
		if len(detection.ComposeFiles) > 0 {
			composePath = detection.ComposeFiles[0]
		}

		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Compose file or scan directory path").
				Description("Path to a compose file or directory to scan").
				Value(&composePath),
			huh.NewInput().
				Title("Server hostname for these services").
				Description("Which server runs these containers?").
				Value(&composeServer),
		))

		// We'll process these after the form runs
		defer func() {
			if composePath != "" && composeServer != "" {
				// Determine if it's a directory or file
				if strings.HasSuffix(composePath, ".yml") || strings.HasSuffix(composePath, ".yaml") || strings.HasSuffix(composePath, ".j2") {
					isTemplate := strings.HasSuffix(composePath, ".j2")
					answers.ComposeFiles = append(answers.ComposeFiles, ComposeFileEntry{
						Path: composePath, Server: composeServer, Template: isTemplate,
					})
				} else {
					answers.ComposeScanDirs = append(answers.ComposeScanDirs, ComposeScanEntry{
						Path: composePath, Server: composeServer,
					})
				}
			}
		}()
	}

	if answers.EnableTailscale {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Tailscale JSON file (optional)").
				Description("Leave empty to use live 'tailscale status --json'").
				Value(&answers.TailscaleJSON),
			huh.NewConfirm().
				Title("Include offline peers?").
				Value(&answers.IncludeOffline),
		))
	}

	// Step 3: Display options
	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Diagram direction").
			Options(
				huh.NewOption("Right (horizontal)", "right"),
				huh.NewOption("Down (vertical)", "down"),
			).
			Value(&answers.Direction),
		huh.NewSelect[string]().
			Title("Detail level").
			Options(
				huh.NewOption("Minimal — servers and groups only", "minimal"),
				huh.NewOption("Standard — services with ports and icons", "standard"),
				huh.NewOption("Detailed — everything including system services", "detailed"),
			).
			Value(&answers.DetailLevel),
		huh.NewConfirm().
			Title("Show Tailscale devices (phones, laptops)?").
			Value(&answers.ShowDevices),
	))

	if len(groups) > 0 {
		form := huh.NewForm(groups...)
		if err := form.Run(); err != nil {
			return nil, err
		}
	}

	return answers, nil
}

func contains(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}
