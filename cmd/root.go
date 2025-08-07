package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "pre-commit-bump",
	Short:   "A tool to bump pre-commit hooks",
	Long:    `pre-commit-bump is a command-line tool designed to help you manage and update pre-commit hooks in your projects.`,
	PreRunE: validateGlobalFlags,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.PersistentFlags().StringP(config.FlagConfig, "c", ".pre-commit-config.yaml", "Path to the pre-commit configuration file")
	rootCmd.PersistentFlags().BoolP(config.FlagVerbose, "v", false, "Enable verbose logging output")
	rootCmd.PersistentFlags().StringP(config.FlagAllow, "a", "major", "Version bump type to allow (major, minor, patch), default is 'major'")

	config.BindFlag(rootCmd.PersistentFlags(), config.FlagConfig)
	config.BindFlag(rootCmd.PersistentFlags(), config.FlagVerbose)
	config.BindFlag(rootCmd.PersistentFlags(), config.FlagAllow)
}

// Execute is the entrypoint for the CLI application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// validateGlobalFlags checks the global flags before executing any command
func validateGlobalFlags(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed(config.FlagConfig) {
		configPath, _ := cmd.Flags().GetString(config.FlagConfig)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return err
		}
	}

	if cmd.Flags().Changed(config.FlagAllow) {
		allow, _ := cmd.Flags().GetString(config.FlagAllow)
		allowValues := []string{"major", "minor", "patch"}
		if !slices.Contains(allowValues, allow) {
			return fmt.Errorf("invalid value for --allow: %s. Allowed values are: %v", allow, allowValues)
		}
	}

	return nil
}
