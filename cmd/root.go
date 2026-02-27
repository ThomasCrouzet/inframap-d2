package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "inframap-d2",
	Short: "Generate D2 infrastructure diagrams from multiple sources",
	Long: `inframap-d2 auto-generates D2 infrastructure diagrams by collecting data
from Ansible inventories, Docker Compose files, and Tailscale status.

The output is a .d2 file that can be rendered with: d2 infra.d2 infra.svg`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: inframap.yml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("inframap")
		viper.SetConfigType("yml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		}
	}
}
