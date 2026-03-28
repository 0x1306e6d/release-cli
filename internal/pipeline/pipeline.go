package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0x1306e6d/release-cli/internal/changelog"
	"github.com/0x1306e6d/release-cli/internal/commits"
	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/detector"
	"github.com/0x1306e6d/release-cli/internal/git"
	"github.com/0x1306e6d/release-cli/internal/propagate"
	"github.com/0x1306e6d/release-cli/internal/publish"
	"github.com/0x1306e6d/release-cli/internal/version"
)

// PackageContext provides monorepo package context for scoped pipeline operations.
// When nil, the pipeline runs in single-project mode.
type PackageContext struct {
	Name      string // Package name (used in commit messages and tag prefix)
	Path      string // Relative path from repo root (e.g., "cli", "workflow/sub")
	TagPrefix string // Tag prefix for namespaced tags (e.g., "cli", "my-app")
	IsForced  bool   // If true, force a patch bump even when no commits found
}

// Options holds the runtime context for a pipeline run.
type Options struct {
	Dir          string
	Config       *config.Config
	DryRun       bool
	BumpOverride *version.BumpType
	Package      *PackageContext // nil for single-project mode
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
	pkg := opts.Package

	// Resolve scoped paths for monorepo.
	detectDir := dir
	var tagPrefix string
	var pathFilter string
	if pkg != nil {
		if pkg.Path != "" {
			detectDir = filepath.Join(dir, pkg.Path)
		}
		tagPrefix = pkg.TagPrefix
		pathFilter = pkg.Path
	}

	// 1. Detect project type.
	registry := detector.DefaultRegistry()
	det, err := registry.Resolve(cfg.Project, detectDir)
	if err != nil {
		return nil, err
	}
	report("Detected project type: %s", det.Name())

	// 2. Read current version.
	prevVer, err := readCurrentVersion(dir, det, detectDir, tagPrefix)
	if err != nil {
		return nil, fmt.Errorf("reading current version: %w", err)
	}
	report("Current version: %s", prevVer.String())

	// 3. Analyze commits.
	fromTag := ""
	if !prevVer.IsZero() {
		fromTag = git.NamespacedTagString(tagPrefix, prevVer.StripPreRelease())
	}

	gitCommits, err := git.LogBetween(dir, fromTag, "HEAD", pathFilter)
	if err != nil {
		return nil, fmt.Errorf("reading commit log: %w", err)
	}

	rawCommits := make([]commits.RawCommit, len(gitCommits))
	for i, c := range gitCommits {
		rawCommits[i] = commits.RawCommit{Hash: c.Hash, Subject: c.Subject, Body: c.Body}
	}

	conv := resolveConvention(cfg)
	parsed, bumpType := commits.Analyze(rawCommits, conv)

	// Apply bump override if provided.
	if opts.BumpOverride != nil {
		bumpType = opts.BumpOverride
		report("Bump override: %s", bumpType.String())
	}

	if bumpType == nil {
		if pkg != nil && pkg.IsForced {
			// Forced release in cascade mode: default to patch bump.
			patch := version.BumpPatch
			bumpType = &patch
			report("No releasable changes found, forced patch bump")
		} else {
			report("No releasable changes found.")
			return nil, nil
		}
	}
	report("Bump type: %s (%d releasable commits)", bumpType.String(), len(parsed))

	// 4. Calculate new version.
	baseVer := prevVer.StripPreRelease()
	newVer := baseVer.Bump(*bumpType)
	report("Version bump: %s → %s", prevVer.CoreString(), newVer.String())

	if opts.DryRun {
		return dryRunReport(cfg, det, prevVer, newVer, parsed, tagPrefix), nil
	}

	// Build hook options for monorepo package context.
	var hookOpts []HookOptions
	if pkg != nil {
		hookOpts = []HookOptions{{PackageName: pkg.Name, PackagePath: pkg.Path}}
	}

	// 5. Pre-bump hook.
	if err := RunHook(dir, cfg.Hooks.PreBump, newVer.String(), prevVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 6. Bump manifest.
	if err := det.WriteVersion(detectDir, detector.Version{Raw: newVer.String()}); err != nil {
		return nil, fmt.Errorf("writing version: %w", err)
	}
	report("✓ Bumped version: %s → %s", prevVer.CoreString(), newVer.String())

	// 7. Propagate.
	if len(cfg.Propagate) > 0 {
		if err := propagate.Propagate(detectDir, cfg.Propagate, newVer.String()); err != nil {
			return nil, err
		}
		report("✓ Propagated version to %d targets", len(cfg.Propagate))
	}

	// 8. Post-bump hook.
	if err := RunHook(dir, cfg.Hooks.PostBump, newVer.String(), prevVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 9. Changelog.
	var changelogContent string
	var releaseBody string
	if cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
		entry := changelog.Generate(newVer.String(), parsed)
		entry.Grouped = cfg.Changes.IsGroupedChangelog()
		if cfg.Changelog.Template != "" {
			changelogContent, err = changelog.RenderCustom(entry, cfg.Changelog.Template)
			if err != nil {
				return nil, err
			}
			releaseBody = changelogContent
		} else {
			changelogContent = entry.Render()
			releaseBody = entry.RenderBody()
		}
		if err := changelog.WriteFile(detectDir, cfg.Changelog.File, changelogContent); err != nil {
			return nil, err
		}
		report("✓ Updated %s", cfg.Changelog.File)
	}

	// 10. Commit.
	var commitLabel string
	if pkg != nil {
		commitLabel = pkg.Name + " " + newVer.String()
	} else {
		commitLabel = newVer.String()
	}
	commitMsg := fmt.Sprintf("Release %s", commitLabel)
	if err := git.CreateCommit(dir, commitMsg, "."); err != nil {
		return nil, fmt.Errorf("creating release commit: %w", err)
	}
	report("✓ Created release commit")

	// 11. Tag.
	tag := git.NamespacedTagString(tagPrefix, newVer)
	if err := git.CreateTag(dir, tag, fmt.Sprintf("Release %s", commitLabel)); err != nil {
		return nil, err
	}
	report("✓ Tagged %s", tag)

	// 11b. Push commit and tag to remote.
	if err := git.Push(dir, tag); err != nil {
		return nil, fmt.Errorf("pushing release: %w", err)
	}
	report("✓ Pushed commit and tag to remote")

	// 12. Pre-publish hook.
	if err := RunHook(dir, cfg.Hooks.PrePublish, newVer.String(), prevVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 13. Publish.
	if err := runGitHubPublish(dir, cfg, tag, newVer.String(), releaseBody); err != nil {
		return nil, err
	}

	// 14. Post-publish hook.
	if err := RunHook(dir, cfg.Hooks.PostPublish, newVer.String(), prevVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 15. Notify (placeholder — implemented in notify package).
	// Notify integration will be wired here once implemented.

	// 16. SNAPSHOT post-release.
	if cfg.Version.Snapshot && det.SnapshotSuffix() != "" {
		snapVer := version.NextSnapshot(newVer, version.NormalizeSnapshotSuffix(det.SnapshotSuffix()))
		if err := det.WriteVersion(detectDir, detector.Version{Raw: snapVer.String()}); err != nil {
			return nil, fmt.Errorf("writing snapshot version: %w", err)
		}
		snapMsg := "Prepare next development iteration"
		if pkg != nil {
			snapMsg += " for " + pkg.Name
		}
		if err := git.CreateCommit(dir, snapMsg, "."); err != nil {
			return nil, fmt.Errorf("creating snapshot commit: %w", err)
		}
		report("✓ Bumped to next development version: %s", snapVer.String())
		if err := git.Push(dir); err != nil {
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

func readCurrentVersion(dir string, det detector.Detector, detectDir, tagPrefix string) (version.Semver, error) {
	// For Go (or any tag-based detector with no manifest), read from git tags.
	v, err := det.ReadVersion(detectDir)
	if err != nil {
		return version.Semver{}, err
	}
	if v.Raw == "" {
		// Tag-based ecosystem: read from git tags (with optional prefix).
		return git.LatestSemverTag(dir, tagPrefix)
	}
	return version.Parse(v.Raw)
}

func resolveConvention(cfg *config.Config) commits.Convention {
	conv, major, minor, patch := cfg.Changes.CommitConventionParams()
	return commits.ResolveConvention(conv, major, minor, patch)
}

func dryRunReport(cfg *config.Config, det detector.Detector, prev, next version.Semver, parsed []commits.ParsedCommit, tagPrefix string) *Result {
	report("[dry-run] Would bump: %s → %s", prev.CoreString(), next.String())
	if len(cfg.Propagate) > 0 {
		report("[dry-run] Would propagate to %d files", len(cfg.Propagate))
	}
	if cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
		report("[dry-run] Would update %s", cfg.Changelog.File)
	}
	tag := git.NamespacedTagString(tagPrefix, next)
	report("[dry-run] Would create tag %s", tag)
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
		TagName:     tag,
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
