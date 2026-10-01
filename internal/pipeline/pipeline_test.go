package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/version"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "test@test.com")
	mustGit(t, dir, "config", "user.name", "Test")

	// Create a bare remote so push works in tests.
	bare := t.TempDir()
	mustGit(t, bare, "init", "--bare")
	mustGit(t, dir, "remote", "add", "origin", bare)

	return dir
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestPipeline_FullRelease_Node(t *testing.T) {
	dir := initTestRepo(t)

	// Create a Node project.
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Add feature commit.
	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add new feature")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	result, err := Run(Options{Dir: dir, Config: cfg})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil (no releasable changes)")
	}

	if result.NewVersion != "1.1.0" {
		t.Errorf("new version = %q, want %q", result.NewVersion, "1.1.0")
	}
	if result.TagName != "v1.1.0" {
		t.Errorf("tag = %q, want %q", result.TagName, "v1.1.0")
	}

	// Verify package.json updated.
	data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	if !strings.Contains(string(data), `"1.1.0"`) {
		t.Errorf("package.json not updated: %s", data)
	}

	// Verify changelog created.
	changelog, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if !strings.Contains(string(changelog), "1.1.0") {
		t.Errorf("CHANGELOG.md not created: %s", changelog)
	}
}

func TestPipeline_CustomGitPolicy(t *testing.T) {
	dir := initTestRepo(t)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "release-1.0.0", "-m", "release")
	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add new feature")

	push := false
	cfg := &config.Config{
		Project:   "node",
		Version:   config.VersionConfig{Scheme: "semver"},
		Changes:   config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Git:       config.GitConfig{Branch: "^main$", TagFormat: "release-{{ .Version }}", Push: &push, CommitArgs: []string{"--no-verify"}},
		Publish:   config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	result, err := Run(Options{Dir: dir, Config: cfg})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result.TagName != "release-1.1.0" {
		t.Errorf("tag = %q, want release-1.1.0", result.TagName)
	}
}

func TestPipeline_VersionOnlyCommit(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	_ = os.WriteFile(filepath.Join(dir, "version.txt"), []byte("VERSION=1.0.0\n"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add new feature")

	cfg := &config.Config{
		Project:   "node",
		Version:   config.VersionConfig{Scheme: "semver"},
		Changes:   config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(true), File: "CHANGELOG.md"},
		Commit:    config.CommitConfig{Mode: "version-only"},
		Propagate: []config.PropagateTarget{{File: "version.txt", Pattern: "VERSION={{.Version}}"}},
		Publish:   config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	if _, err := Run(Options{Dir: dir, Config: cfg}); err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "CHANGELOG.md")); !os.IsNotExist(err) {
		t.Fatalf("version-only release wrote changelog: %v", err)
	}

	cmd := exec.Command("git", "show", "--format=", "--name-only", "HEAD")
	cmd.Dir = dir
	files, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(string(files)); strings.Join(got, ",") != "package.json,version.txt" {
		t.Errorf("release commit files = %q, want package.json and version.txt", files)
	}
}

