package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates and modify the \".pre-commit-config.yaml\" file",
	Long: `Checks for available updates and modifies the ".pre-commit-config.yaml" file with the latest versions of the hooks. 
Generates a "summary.md" file that can be used to review the changes made.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("no-summary", "n", false, "Disable summary generation")
	updateCmd.Flags().BoolP("dry-run", "d", false, "Perform a dry run showing only the diff of the \".pre-commit-config.yaml\" file without modifying it")
}
