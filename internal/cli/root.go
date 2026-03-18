package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var (
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "release-cli",
	Short: "Automate your release workflow from the command line",
	Long: `release-cli detects your project type and executes a consistent release
pipeline — version bump, changelog, tagging, publishing, and notifications.

Configure via .release.yaml in your project root.`,
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview all steps without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// IsDryRun returns whether the --dry-run flag is set.
func IsDryRun() bool {
	return dryRun
}

// IsVerbose returns whether the --verbose flag is set.
func IsVerbose() bool {
	return verbose
}

// Verbosef prints a formatted message when verbose mode is enabled.
func Verbosef(format string, args ...any) {
	if verbose {
		fmt.Printf(format, args...)
	}
}
