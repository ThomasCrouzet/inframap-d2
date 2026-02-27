package collector

import (
	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

// CollectResult holds the result of a single collector run.
type CollectResult struct {
	Name    string
	Skipped bool
	Detail  string
	Err     error
}

// Collect runs all registered collectors and merges the results.
func Collect(cfg *config.Config) (*model.Infrastructure, []CollectResult, error) {
	infra := model.NewInfrastructure()
	rawSources := cfg.RawSources

	var results []CollectResult

	for _, c := range All() {
		meta := c.Metadata()

		if !c.Enabled(rawSources) {
			results = append(results, CollectResult{Name: meta.DisplayName, Skipped: true})
			continue
		}

		// Extract this collector's config section
		section, _ := rawSources[meta.ConfigKey].(map[string]any)
		if err := c.Configure(section); err != nil {
			cerr := &CollectorError{Collector: meta.DisplayName, Err: err}
			results = append(results, CollectResult{Name: meta.DisplayName, Err: cerr})
			return nil, results, cerr
		}

		if err := c.Collect(infra); err != nil {
			cerr := &CollectorError{Collector: meta.DisplayName, Err: err}
			results = append(results, CollectResult{Name: meta.DisplayName, Err: cerr})
			return nil, results, cerr
		}

		results = append(results, CollectResult{Name: meta.DisplayName})
	}

	// Merge and correlate
	Merge(infra)

	return infra, results, nil
}

// CollectLegacy runs collectors using the typed config (for backward compatibility with tests).
func CollectLegacy(cfg *config.Config) (*model.Infrastructure, error) {
	infra := model.NewInfrastructure()

	// Run Ansible collector
	if cfg.Sources.Ansible.Inventory != "" {
		ac := &AnsibleCollector{
			InventoryPath: cfg.Sources.Ansible.Inventory,
			GroupVarsPath: cfg.Sources.Ansible.GroupVars,
			PrimaryGroup:  cfg.Sources.Ansible.PrimaryGroup,
		}
		if err := ac.Collect(infra); err != nil {
			return nil, err
		}
	}

	// Run Docker Compose collector
	cc := &ComposeCollector{
		Files:    cfg.Sources.Compose.Files,
		ScanDirs: cfg.Sources.Compose.ScanDirs,
	}
	if err := cc.Collect(infra); err != nil {
		return nil, err
	}

	// Run Tailscale collector
	if cfg.Sources.Tailscale.Enabled {
		tc := &TailscaleCollector{
			JsonFile:       cfg.Sources.Tailscale.JsonFile,
			IncludeOffline: cfg.Sources.Tailscale.IncludeOffline,
		}
		if err := tc.Collect(infra); err != nil {
			return nil, err
		}
	}

	// Merge and correlate
	Merge(infra)

	return infra, nil
}
