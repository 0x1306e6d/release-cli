package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show release status for the current project",
	Long: `Display the current version, last release, commits since last release,
and a preview of the next version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("status: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
