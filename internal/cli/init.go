package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .release.yaml for the current project",
	Long: `Detect the project type from manifest files and generate a .release.yaml
configuration file with sensible defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("init: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
