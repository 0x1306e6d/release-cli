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
	gh "github.com/0x1306e6d/release-cli/internal/github"
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
	baseVer, fromTag, err := resolveReleaseBase(dir, prevVer, tagPrefix)
	if err != nil {
		return nil, fmt.Errorf("resolving release base: %w", err)
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
	newVer := baseVer.Bump(*bumpType)
	report("Version bump: %s → %s", baseVer.CoreString(), newVer.String())

	packageName := ""
	if pkg != nil {
		packageName = pkg.Name
	}
	projectName := cfg.Name
	if projectName == "" {
		projectName = cfg.Project
	}
	nextVer := ""
	if cfg.Version.Snapshot && det.SnapshotSuffix() != "" {
		nextVer = version.NextSnapshot(newVer, version.NormalizeSnapshotSuffix(det.SnapshotSuffix())).String()
	}
	messageData := commitMessageData{
		ReleaseVersion: newVer.String(),
		NextVersion:    nextVer,
		Project:        projectName,
		Package:        packageName,
	}
	commitMsg, err := renderCommitMessage("release", cfg.Commit.Release, defaultReleaseCommitMessage(packageName, newVer.String()), messageData)
	if err != nil {
		return nil, err
	}
	var snapMsg string
	if nextVer != "" {
		snapMsg, err = renderCommitMessage("next", cfg.Commit.Next, defaultNextCommitMessage(packageName), messageData)
		if err != nil {
			return nil, err
		}
	}

	if opts.DryRun {
		return dryRunReport(cfg, det, baseVer, newVer, parsed, tagPrefix), nil
	}

	// Build hook options for monorepo package context.
	var hookOpts []HookOptions
	if pkg != nil {
		hookOpts = []HookOptions{{PackageName: pkg.Name, PackagePath: pkg.Path}}
	}

	// 5. Pre-bump hook.
	if err := RunHook(dir, cfg.Hooks.PreBump, newVer.String(), baseVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 6. Bump manifest.
	if err := det.WriteVersion(detectDir, detector.Version{Raw: newVer.String()}); err != nil {
		return nil, fmt.Errorf("writing version: %w", err)
	}
	report("✓ Bumped version: %s → %s", baseVer.CoreString(), newVer.String())

	// 7. Propagate.
	if len(cfg.Propagate) > 0 {
		if err := propagate.Propagate(detectDir, cfg.Propagate, newVer.String()); err != nil {
			return nil, err
		}
		report("✓ Propagated version to %d targets", len(cfg.Propagate))
	}

	// 8. Post-bump hook.
	if err := RunHook(dir, cfg.Hooks.PostBump, newVer.String(), baseVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 9. Changelog.
	var changelogContent string
	var releaseBody string
	if includesReleaseArtifacts(cfg) && cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
		refs := resolveReferences(dir, parsed)
		entry := changelog.Generate(newVer.String(), parsed, refs)
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
	if err := git.CreateCommit(dir, commitMsg, releaseFiles(dir, detectDir, det, cfg)...); err != nil {
		return nil, fmt.Errorf("creating release commit: %w", err)
	}
	report("✓ Created release commit")

	// 11. Tag.
	tag := git.NamespacedTagString(tagPrefix, newVer)
	if err := git.CreateTag(dir, tag, fmt.Sprintf("Release %s", releaseCommitLabel(packageName, newVer.String()))); err != nil {
		return nil, err
	}
	report("✓ Tagged %s", tag)

	// 11b. Push commit and tag to remote.
	if err := git.Push(dir, tag); err != nil {
		return nil, fmt.Errorf("pushing release: %w", err)
	}
	report("✓ Pushed commit and tag to remote")

	// 12. Pre-publish hook.
	if err := RunHook(dir, cfg.Hooks.PrePublish, newVer.String(), baseVer.CoreString(), cfg.Project, hookOpts...); err != nil {
		return nil, err
	}

	// 13. Publish.
	if err := runGitHubPublish(dir, cfg, tag, newVer.String(), releaseBody); err != nil {
		return nil, err
	}

	// 14. Post-publish hook.
	if err := RunHook(dir, cfg.Hooks.PostPublish, newVer.String(), baseVer.CoreString(), cfg.Project, hookOpts...); err != nil {
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
		if err := git.CreateCommit(dir, snapMsg, versionFiles(dir, detectDir, det)...); err != nil {
			return nil, fmt.Errorf("creating snapshot commit: %w", err)
		}
		report("✓ Bumped to next development version: %s", snapVer.String())
		if err := git.Push(dir); err != nil {
			return nil, fmt.Errorf("pushing snapshot commit: %w", err)
		}
		report("✓ Pushed snapshot commit to remote")
	}

	return &Result{
		PrevVersion: baseVer.CoreString(),
		NewVersion:  newVer.String(),
		TagName:     tag,
	}, nil
}

func resolveReleaseBase(dir string, manifestVer version.Semver, tagPrefix string) (version.Semver, string, error) {
	if manifestVer.IsPreRelease() {
		baseVer, err := git.LatestSemverTag(dir, tagPrefix)
		if err != nil {
			return version.Semver{}, "", err
		}
		if baseVer.IsZero() {
			return baseVer, "", nil
		}
		return baseVer, git.NamespacedTagString(tagPrefix, baseVer), nil
	}

	baseVer := manifestVer.StripPreRelease()
	if baseVer.IsZero() {
		return baseVer, "", nil
	}
	return baseVer, git.NamespacedTagString(tagPrefix, baseVer), nil
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
	if includesReleaseArtifacts(cfg) && cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
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

func includesReleaseArtifacts(cfg *config.Config) bool {
	return cfg.Commit.Mode != "version-only"
}

func releaseFiles(dir, detectDir string, det detector.Detector, cfg *config.Config) []string {
	files := versionFiles(dir, detectDir, det)
	for _, target := range cfg.Propagate {
		files = append(files, releasePath(dir, detectDir, target.File))
	}
	if includesReleaseArtifacts(cfg) && cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
		files = append(files, releasePath(dir, detectDir, cfg.Changelog.File))
	}
	return uniqueFiles(files)
}

func versionFiles(dir, detectDir string, det detector.Detector) []string {
	files := make([]string, 0, len(det.VersionFiles()))
	for _, file := range det.VersionFiles() {
		files = append(files, releasePath(dir, detectDir, file))
	}
	return files
}

func releasePath(dir, detectDir, file string) string {
	prefix, err := filepath.Rel(dir, detectDir)
	if err != nil || prefix == "." {
		return filepath.Clean(file)
	}
	return filepath.Join(prefix, file)
}

func uniqueFiles(files []string) []string {
	seen := make(map[string]bool, len(files))
	out := make([]string, 0, len(files))
	for _, file := range files {
		if file != "" && !seen[file] {
			seen[file] = true
			out = append(out, file)
		}
	}
	return out
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

// resolveReferences resolves PR/issue references for each commit.
// Tries the GitHub API first (authoritative PR associations + linked issues).
// Falls back to regex extraction from commit text when the API is unavailable.
func resolveReferences(dir string, parsed []commits.ParsedCommit) map[string][]string {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		owner, repo, err := git.RemoteOwnerRepo(dir)
		if err != nil {
			report("⚠ Cannot determine repository owner/name, falling back to regex references: %v", err)
		} else {
			client := &gh.Client{Token: token, Owner: owner, Repo: repo}
			shas := make([]string, len(parsed))
			for i, p := range parsed {
				shas[i] = p.Hash
			}
			return client.ResolveCommitPRs(shas)
		}
	}

	// Fallback: regex extraction from commit messages.
	refs := make(map[string][]string)
	for _, p := range parsed {
		extracted := changelog.ExtractReferences(p.Subject, p.Body)
		if len(extracted) > 0 {
			refs[p.Hash] = extracted
		}
	}
	return refs
}

func report(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
