package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Execute the release pipeline",
	Long: `Load .release.yaml, detect the project type, and run the full release
pipeline: analyze commits, bump version, update changelog, tag, publish,
and notify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("release: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
