package cmd

import (
	"os"

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

// Execute is the entrypoint for the CLI application
func Execute() {
	rootCmd.PersistentFlags().StringP("config", "c", ".pre-commit-config.yaml", "Path to the pre-commit configuration file")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
