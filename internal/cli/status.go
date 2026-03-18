package cli

import (
	"fmt"
	"os"

	"github.com/0x1306e6d/release-cli/internal/commits"
	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/detector"
	"github.com/0x1306e6d/release-cli/internal/git"
	"github.com/0x1306e6d/release-cli/internal/version"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show release status for the current project",
	Long: `Display the current version, last release, commits since last release,
and a preview of the next version.`,
	RunE: runStatus,
}

func init() {
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

	registry := detector.DefaultRegistry()
	det, err := registry.Resolve(cfg.Project, dir)
	if err != nil {
		return err
	}

	// Read current version.
	var currentVer version.Semver
	v, verErr := det.ReadVersion(dir)
	if verErr == nil && v.Raw != "" {
		currentVer, _ = version.Parse(v.Raw)
	} else {
		currentVer, _ = git.LatestSemverTag(dir)
	}

	// Get last release tag.
	lastTag, _ := git.LatestSemverTag(dir)

	// Count commits since last release.
	fromTag := ""
	if lastTag.Major != 0 || lastTag.Minor != 0 || lastTag.Patch != 0 {
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

func resolveStatusConvention(cfg *config.Config) commits.Convention {
	switch cfg.Commits.Convention {
	case "angular":
		return &commits.AngularCommits{}
	case "custom":
		return commits.NewCustomCommits(
			cfg.Commits.Types.Major,
			cfg.Commits.Types.Minor,
			cfg.Commits.Types.Patch,
		)
	default:
		return &commits.ConventionalCommits{}
	}
}
