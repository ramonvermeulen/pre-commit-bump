package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available updates without modifying the \".pre-commit-config.yaml\" file",
	Long: `Check for available updates without modifying the ".pre-commit-config.yaml" file.
This command will exit with a non-zero status code if there are updates available.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
