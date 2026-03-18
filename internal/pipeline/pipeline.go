package pipeline

import (
	"fmt"
	"os"

	"github.com/0x1306e6d/release-cli/internal/changelog"
	"github.com/0x1306e6d/release-cli/internal/commits"
	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/detector"
	"github.com/0x1306e6d/release-cli/internal/git"
	"github.com/0x1306e6d/release-cli/internal/propagate"
	"github.com/0x1306e6d/release-cli/internal/publish"
	"github.com/0x1306e6d/release-cli/internal/version"
)

// Options holds the runtime context for a pipeline run.
type Options struct {
	Dir    string
	Config *config.Config
	DryRun bool
}

// Result holds the outcome of a pipeline run.
type Result struct {
	PrevVersion string
	NewVersion  string
	TagName     string
}

// Run executes the full release pipeline.
func Run(opts Options) (*Result, error) {
	cfg := opts.Config
	dir := opts.Dir

	// 1. Detect project type.
	registry := detector.DefaultRegistry()
	det, err := registry.Resolve(cfg.Project, dir)
	if err != nil {
		return nil, err
	}
	report("Detected project type: %s", det.Name())

	// 2. Read current version.
	prevVer, err := readCurrentVersion(dir, det)
	if err != nil {
		return nil, fmt.Errorf("reading current version: %w", err)
	}
	report("Current version: %s", prevVer.String())

	// 3. Analyze commits.
	fromTag := ""
	if prevVer.Major != 0 || prevVer.Minor != 0 || prevVer.Patch != 0 {
		fromTag = prevVer.StripPreRelease().TagString()
	}

	gitCommits, err := git.LogBetween(dir, fromTag, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("reading commit log: %w", err)
	}

	rawCommits := make([]commits.RawCommit, len(gitCommits))
	for i, c := range gitCommits {
		rawCommits[i] = commits.RawCommit{Hash: c.Hash, Subject: c.Subject, Body: c.Body}
	}

	conv := resolveConvention(cfg)
	parsed, bumpType := commits.Analyze(rawCommits, conv)

	if bumpType == nil {
		report("No releasable changes found.")
		return nil, nil
	}
	report("Bump type: %s (%d releasable commits)", bumpType.String(), len(parsed))

	// 4. Calculate new version.
	baseVer := prevVer.StripPreRelease()
	newVer := baseVer.Bump(*bumpType)
	report("Version bump: %s → %s", prevVer.CoreString(), newVer.String())

	if opts.DryRun {
		return dryRunReport(cfg, det, prevVer, newVer, parsed), nil
	}

	// 5. Pre-bump hook.
	if err := RunHook(dir, cfg.Hooks.PreBump, newVer.String(), prevVer.CoreString(), cfg.Project); err != nil {
		return nil, err
	}

	// 6. Bump manifest.
	if err := det.WriteVersion(dir, detector.Version{Raw: newVer.String()}); err != nil {
		return nil, fmt.Errorf("writing version: %w", err)
	}
	report("✓ Bumped version: %s → %s", prevVer.CoreString(), newVer.String())

	// 7. Propagate.
	if len(cfg.Propagate) > 0 {
		if err := propagate.Propagate(dir, cfg.Propagate, newVer.String()); err != nil {
			return nil, err
		}
		report("✓ Propagated version to %d targets", len(cfg.Propagate))
	}

	// 8. Post-bump hook.
	if err := RunHook(dir, cfg.Hooks.PostBump, newVer.String(), prevVer.CoreString(), cfg.Project); err != nil {
		return nil, err
	}

	// 9. Changelog.
	var changelogContent string
	if cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
		entry := changelog.Generate(newVer.String(), parsed)
		if cfg.Changelog.Template != "" {
			changelogContent, err = changelog.RenderCustom(entry, cfg.Changelog.Template)
			if err != nil {
				return nil, err
			}
		} else {
			changelogContent = entry.Render()
		}
		if err := changelog.WriteFile(dir, cfg.Changelog.File, changelogContent); err != nil {
			return nil, err
		}
		report("✓ Updated %s", cfg.Changelog.File)
	}

	// 10. Commit.
	commitMsg := fmt.Sprintf("Release %s", newVer.String())
	if err := git.CreateCommit(dir, commitMsg, "."); err != nil {
		return nil, fmt.Errorf("creating release commit: %w", err)
	}
	report("✓ Created release commit")

	// 11. Tag.
	tag := newVer.TagString()
	if err := git.CreateTag(dir, tag, fmt.Sprintf("Release %s", newVer.String())); err != nil {
		return nil, err
	}
	report("✓ Tagged %s", tag)

	// 11b. Push commit and tag to remote.
	if err := git.Push(dir, tag); err != nil {
		return nil, fmt.Errorf("pushing release: %w", err)
	}
	report("✓ Pushed commit and tag to remote")

	// 12. Pre-publish hook.
	if err := RunHook(dir, cfg.Hooks.PrePublish, newVer.String(), prevVer.CoreString(), cfg.Project); err != nil {
		return nil, err
	}

	// 13. Publish.
	if err := runGitHubPublish(dir, cfg, tag, newVer.String(), changelogContent); err != nil {
		return nil, err
	}

	// 14. Post-publish hook.
	if err := RunHook(dir, cfg.Hooks.PostPublish, newVer.String(), prevVer.CoreString(), cfg.Project); err != nil {
		return nil, err
	}

	// 15. Notify (placeholder — implemented in notify package).
	// Notify integration will be wired here once implemented.

	// 16. SNAPSHOT post-release.
	if cfg.Version.Snapshot && det.SnapshotSuffix() != "" {
		snapVer := version.NextSnapshot(newVer, version.NormalizeSnapshotSuffix(det.SnapshotSuffix()))
		if err := det.WriteVersion(dir, detector.Version{Raw: snapVer.String()}); err != nil {
			return nil, fmt.Errorf("writing snapshot version: %w", err)
		}
		snapMsg := "Prepare next development iteration"
		if err := git.CreateCommit(dir, snapMsg, "."); err != nil {
			return nil, fmt.Errorf("creating snapshot commit: %w", err)
		}
		report("✓ Bumped to next development version: %s", snapVer.String())
		if err := git.Push(dir, ""); err != nil {
			return nil, fmt.Errorf("pushing snapshot commit: %w", err)
		}
		report("✓ Pushed snapshot commit to remote")
	}

	return &Result{
		PrevVersion: prevVer.CoreString(),
		NewVersion:  newVer.String(),
		TagName:     tag,
	}, nil
}

