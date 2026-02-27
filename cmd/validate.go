package cmd

import (
	"fmt"
	"os"

	"github.com/ThomasCrouzet/inframap-d2/internal/collector"
	"github.com/ThomasCrouzet/inframap-d2/internal/config"
	"github.com/ThomasCrouzet/inframap-d2/internal/ui"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your inframap.yml configuration",
	Long: `Check that all configured sources are valid: files exist, binaries
are available, and paths are correct.`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprint(os.Stderr, ui.FormatError("Failed to load config", err.Error(), "run 'inframap-d2 init' to create a config file"))
		return err
	}

	fmt.Println(ui.Bold("Validating inframap.yml..."))

	rawSources := cfg.RawSources
	passed := 0
	failed := 0

	for _, c := range collector.All() {
		meta := c.Metadata()

		if !c.Enabled(rawSources) {
			continue
		}

		// Configure the collector
		section, _ := rawSources[meta.ConfigKey].(map[string]any)
		if err := c.Configure(section); err != nil {
			ui.ValidationErr(meta.DisplayName, err.Error(), "")
			failed++
			continue
		}

		// Run validation
		errs := c.Validate()
		if len(errs) == 0 {
			ui.ValidationOK(meta.DisplayName, "configuration valid")
			passed++
		} else {
			for _, ve := range errs {
				ui.ValidationErr(ve.Field, ve.Message, ve.Suggestion)
				failed++
			}
		}
	}

	fmt.Println()
	if failed == 0 {
		ui.Success(fmt.Sprintf("%d checks passed, 0 errors", passed))
	} else {
		fmt.Printf("%d checks passed, %d errors\n", passed, failed)
	}

	if failed > 0 {
		return fmt.Errorf("%d validation errors", failed)
	}
	return nil
}
