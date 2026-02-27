package config

import "github.com/spf13/viper"

type Config struct {
	Output     string       `mapstructure:"output"`
	Layout     string       `mapstructure:"layout"`
	Direction  string       `mapstructure:"direction"`
	Theme      string       `mapstructure:"theme"`
	Sources    Sources      `mapstructure:"sources"`
	Display    Display      `mapstructure:"display"`
	Render     RenderConfig `mapstructure:"render"`
	RawSources map[string]any
}

type Sources struct {
	Ansible   AnsibleSource   `mapstructure:"ansible"`
	Compose   ComposeSource   `mapstructure:"compose"`
	Tailscale TailscaleSource `mapstructure:"tailscale"`
}

type AnsibleSource struct {
	Inventory    string `mapstructure:"inventory"`
	GroupVars    string `mapstructure:"group_vars"`
	PrimaryGroup string `mapstructure:"primary_group"`
}

type ComposeSource struct {
	Files    []ComposeFile `mapstructure:"files"`
	ScanDirs []ScanDir     `mapstructure:"scan_dirs"`
}

type ComposeFile struct {
	Path     string `mapstructure:"path"`
	Server   string `mapstructure:"server"`
	Template bool   `mapstructure:"template"`
}

type ScanDir struct {
	Path   string `mapstructure:"path"`
	Server string `mapstructure:"server"`
}

type TailscaleSource struct {
	Enabled        bool   `mapstructure:"enabled"`
	JsonFile       string `mapstructure:"json_file"`
	IncludeOffline bool   `mapstructure:"include_offline"`
}

type Display struct {
	ShowDevices bool   `mapstructure:"show_devices"`
	ShowVolumes bool   `mapstructure:"show_volumes"`
	GroupBy     string `mapstructure:"group_by"`
}

type RenderConfig struct {
	DetailLevel string `mapstructure:"detail_level"` // minimal, standard, detailed
	AutoRender  bool   `mapstructure:"auto_render"`
	Format      string `mapstructure:"format"` // svg, png
}

func Load() (*Config, error) {
	cfg := &Config{
		Output:    "infrastructure.d2",
		Layout:    "dagre",
		Direction: "right",
		Theme:     "default",
	}
	cfg.Sources.Tailscale.Enabled = true
	cfg.Display.ShowDevices = true
	cfg.Display.GroupBy = "category"
	cfg.Render.DetailLevel = "standard"
	cfg.Render.Format = "svg"

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Populate RawSources for the registry-based orchestrator
	cfg.RawSources = viper.GetStringMap("sources")

	return cfg, nil
}
