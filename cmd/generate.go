package cmd

import (
	"fmt"
	"os"

	"github.com/ThomasCrouzet/inframap-d2/internal/collector"
	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/model"
	"github.com/ThomasCrouzet/inframap-d2/internal/render"
	"github.com/ThomasCrouzet/inframap-d2/internal/ui"
	"github.com/spf13/cobra"
)

var (
	outputFile       string
	ansibleInventory string
	ansibleGroupVars string
	composeScanDirs  []string
	composeFiles     []string
	tailscaleEnabled bool
	tailscaleJSON    string
	detailLevel      string
	autoRender       bool
	renderFormat     string
	themeName        string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a D2 infrastructure diagram",
	Long: `Collect infrastructure data from Ansible, Docker Compose, and Tailscale,
then generate a D2 diagram file.`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output D2 file path")
	generateCmd.Flags().StringVar(&ansibleInventory, "ansible-inventory", "", "path to Ansible hosts.yml")
	generateCmd.Flags().StringVar(&ansibleGroupVars, "ansible-group-vars", "", "path to Ansible group_vars/")
	generateCmd.Flags().StringSliceVar(&composeScanDirs, "compose-scan-dir", nil, "directories to scan for compose files (format: path:server)")
	generateCmd.Flags().StringSliceVar(&composeFiles, "compose-file", nil, "compose files (format: path:server)")
	generateCmd.Flags().BoolVar(&tailscaleEnabled, "tailscale", false, "collect Tailscale status")
	generateCmd.Flags().StringVar(&tailscaleJSON, "tailscale-json", "", "path to tailscale status JSON file (instead of live)")
	generateCmd.Flags().StringVar(&detailLevel, "detail", "", "detail level: minimal, standard, detailed")
	generateCmd.Flags().BoolVar(&autoRender, "render", false, "auto-render to SVG/PNG after generating D2 (requires d2)")
	generateCmd.Flags().StringVar(&renderFormat, "format", "", "output format for --render: svg, png (default: svg)")
	generateCmd.Flags().StringVar(&themeName, "theme", "", "color theme: default, dark, monochrome, ocean")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprint(os.Stderr, ui.FormatError("Failed to load config", err.Error(), "run 'inframap-d2 init' to create a config file"))
		return err
	}

	applyFlagOverrides(cfg)

	fmt.Println(ui.Bold("Collecting infrastructure data..."))

	infra, results, err := collector.Collect(cfg)

	// Print collector results
	for _, r := range results {
		if r.Skipped {
			ui.CollectorSkipped(r.Name)
		} else if r.Err != nil {
			fmt.Fprint(os.Stderr, ui.FormatError(r.Name+" failed", r.Err.Error(), ""))
		} else {
			ui.CollectorDone(r.Name, r.Detail)
		}
	}

	if err != nil {
		return err
	}

	output := cfg.Output
	if output == "" {
		output = "infrastructure.d2"
	}

	d2Content := render.RenderD2(infra, cfg)

	if err := os.WriteFile(output, []byte(d2Content), 0644); err != nil {
		fmt.Fprint(os.Stderr, ui.FormatError("Failed to write output", err.Error(), ""))
		return err
	}

	ui.Success(fmt.Sprintf("Generated %s (%d servers, %d services)", output, len(infra.Servers), countServices(infra)))

	// Auto-render if requested
	if cfg.Render.AutoRender {
		if err := autoRenderD2(output, cfg.Render.Format); err != nil {
			fmt.Fprint(os.Stderr, ui.FormatError("Auto-render failed", err.Error(), "install d2: https://d2lang.com/tour/install"))
		}
	}

	return nil
}

func applyFlagOverrides(cfg *config.Config) {
	if outputFile != "" {
		cfg.Output = outputFile
	}
	if ansibleInventory != "" {
		cfg.Sources.Ansible.Inventory = ansibleInventory
	}
	if ansibleGroupVars != "" {
		cfg.Sources.Ansible.GroupVars = ansibleGroupVars
	}
	if tailscaleEnabled {
		cfg.Sources.Tailscale.Enabled = true
	}
	if tailscaleJSON != "" {
		cfg.Sources.Tailscale.JsonFile = tailscaleJSON
		cfg.Sources.Tailscale.Enabled = true
	}
	for _, dir := range composeScanDirs {
		parts := splitColonPair(dir)
		cfg.Sources.Compose.ScanDirs = append(cfg.Sources.Compose.ScanDirs, config.ScanDir{
			Path:   parts[0],
			Server: parts[1],
		})
	}
	for _, f := range composeFiles {
		parts := splitColonPair(f)
		cfg.Sources.Compose.Files = append(cfg.Sources.Compose.Files, config.ComposeFile{
			Path:   parts[0],
			Server: parts[1],
		})
	}
	if detailLevel != "" {
		cfg.Render.DetailLevel = detailLevel
	}
	if autoRender {
		cfg.Render.AutoRender = true
	}
	if renderFormat != "" {
		cfg.Render.Format = renderFormat
	}
	if themeName != "" {
		cfg.Theme = themeName
	}
}

func splitColonPair(s string) [2]string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			return [2]string{s[:i], s[i+1:]}
		}
	}
	return [2]string{s, ""}
}

func countServices(infra *model.Infrastructure) int {
	count := 0
	for _, s := range infra.Servers {
		count += len(s.Services)
	}
	return count
}

func autoRenderD2(d2File, format string) error {
	if format == "" {
		format = "svg"
	}

	// Check if d2 is available
	d2Path, err := findExecutable("d2")
	if err != nil {
		return fmt.Errorf("d2 not found in PATH — install it from https://d2lang.com/tour/install")
	}

	ext := "." + format
	outFile := d2File[:len(d2File)-len(".d2")] + ext

	cmd := execCommand(d2Path, d2File, outFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("d2 render failed: %w", err)
	}

	ui.Success(fmt.Sprintf("Rendered %s", outFile))
	return nil
}