func readCurrentVersion(dir string, det detector.Detector) (version.Semver, error) {
	// For Go (or any tag-based detector with no manifest), read from git tags.
	v, err := det.ReadVersion(dir)
	if err != nil {
		return version.Semver{}, err
	}
	if v.Raw == "" {
		// Tag-based ecosystem: read from git tags.
		return git.LatestSemverTag(dir)
	}
	return version.Parse(v.Raw)
}

func resolveConvention(cfg *config.Config) commits.Convention {
	switch cfg.Commits.Convention {
	case "angular":
		return &commits.AngularCommits{}
	case "freeform":
		return &commits.FreeformCommits{}
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

func dryRunReport(cfg *config.Config, det detector.Detector, prev, next version.Semver, parsed []commits.ParsedCommit) *Result {
	report("[dry-run] Would bump: %s → %s", prev.CoreString(), next.String())
	if len(cfg.Propagate) > 0 {
		report("[dry-run] Would propagate to %d files", len(cfg.Propagate))
	}
	if cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
		report("[dry-run] Would update %s", cfg.Changelog.File)
	}
	report("[dry-run] Would create tag %s", next.TagString())
	if cfg.Publish.GitHub.Enabled == nil || *cfg.Publish.GitHub.Enabled {
		report("[dry-run] Would publish GitHub Release")
	}
	if cfg.Version.Snapshot && det.SnapshotSuffix() != "" {
		snapVer := version.NextSnapshot(next, version.NormalizeSnapshotSuffix(det.SnapshotSuffix()))
		report("[dry-run] Would bump to %s after release", snapVer.String())
	}
	return &Result{
		PrevVersion: prev.CoreString(),
		NewVersion:  next.String(),
		TagName:     next.TagString(),
	}
}

func runGitHubPublish(dir string, cfg *config.Config, tag, ver, changelogBody string) error {
	if cfg.Publish.GitHub.Enabled != nil && !*cfg.Publish.GitHub.Enabled {
		return nil
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		report("⚠ GITHUB_TOKEN not set, skipping GitHub Release")
		return nil
	}

	owner, repo, err := git.RemoteOwnerRepo(dir)
	if err != nil {
		report("⚠ Cannot determine repository owner/name, skipping GitHub Release: %v", err)
		return nil
	}

	pub := &publish.GitHubPublisher{
		Token:     token,
		Owner:     owner,
		Repo:      repo,
		Draft:     cfg.Publish.GitHub.Draft,
		Artifacts: cfg.Publish.GitHub.Artifacts,
		Dir:       dir,
	}

	if err := pub.Publish(publish.ReleaseInfo{
		TagName:       tag,
		Version:       ver,
		ChangelogBody: changelogBody,
		Project:       cfg.Project,
	}); err != nil {
		return fmt.Errorf("publishing GitHub release: %w", err)
	}

	report("✓ Published GitHub Release")
	return nil
}

func report(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
