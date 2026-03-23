package cli

import (
	"fmt"
	"os"

	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/pipeline"
	"github.com/0x1306e6d/release-cli/internal/version"
	"github.com/spf13/cobra"
)

var bumpFlag string

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Execute the release pipeline",
	Long: `Load .release.yaml, detect the project type, and run the full release
pipeline: analyze commits, bump version, update changelog, tag, publish,
and notify.`,
	RunE: runRelease,
}

func init() {
	releaseCmd.Flags().StringVar(&bumpFlag, "bump", "", "Override version bump level (major, minor, patch)")
	rootCmd.AddCommand(releaseCmd)
}

func runRelease(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, warnings, err := config.Load(dir)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}

	var bumpOverride *version.BumpType
	if bumpFlag != "" {
		bt, err := version.ParseBumpType(bumpFlag)
		if err != nil {
			return err
		}
		bumpOverride = &bt
	}

	result, err := pipeline.Run(pipeline.Options{
		Dir:          dir,
		Config:       cfg,
		DryRun:       dryRun,
		BumpOverride: bumpOverride,
	})
	if err != nil {
		return err
	}
	if result == nil {
		return nil // no releasable changes
	}

	fmt.Printf("\n✓ Released %s (tag: %s)\n", result.NewVersion, result.TagName)
	return nil
}