func TestPipeline_IncludesConfiguredReleaseFile(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add new feature")

	cfg := &config.Config{
		Project:   "node",
		Version:   config.VersionConfig{Scheme: "semver"},
		Changes:   config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Commit:    config.CommitConfig{Include: []string{"release-metadata.json"}},
		Hooks:     config.HooksConfig{PostBump: "echo release > release-metadata.json"},
		Publish:   config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	if _, err := Run(Options{Dir: dir, Config: cfg}); err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	cmd := exec.Command("git", "show", "--format=", "--name-only", "HEAD")
	cmd.Dir = dir
	files, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(string(files)); strings.Join(got, ",") != "package.json,release-metadata.json" {
		t.Errorf("release commit files = %q, want package.json and release-metadata.json", files)
	}
}

func TestPipeline_NoReleasableChanges(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Non-conventional commit.
	_ = os.WriteFile(filepath.Join(dir, "readme.md"), []byte("update"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "update docs")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
	}

	result, err := Run(Options{Dir: dir, Config: cfg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for no releasable changes")
	}
}

func TestPipeline_DryRun(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	_ = os.WriteFile(filepath.Join(dir, "feat.js"), []byte("//"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: new thing")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
	}

	result, err := Run(Options{Dir: dir, Config: cfg, DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result in dry-run")
	}

	// Verify nothing changed.
	data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	if strings.Contains(string(data), "1.1.0") {
		t.Error("dry-run should not modify package.json")
	}
}

func TestPipeline_PreReleaseManifestUsesLatestExistingTag(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.2.0-dev"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.1.0", "-m", "v1.1.0")

	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add feature")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(false),
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	result, err := Run(Options{Dir: dir, Config: cfg, DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result in dry-run")
	}
	if result.PrevVersion != "1.1.0" {
		t.Errorf("previous version = %q, want %q", result.PrevVersion, "1.1.0")
	}
	if result.NewVersion != "1.2.0" {
		t.Errorf("new version = %q, want %q", result.NewVersion, "1.2.0")
	}
	if result.TagName != "v1.2.0" {
		t.Errorf("tag = %q, want %q", result.TagName, "v1.2.0")
	}
}

func TestPipeline_FreeformConvention(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Plain English commits (not conventional).
	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "Add user export feature")

	_ = os.WriteFile(filepath.Join(dir, "fix.js"), []byte("// fix"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "Fix login bug")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		// No Changes configured — defaults to accept-all, flat changelog
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	result, err := Run(Options{Dir: dir, Config: cfg})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil — no convention should treat all commits as releasable")
	}

	if result.NewVersion != "1.0.1" {
		t.Errorf("new version = %q, want %q (no convention = patch bump)", result.NewVersion, "1.0.1")
	}
	if result.TagName != "v1.0.1" {
		t.Errorf("tag = %q, want %q", result.TagName, "v1.0.1")
	}

	// Verify flat changelog (no ### headings) when no convention is configured.
	changelog, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if strings.Contains(string(changelog), "###") {
		t.Errorf("no-convention changelog should not have ### headings:\n%s", changelog)
	}
}

func TestPipeline_BumpOverride(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Add a fix commit (normally patch).
	_ = os.WriteFile(filepath.Join(dir, "fix.js"), []byte("// fix"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "fix: resolve null pointer")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	bumpMinor := version.BumpMinor
	result, err := Run(Options{Dir: dir, Config: cfg, BumpOverride: &bumpMinor})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewVersion != "1.1.0" {
		t.Errorf("new version = %q, want %q (override minor)", result.NewVersion, "1.1.0")
	}
}

func TestPipeline_ExplicitReleaseVersion(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test","version":"1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	cfg := &config.Config{
		Project:   "node",
		Version:   config.VersionConfig{Scheme: "semver"},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Publish:   config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	result, err := Run(Options{Dir: dir, Config: cfg, ReleaseVersion: "2.0.0"})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil || result.NewVersion != "2.0.0" || result.TagName != "v2.0.0" {
		t.Fatalf("result = %#v, want version 2.0.0 and tag v2.0.0", result)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	if !strings.Contains(string(data), `"2.0.0"`) {
		t.Errorf("package.json = %s, want 2.0.0", data)
	}
}

func TestPipeline_InvalidExplicitVersionDoesNotWriteFiles(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test","version":"1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	cfg := &config.Config{Project: "node", Version: config.VersionConfig{Scheme: "semver"}}
	if _, err := Run(Options{Dir: dir, Config: cfg, ReleaseVersion: "1.0.0"}); err == nil {
		t.Fatal("expected non-incrementing release version error")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	if !strings.Contains(string(data), `"1.0.0"`) {
		t.Errorf("package.json was modified: %s", data)
	}
}

func TestPipeline_ExplicitNextSnapshotVersion(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "build.gradle"), []byte("plugins { id 'java' }\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "gradle.properties"), []byte("version=1.0.0-SNAPSHOT\n"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v0.9.0", "-m", "v0.9.0")

	cfg := &config.Config{
		Project:   "java-gradle",
		Version:   config.VersionConfig{Scheme: "semver", Snapshot: true},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Publish:   config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	if _, err := Run(Options{Dir: dir, Config: cfg, ReleaseVersion: "1.0.0", NextVersion: "1.1.0-SNAPSHOT"}); err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "gradle.properties"))
	if got := string(data); got != "version=1.1.0-SNAPSHOT\n" {
		t.Errorf("gradle.properties = %q, want next snapshot version", got)
	}
}

func TestResolveNextVersion(t *testing.T) {
	release := version.Semver{Major: 1, Minor: 0, Patch: 0}
	if _, err := resolveNextVersion(release, "1.0.0-SNAPSHOT", true, "SNAPSHOT"); err == nil {
		t.Fatal("expected non-incrementing snapshot error")
	}
	if _, err := resolveNextVersion(release, "1.1.0-dev", true, "SNAPSHOT"); err == nil {
		t.Fatal("expected invalid snapshot suffix error")
	}
}

func TestPipeline_RejectsOrphanNextVersion(t *testing.T) {
	if _, err := Run(Options{NextVersion: "1.1.0-SNAPSHOT"}); err == nil {
		t.Fatal("expected next version without release version error")
	}
}

func TestPipeline_NoCategorize_FlatChangelog(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Plain commits — no convention configured.
	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "Add user export feature")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		// No Changes configured — defaults to accept-all, flat changelog
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	result, err := Run(Options{Dir: dir, Config: cfg})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewVersion != "1.0.1" {
		t.Errorf("new version = %q, want %q (no convention = patch bump)", result.NewVersion, "1.0.1")
	}

	// Verify flat changelog (no ### headings).
	changelog, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if strings.Contains(string(changelog), "###") {
		t.Errorf("expected flat changelog without ### headings:\n%s", changelog)
	}
	if !strings.Contains(string(changelog), "- Add user export feature") {
		t.Errorf("expected commit in changelog:\n%s", changelog)
	}
}

func TestPipeline_MonorepoSinglePackage(t *testing.T) {
	dir := initTestRepo(t)

	// Create root with a "cli" module.
	_ = os.MkdirAll(filepath.Join(dir, "cli"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "cli", "package.json"), []byte(`{"name": "cli", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "cli/v1.0.0", "-m", "cli v1.0.0")

	// Add a feature commit touching cli/.
	_ = os.WriteFile(filepath.Join(dir, "cli", "feature.js"), []byte("// new"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add cli feature")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
		Commit: config.CommitConfig{Release: "Release {{ .Package }} {{ .ReleaseVersion }}"},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	result, err := Run(Options{
		Dir:    dir,
		Config: cfg,
		Package: &PackageContext{
			Name:      "cli",
			Path:      "cli",
			TagPrefix: "cli",
		},
	})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewVersion != "1.1.0" {
		t.Errorf("new version = %q, want %q", result.NewVersion, "1.1.0")
	}
	if result.TagName != "cli/v1.1.0" {
		t.Errorf("tag = %q, want %q", result.TagName, "cli/v1.1.0")
	}

	// Verify changelog written under cli/.
	changelog, _ := os.ReadFile(filepath.Join(dir, "cli", "CHANGELOG.md"))
	if !strings.Contains(string(changelog), "1.1.0") {
		t.Errorf("cli/CHANGELOG.md not created: %s", changelog)
	}
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = dir
	message, err := cmd.Output()
	if err != nil {
		t.Fatalf("reading release commit message: %v", err)
	}
	if got := strings.TrimSpace(string(message)); got != "Release cli 1.1.0" {
		t.Errorf("commit message = %q, want %q", got, "Release cli 1.1.0")
	}
}

func TestPipeline_PackageSnapshotUsesLatestExistingPackageTag(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.MkdirAll(filepath.Join(dir, "package"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "package", "build.gradle"), []byte("plugins { id 'java' }\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "package", "gradle.properties"), []byte("version=1.2.1-SNAPSHOT\ngroup=com.example\n"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "package/v1.2.0", "-m", "package v1.2.0")

	_ = os.MkdirAll(filepath.Join(dir, "package", "src"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "package", "src", "Main.java"), []byte("class Main {}\n"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "fix: update package")

	cfg := &config.Config{
		Project: "java-gradle",
		Version: config.VersionConfig{Scheme: "semver", Snapshot: true},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(false),
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	result, err := Run(Options{
		Dir:    dir,
		Config: cfg,
		DryRun: true,
		Package: &PackageContext{
			Name:      "package",
			Path:      "package",
			TagPrefix: "package",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result in dry-run")
	}
	if result.PrevVersion != "1.2.0" {
		t.Errorf("previous version = %q, want %q", result.PrevVersion, "1.2.0")
	}
	if result.NewVersion != "1.2.1" {
		t.Errorf("new version = %q, want %q", result.NewVersion, "1.2.1")
	}
	if result.TagName != "package/v1.2.1" {
		t.Errorf("tag = %q, want %q", result.TagName, "package/v1.2.1")
	}
}

func TestBatchRelease_PreReleaseManifestUsesLatestPackageTag(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.MkdirAll(filepath.Join(dir, "cli"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "cli", "package.json"), []byte(`{"name": "cli", "version": "1.2.0-dev"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "cli/v1.1.0", "-m", "cli v1.1.0")

	_ = os.WriteFile(filepath.Join(dir, "cli", "feature.js"), []byte("// feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add cli feature")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(false),
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	results, err := BatchRelease(
		dir,
		[]*PackageContext{{Name: "cli", Path: "cli", TagPrefix: "cli"}},
		[]*config.Config{cfg},
		true,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}
	if results[0].PrevVersion != "1.1.0" {
		t.Errorf("previous version = %q, want %q", results[0].PrevVersion, "1.1.0")
	}
	if results[0].NewVersion != "1.2.0" {
		t.Errorf("new version = %q, want %q", results[0].NewVersion, "1.2.0")
	}
	if results[0].TagName != "cli/v1.2.0" {
		t.Errorf("tag = %q, want %q", results[0].TagName, "cli/v1.2.0")
	}
}

func TestPipeline_MonorepoForcedRelease(t *testing.T) {
	dir := initTestRepo(t)

	// Create root with a "lib" module.
	_ = os.MkdirAll(filepath.Join(dir, "lib"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "lib", "package.json"), []byte(`{"name": "lib", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "lib/v1.0.0", "-m", "lib v1.0.0")

	// Add a commit that does NOT touch lib/ — only root.
	_ = os.WriteFile(filepath.Join(dir, "root.txt"), []byte("root change"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: root change")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Changes: config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{
			Enabled: boolPtr(true),
			File:    "CHANGELOG.md",
		},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	// Without forced: no releasable changes for lib.
	result, err := Run(Options{
		Dir:    dir,
		Config: cfg,
		Package: &PackageContext{
			Name:      "lib",
			Path:      "lib",
			TagPrefix: "lib",
			IsForced:  false,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result when no commits touch lib/")
	}

	// With forced: should get a patch bump.
	result, err = Run(Options{
		Dir:    dir,
		Config: cfg,
		Package: &PackageContext{
			Name:      "lib",
			Path:      "lib",
			TagPrefix: "lib",
			IsForced:  true,
		},
	})
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result with forced release")
	}
	if result.NewVersion != "1.0.1" {
		t.Errorf("new version = %q, want %q (forced patch)", result.NewVersion, "1.0.1")
	}
	if result.TagName != "lib/v1.0.1" {
		t.Errorf("tag = %q, want %q", result.TagName, "lib/v1.0.1")
	}
}

func TestPipeline_PathFilterExcludesOtherPackages(t *testing.T) {
	dir := initTestRepo(t)

	// Create two modules: cli and lib.
	_ = os.MkdirAll(filepath.Join(dir, "cli"), 0755)
	_ = os.MkdirAll(filepath.Join(dir, "lib"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "cli", "package.json"), []byte(`{"name": "cli", "version": "1.0.0"}`), 0644)
	_ = os.WriteFile(filepath.Join(dir, "lib", "package.json"), []byte(`{"name": "lib", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "cli/v1.0.0", "-m", "cli v1.0.0")
	mustGit(t, dir, "tag", "-a", "lib/v1.0.0", "-m", "lib v1.0.0")

	// Add commit only in lib/.
	_ = os.WriteFile(filepath.Join(dir, "lib", "feature.js"), []byte("// lib feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: lib feature")

	cfg := &config.Config{
		Project:   "node",
		Version:   config.VersionConfig{Scheme: "semver"},
		Changes:   config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Publish: config.PublishConfig{
			GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)},
		},
	}

	// cli should have no releasable changes.
	result, err := Run(Options{
		Dir:    dir,
		Config: cfg,
		Package: &PackageContext{
			Name:      "cli",
			Path:      "cli",
			TagPrefix: "cli",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for cli — commit was in lib/, not cli/")
	}
}

func TestPipeline_CommitMessageTemplates(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "build.gradle"), []byte("plugins { id 'java' }\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "gradle.properties"), []byte("version=1.0.0-SNAPSHOT\n"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v0.9.0", "-m", "v0.9.0")

	_ = os.WriteFile(filepath.Join(dir, "feature.java"), []byte("class Feature {}\n"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add feature")

	cfg := &config.Config{
		Project:   "java-gradle",
		Version:   config.VersionConfig{Scheme: "semver", Snapshot: true},
		Changes:   config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Commit: config.CommitConfig{
			Release: "Release {{ .Project }} {{ .ReleaseVersion }} -> {{ .NextVersion }}",
			Next:    "Next {{ .ReleaseVersion }} -> {{ .NextVersion }}",
		},
		Publish: config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	if _, err := Run(Options{Dir: dir, Config: cfg}); err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	cmd := exec.Command("git", "log", "--format=%s", "-2")
	cmd.Dir = dir
	log, err := cmd.Output()
	if err != nil {
		t.Fatalf("reading commit messages: %v", err)
	}
	if got, want := string(log), "Next 0.10.0 -> 0.11.0-SNAPSHOT\nRelease java-gradle 0.10.0 -> 0.11.0-SNAPSHOT\n"; got != want {
		t.Errorf("commit messages = %q, want %q", got, want)
	}
}

func TestPipeline_InvalidCommitMessageTemplateDoesNotCommit(t *testing.T) {
	dir := initTestRepo(t)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test","version":"1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")
	_ = os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add feature")

	cfg := &config.Config{
		Project:   "node",
		Version:   config.VersionConfig{Scheme: "semver"},
		Changes:   config.ChangesConfig{Commits: &config.CommitsConfig{Convention: "conventional"}},
		Changelog: config.ChangelogConfig{Enabled: boolPtr(false)},
		Commit:    config.CommitConfig{Release: "{{ .Missing }}"},
		Publish:   config.PublishConfig{GitHub: config.GitHubPublishConfig{Enabled: boolPtr(false)}},
	}

	if _, err := Run(Options{Dir: dir, Config: cfg}); err == nil {
		t.Fatal("expected invalid template error")
	}

	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = dir
	count, err := cmd.Output()
	if err != nil {
		t.Fatalf("counting commits: %v", err)
	}
	if got := strings.TrimSpace(string(count)); got != "2" {
		t.Errorf("commit count = %s, want 2", got)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
