package cmd

import (
	"fmt"
	"os"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/ramonvermeulen/pre-commit-bump/core/bumper"
	"github.com/ramonvermeulen/pre-commit-bump/core/parser"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available updates without modifying the \".pre-commit-config.yaml\" file",
	Long: `Check for available updates without modifying the ".pre-commit-config.yaml" file.
This command will exit with a non-zero status code if there are updates available.`,
	Run: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) {
	cfg, err := config.FromViper()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	cfg.Logger.Sugar().Debugf("Starting check command - config_path: %s", cfg.PreCommitConfigPath)

	p := parser.NewParser(cfg.Logger)
	bmp := bumper.NewBumper(p, cfg)

	if err := bmp.Check(); err != nil {
		fmt.Fprintf(os.Stderr, "Check failed: %v\n", err)
		os.Exit(1)
	}

	cfg.Logger.Sugar().Info("Check completed successfully, all hooks are up-to-date")
}
