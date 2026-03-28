package cli

import (
	"fmt"
	"os"

	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/monorepo"
	"github.com/0x1306e6d/release-cli/internal/pipeline"
	"github.com/0x1306e6d/release-cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	bumpFlag     string
	packageFlags []string
	allFlag      bool
)

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
	releaseCmd.Flags().StringArrayVar(&packageFlags, "package", nil, "Release specific package(s) in monorepo mode")
	releaseCmd.Flags().BoolVar(&allFlag, "all", false, "Release all packages (monorepo mode)")
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

	if cfg.IsMonorepo() {
		return runMonorepoRelease(dir, cfg, bumpOverride)
	}

	if len(packageFlags) > 0 {
		return fmt.Errorf(`--package flag requires monorepo mode (use "modules" in config)`)
	}
	if allFlag {
		return fmt.Errorf(`--all flag requires monorepo mode (use "modules" in config)`)
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

func runMonorepoRelease(dir string, cfg *config.Config, bumpOverride *version.BumpType) error {
	if len(packageFlags) == 0 && !allFlag {
		return fmt.Errorf("monorepo mode requires --package <name> or --all")
	}

	// Load the package tree.
	tree, err := monorepo.LoadTree(dir)
	if err != nil {
		return fmt.Errorf("loading package tree: %w", err)
	}
	if err := monorepo.ValidateTree(tree); err != nil {
		return fmt.Errorf("validating package tree: %w", err)
	}

	// Resolve target packages and determine which are forced.
	var targets []*monorepo.PackageNode
	forcedSet := make(map[string]bool)
	if allFlag {
		targets = tree.Flatten()
	} else {
		seen := make(map[string]bool)
		for _, name := range packageFlags {
			nodes, err := tree.Resolve(name)
			if err != nil {
				return err
			}
			for _, n := range nodes {
				forcedSet[n.Name] = true
				if !seen[n.Name] {
					seen[n.Name] = true
					targets = append(targets, n)
				}
			}
		}
	}

	if len(targets) == 0 {
		fmt.Println("No packages to release.")
		return nil
	}

	var pkgContexts []*pipeline.PackageContext
	var pkgConfigs []*config.Config
	for _, t := range targets {
		pkgContexts = append(pkgContexts, &pipeline.PackageContext{
			Name:      t.Name,
			Path:      t.Path,
			TagPrefix: t.TagPrefix(),
			IsForced:  allFlag || forcedSet[t.Name],
		})
		pkgConfigs = append(pkgConfigs, t.Config)
	}

	// Run batched release.
	results, err := pipeline.BatchRelease(dir, pkgContexts, pkgConfigs, dryRun, bumpOverride)
	if err != nil {
		return err
	}
	if results == nil {
		return nil
	}

	fmt.Println()
	for _, r := range results {
		fmt.Printf("✓ Released %s (tag: %s)\n", r.NewVersion, r.TagName)
	}
	return nil
}
