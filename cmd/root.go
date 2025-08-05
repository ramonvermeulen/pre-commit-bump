package cmd

import (
	"os"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pre-commit-bump",
	Short: "A tool to bump pre-commit hooks",
	Long:  `pre-commit-bump is a command-line tool designed to help you manage and update pre-commit hooks in your projects.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.PersistentFlags().StringP(config.FlagConfig, "c", ".pre-commit-config.yaml", "Path to the pre-commit configuration file")
	rootCmd.PersistentFlags().BoolP(config.FlagVerbose, "v", false, "Enable verbose logging output")

	config.BindFlag(rootCmd.PersistentFlags(), config.FlagConfig)
	config.BindFlag(rootCmd.PersistentFlags(), config.FlagVerbose)
}

// Execute is the entrypoint for the CLI application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
