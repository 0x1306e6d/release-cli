package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/0x1306e6d/release-cli/internal/commits"
	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/detector"
	"github.com/0x1306e6d/release-cli/internal/git"
	"github.com/0x1306e6d/release-cli/internal/monorepo"
	"github.com/0x1306e6d/release-cli/internal/version"
	"github.com/spf13/cobra"
)

var statusPackageFlag string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show release status for the current project",
	Long: `Display the current version, last release, commits since last release,
and a preview of the next version.`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&statusPackageFlag, "package", "", "Show status for a specific package (monorepo mode)")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	if cfg.IsMonorepo() {
		return runMonorepoStatus(dir)
	}

	if statusPackageFlag != "" {
		return fmt.Errorf(`--package flag requires monorepo mode (use "modules" in config)`)
	}

	return showSingleProjectStatus(dir, cfg)
}

func showSingleProjectStatus(dir string, cfg *config.Config) error {
	registry := detector.DefaultRegistry()
	det, err := registry.Resolve(cfg.Project, dir)
	if err != nil {
		return err
	}

	// Get last release tag.
	lastTag, _ := git.LatestSemverTag(dir)

	// Read current version.
	var currentVer version.Semver
	v, verErr := det.ReadVersion(dir)
	if verErr == nil && v.Raw != "" {
		currentVer, _ = version.Parse(v.Raw)
	} else {
		currentVer = lastTag
	}

	// Count commits since last release.
	fromTag := ""
	if !lastTag.IsZero() {
		fromTag = lastTag.StripPreRelease().TagString()
	}
	gitCommits, _ := git.LogBetween(dir, fromTag, "HEAD")

	// Analyze for next version preview.
	rawCommits := make([]commits.RawCommit, len(gitCommits))
	for i, c := range gitCommits {
		rawCommits[i] = commits.RawCommit{Hash: c.Hash, Subject: c.Subject, Body: c.Body}
	}
	conv := resolveStatusConvention(cfg)
	_, bumpType := commits.Analyze(rawCommits, conv)

	fmt.Printf("Project:     %s (%s)\n", cfg.Project, det.Name())
	fmt.Printf("Current:     %s\n", currentVer.String())
	fmt.Printf("Last release: %s\n", lastTag.TagString())
	fmt.Printf("Commits since: %d\n", len(gitCommits))

	if bumpType != nil {
		nextVer := currentVer.StripPreRelease().Bump(*bumpType)
		fmt.Printf("Next version: %s (%s bump)\n", nextVer.String(), bumpType.String())
	} else {
		fmt.Println("Next version: (no releasable changes)")
	}

	return nil
}

func runMonorepoStatus(dir string) error {
	tree, err := monorepo.LoadTree(dir)
	if err != nil {
		return fmt.Errorf("loading package tree: %w", err)
	}

	// Single package status.
	if statusPackageFlag != "" {
		node := tree.Find(statusPackageFlag)
		if node == nil {
			return fmt.Errorf("package %q not found in config", statusPackageFlag)
		}
		if node.Config.IsContainer() {
			return fmt.Errorf("package %q is a container (no project declared); target one of its child packages or omit --package to see all", statusPackageFlag)
		}

		detectDir := dir
		if node.Path != "" {
			detectDir = filepath.Join(dir, node.Path)
		}

		return showPackageStatus(dir, detectDir, node)
	}

	// Summary table for all packages.
	allNodes := tree.Flatten()
	registry := detector.DefaultRegistry()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "PACKAGE\tVERSION\tPENDING\tNEXT BUMP")

	for _, node := range allNodes {
		if node.Config.IsContainer() {
			continue
		}

		detectDir := dir
		if node.Path != "" {
			detectDir = filepath.Join(dir, node.Path)
		}

		det, err := registry.Resolve(node.Config.Project, detectDir)
		if err != nil {
			_, _ = fmt.Fprintf(w, "%s\t(error)\t-\t-\n", node.Name)
			continue
		}

		var currentVer version.Semver
		v, verErr := det.ReadVersion(detectDir)
		if verErr == nil && v.Raw != "" {
			currentVer, _ = version.Parse(v.Raw)
		} else {
			currentVer, _ = git.LatestSemverTag(dir, node.TagPrefix())
		}

		fromTag := ""
		if !currentVer.IsZero() {
			fromTag = git.NamespacedTagString(node.TagPrefix(), currentVer.StripPreRelease())
		}

		gitCommits, _ := git.LogBetween(dir, fromTag, "HEAD", node.Path)

		rawCommits := make([]commits.RawCommit, len(gitCommits))
		for i, c := range gitCommits {
			rawCommits[i] = commits.RawCommit{Hash: c.Hash, Subject: c.Subject, Body: c.Body}
		}
		conv := resolveStatusConvention(node.Config)
		_, bumpType := commits.Analyze(rawCommits, conv)

		bumpStr := "-"
		if bumpType != nil {
			bumpStr = bumpType.String()
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", node.Name, currentVer.String(), len(gitCommits), bumpStr)
	}

	return w.Flush()
}

func showPackageStatus(dir, detectDir string, node *monorepo.PackageNode) error {
	registry := detector.DefaultRegistry()
	det, err := registry.Resolve(node.Config.Project, detectDir)
	if err != nil {
		return err
	}

	var currentVer version.Semver
	v, verErr := det.ReadVersion(detectDir)
	if verErr == nil && v.Raw != "" {
		currentVer, _ = version.Parse(v.Raw)
	} else {
		currentVer, _ = git.LatestSemverTag(dir, node.TagPrefix())
	}

	fromTag := ""
	if currentVer.Major != 0 || currentVer.Minor != 0 || currentVer.Patch != 0 {
		fromTag = git.NamespacedTagString(node.TagPrefix(), currentVer.StripPreRelease())
	}

	gitCommits, _ := git.LogBetween(dir, fromTag, "HEAD", node.Path)

	rawCommits := make([]commits.RawCommit, len(gitCommits))
	for i, c := range gitCommits {
		rawCommits[i] = commits.RawCommit{Hash: c.Hash, Subject: c.Subject, Body: c.Body}
	}
	conv := resolveStatusConvention(node.Config)
	_, bumpType := commits.Analyze(rawCommits, conv)

	fmt.Printf("Package:     %s\n", node.Name)
	fmt.Printf("Project:     %s (%s)\n", node.Config.Project, det.Name())
	fmt.Printf("Path:        %s\n", node.Path)
	fmt.Printf("Current:     %s\n", currentVer.String())
	lastTag := git.NamespacedTagString(node.TagPrefix(), currentVer.StripPreRelease())
	fmt.Printf("Last release: %s\n", lastTag)
	fmt.Printf("Commits since: %d\n", len(gitCommits))

	if bumpType != nil {
		nextVer := currentVer.StripPreRelease().Bump(*bumpType)
		fmt.Printf("Next version: %s (%s bump)\n", nextVer.String(), bumpType.String())
	} else {
		fmt.Println("Next version: (no releasable changes)")
	}

	return nil
}

func resolveStatusConvention(cfg *config.Config) commits.Convention {
	conv, major, minor, patch := cfg.Changes.CommitConventionParams()
	return commits.ResolveConvention(conv, major, minor, patch)
}
