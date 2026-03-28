package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/0x1306e6d/release-cli/internal/changelog"
	"github.com/0x1306e6d/release-cli/internal/commits"
	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/detector"
	"github.com/0x1306e6d/release-cli/internal/git"
	"github.com/0x1306e6d/release-cli/internal/propagate"
	"github.com/0x1306e6d/release-cli/internal/version"
)

// PackageResult holds per-package pipeline results for batched releases.
type PackageResult struct {
	Package     *PackageContext
	Config      *config.Config
	Detector    detector.Detector
	DetectDir   string
	PrevVersion version.Semver
	NewVersion  version.Semver
	TagName     string
	Parsed      []commits.ParsedCommit
	ReleaseBody string
}

// BatchRelease coordinates a batched release of multiple packages.
// Each package runs detect/analyze/bump/propagate/changelog independently,
// then one commit and multiple tags are created.
func BatchRelease(dir string, packages []*PackageContext, configs []*config.Config, dryRun bool, bumpOverride *version.BumpType) ([]*Result, error) {
	if len(packages) != len(configs) {
		return nil, fmt.Errorf("packages and configs must have the same length")
	}

	var results []*PackageResult
	registry := detector.DefaultRegistry()

	// Phase 1: Per-package detect/analyze/bump/propagate/changelog.
	for i, pkg := range packages {
		cfg := configs[i]

		detectDir := dir
		if pkg.Path != "" {
			detectDir = filepath.Join(dir, pkg.Path)
		}

		det, err := registry.Resolve(cfg.Project, detectDir)
		if err != nil {
			return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
		}
		report("[%s] Detected project type: %s", pkg.Name, det.Name())

		prevVer, err := readCurrentVersion(dir, det, detectDir, pkg.TagPrefix)
		if err != nil {
			return nil, fmt.Errorf("package %q: reading current version: %w", pkg.Name, err)
		}
		report("[%s] Current version: %s", pkg.Name, prevVer.String())

		fromTag := ""
		if !prevVer.IsZero() {
			fromTag = git.NamespacedTagString(pkg.TagPrefix, prevVer.StripPreRelease())
		}

		gitCommits, err := git.LogBetween(dir, fromTag, "HEAD", pkg.Path)
		if err != nil {
			return nil, fmt.Errorf("package %q: reading commit log: %w", pkg.Name, err)
		}

		rawCommits := make([]commits.RawCommit, len(gitCommits))
		for j, c := range gitCommits {
			rawCommits[j] = commits.RawCommit{Hash: c.Hash, Subject: c.Subject, Body: c.Body}
		}

		conv := resolveConvention(cfg)
		parsed, bt := commits.Analyze(rawCommits, conv)

		if bumpOverride != nil {
			bt = bumpOverride
		}

		if bt == nil {
			if pkg.IsForced {
				patch := version.BumpPatch
				bt = &patch
				report("[%s] No releasable changes, forced patch bump", pkg.Name)
			} else {
				report("[%s] No releasable changes found, skipping", pkg.Name)
				continue
			}
		}

		// Calculate new version.
		baseVer := prevVer.StripPreRelease()
		newVer := baseVer.Bump(*bt)
		report("[%s] Version bump: %s → %s", pkg.Name, prevVer.CoreString(), newVer.String())

		if dryRun {
			tag := git.NamespacedTagString(pkg.TagPrefix, newVer)
			report("[%s] [dry-run] Would bump: %s → %s, tag: %s", pkg.Name, prevVer.CoreString(), newVer.String(), tag)
			results = append(results, &PackageResult{
				Package:     pkg,
				Config:      cfg,
				PrevVersion: prevVer,
				NewVersion:  newVer,
				TagName:     tag,
			})
			continue
		}

		hookOpts := HookOptions{PackageName: pkg.Name, PackagePath: pkg.Path}

		// Pre-bump hook.
		if err := RunHook(dir, cfg.Hooks.PreBump, newVer.String(), prevVer.CoreString(), cfg.Project, hookOpts); err != nil {
			return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
		}

		if err := det.WriteVersion(detectDir, detector.Version{Raw: newVer.String()}); err != nil {
			return nil, fmt.Errorf("package %q: writing version: %w", pkg.Name, err)
		}
		report("[%s] ✓ Bumped version: %s → %s", pkg.Name, prevVer.CoreString(), newVer.String())

		// Propagate.
		if len(cfg.Propagate) > 0 {
			if err := propagate.Propagate(detectDir, cfg.Propagate, newVer.String()); err != nil {
				return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
			}
			report("[%s] ✓ Propagated version to %d targets", pkg.Name, len(cfg.Propagate))
		}

		// Post-bump hook.
		if err := RunHook(dir, cfg.Hooks.PostBump, newVer.String(), prevVer.CoreString(), cfg.Project, hookOpts); err != nil {
			return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
		}

		var releaseBody string
		if cfg.Changelog.Enabled != nil && *cfg.Changelog.Enabled {
			entry := changelog.Generate(newVer.String(), parsed)
			entry.Grouped = cfg.Changes.IsGroupedChangelog()
			var changelogContent string
			if cfg.Changelog.Template != "" {
				changelogContent, err = changelog.RenderCustom(entry, cfg.Changelog.Template)
				if err != nil {
					return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
				}
				releaseBody = changelogContent
			} else {
				changelogContent = entry.Render()
				releaseBody = entry.RenderBody()
			}
			if err := changelog.WriteFile(detectDir, cfg.Changelog.File, changelogContent); err != nil {
				return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
			}
			report("[%s] ✓ Updated %s", pkg.Name, cfg.Changelog.File)
		}

		tag := git.NamespacedTagString(pkg.TagPrefix, newVer)
		results = append(results, &PackageResult{
			Package:     pkg,
			Config:      cfg,
			Detector:    det,
			DetectDir:   detectDir,
			PrevVersion: prevVer,
			NewVersion:  newVer,
			TagName:     tag,
			Parsed:      parsed,
			ReleaseBody: releaseBody,
		})
	}

	if len(results) == 0 {
		report("No packages have releasable changes.")
		return nil, nil
	}

	if dryRun {
		var out []*Result
		for _, r := range results {
			out = append(out, &Result{
				PrevVersion: r.PrevVersion.CoreString(),
				NewVersion:  r.NewVersion.String(),
				TagName:     r.TagName,
			})
		}
		return out, nil
	}

	// Phase 2: Batched commit.
	var labels []string
	for _, r := range results {
		labels = append(labels, r.Package.Name+" "+r.NewVersion.String())
	}
	commitMsg := fmt.Sprintf("Release %s", strings.Join(labels, ", "))
	if err := git.CreateCommit(dir, commitMsg, "."); err != nil {
		return nil, fmt.Errorf("creating batched release commit: %w", err)
	}
	report("✓ Created release commit: %s", commitMsg)

	// Phase 2b: Create tags.
	for _, r := range results {
		if err := git.CreateTag(dir, r.TagName, fmt.Sprintf("Release %s %s", r.Package.Name, r.NewVersion.String())); err != nil {
			return nil, fmt.Errorf("creating tag %s: %w", r.TagName, err)
		}
		report("✓ Tagged %s", r.TagName)
	}

	// Phase 2c: Push commit and all tags.
	var tagNames []string
	for _, r := range results {
		tagNames = append(tagNames, r.TagName)
	}
	if err := git.Push(dir, tagNames...); err != nil {
		return nil, fmt.Errorf("pushing batched release: %w", err)
	}
	report("✓ Pushed commit and tags to remote")

	// Phase 3: Per-package publish and notify.
	var finalResults []*Result
	for _, r := range results {
		hookOpts := HookOptions{PackageName: r.Package.Name, PackagePath: r.Package.Path}

		// Pre-publish hook.
		if err := RunHook(dir, r.Config.Hooks.PrePublish, r.NewVersion.String(), r.PrevVersion.CoreString(), r.Config.Project, hookOpts); err != nil {
			return nil, fmt.Errorf("package %q: %w", r.Package.Name, err)
		}

		// Publish.
		if err := runGitHubPublish(dir, r.Config, r.TagName, r.NewVersion.String(), r.ReleaseBody); err != nil {
			return nil, fmt.Errorf("package %q: %w", r.Package.Name, err)
		}

		// Post-publish hook.
		if err := RunHook(dir, r.Config.Hooks.PostPublish, r.NewVersion.String(), r.PrevVersion.CoreString(), r.Config.Project, hookOpts); err != nil {
			return nil, fmt.Errorf("package %q: %w", r.Package.Name, err)
		}

		finalResults = append(finalResults, &Result{
			PrevVersion: r.PrevVersion.CoreString(),
			NewVersion:  r.NewVersion.String(),
			TagName:     r.TagName,
		})
	}

	// Phase 4: Batched SNAPSHOT post-release.
	var snapResults []*PackageResult
	for _, r := range results {
		if r.Config.Version.Snapshot && r.Detector != nil && r.Detector.SnapshotSuffix() != "" {
			snapResults = append(snapResults, r)
		}
	}
	if len(snapResults) > 0 {
		for _, r := range snapResults {
			snapVer := version.NextSnapshot(r.NewVersion, version.NormalizeSnapshotSuffix(r.Detector.SnapshotSuffix()))
			if err := r.Detector.WriteVersion(r.DetectDir, detector.Version{Raw: snapVer.String()}); err != nil {
				return nil, fmt.Errorf("package %q: writing snapshot version: %w", r.Package.Name, err)
			}
			report("[%s] ✓ Bumped to next development version: %s", r.Package.Name, snapVer.String())
		}
		if err := git.CreateCommit(dir, "Prepare next development iteration", "."); err != nil {
			return nil, fmt.Errorf("creating snapshot commit: %w", err)
		}
		if err := git.Push(dir); err != nil {
			return nil, fmt.Errorf("pushing snapshot commit: %w", err)
		}
		report("✓ Pushed snapshot commit to remote")
	}

	return finalResults, nil
}
