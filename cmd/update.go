package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/ramonvermeulen/pre-commit-bump/core/bumper"
	"github.com/ramonvermeulen/pre-commit-bump/core/io"
	"github.com/ramonvermeulen/pre-commit-bump/core/parser"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates and modify the \".pre-commit-config.yaml\" file",
	Long: `Checks for available updates and modifies the ".pre-commit-config.yaml" file with the latest versions of the hooks. 
Generates a "summary.md" file that can be used to review the changes made.`,
	Run: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP(config.FlagNoSummary, "n", false, "Disable summary generation")
	updateCmd.Flags().BoolP(config.FlagDryRun, "d", false, "Perform a dry run showing only the diff of the \".pre-commit-config.yaml\" file without modifying it")

	config.BindFlag(updateCmd.Flags(), config.FlagNoSummary)
	config.BindFlag(updateCmd.Flags(), config.FlagDryRun)
}

func runUpdate(cmd *cobra.Command, args []string) {
	cfg, err := config.FromViper()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	cfg.Logger.Sugar().Debugf("Starting update command - config_path: %s, dry_run: %t, no_summary: %t",
		cfg.PreCommitConfigPath, cfg.DryRun, cfg.NoSummary)

	filesystem := io.NewOSFileSystem()
	httpClient := &http.Client{
		Timeout: config.DefaultHTTPTimeout,
	}
	resultWriter := io.NewResultWriter(filesystem, cfg.Logger)
	p := parser.NewParser(cfg.Logger)

	bmp := bumper.NewBumper(p, cfg, resultWriter, httpClient)

	if err := bmp.Update(); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	cfg.Logger.Sugar().Info("Update completed successfully")
}
