package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x1306e6d/release-cli/internal/config"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "test@test.com")
	mustGit(t, dir, "config", "user.name", "Test")
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
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Add feature commit.
	os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new feature"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: add new feature")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Commits: config.CommitsConfig{Convention: "conventional"},
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

func TestPipeline_NoReleasableChanges(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Non-conventional commit.
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("update"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "update docs")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Commits: config.CommitsConfig{Convention: "conventional"},
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

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	os.WriteFile(filepath.Join(dir, "feat.js"), []byte("//"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feat: new thing")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Commits: config.CommitsConfig{Convention: "conventional"},
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

func TestPipeline_FreeformConvention(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	mustGit(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Plain English commits (not conventional).
	os.WriteFile(filepath.Join(dir, "feature.js"), []byte("// new"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "Add user export feature")

	os.WriteFile(filepath.Join(dir, "fix.js"), []byte("// fix"), 0644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "Fix login bug")

	cfg := &config.Config{
		Project: "node",
		Version: config.VersionConfig{Scheme: "semver"},
		Commits: config.CommitsConfig{Convention: "freeform"},
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
		t.Fatal("expected result, got nil — freeform should treat all commits as releasable")
	}

	if result.NewVersion != "1.0.1" {
		t.Errorf("new version = %q, want %q (freeform = patch bump)", result.NewVersion, "1.0.1")
	}
	if result.TagName != "v1.0.1" {
		t.Errorf("tag = %q, want %q", result.TagName, "v1.0.1")
	}
}

func boolPtr(b bool) *bool {
	return &b
}
