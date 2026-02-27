package cmd

import (
	"fmt"
	"os"

	"github.com/ThomasCrouzet/inframap-d2/internal/ui"
	"github.com/ThomasCrouzet/inframap-d2/internal/wizard"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create an inframap.yml config file interactively",
	Long: `Scan your environment for infrastructure sources (Ansible, Docker Compose,
Tailscale) and generate a config file through an interactive wizard.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "inframap.yml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("%s already exists.\n", configPath)
		fmt.Print("Overwrite? [y/N] ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Detect environment
	fmt.Println(ui.Bold("Scanning environment..."))
	detection := wizard.Detect(nil)

	// Run wizard
	answers, err := wizard.Run(detection)
	if err != nil {
		return fmt.Errorf("wizard: %w", err)
	}

	// Generate config
	content, err := wizard.GenerateConfig(*answers)
	if err != nil {
		return fmt.Errorf("generating config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	ui.Success(fmt.Sprintf("Created %s", configPath))
	fmt.Println()
	fmt.Printf("Next step: %s\n", ui.Bold("inframap-d2 generate"))
	fmt.Printf("           %s\n", ui.Hint("or edit inframap.yml to fine-tune your config"))

	return nil
}
