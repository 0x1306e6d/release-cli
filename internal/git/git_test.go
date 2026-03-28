package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0x1306e6d/release-cli/internal/version"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustRun(t, dir, "init")
	mustRun(t, dir, "config", "user.email", "test@test.com")
	mustRun(t, dir, "config", "user.name", "Test")
	// Create initial commit.
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "initial commit")
	return dir
}

func mustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := run(dir, args...); err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
}

func TestLatestSemverTag_NoTags(t *testing.T) {
	dir := initTestRepo(t)
	v, err := LatestSemverTag(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "0.0.0" {
		t.Errorf("got %q, want %q", v.String(), "0.0.0")
	}
}

func TestLatestSemverTag_WithTags(t *testing.T) {
	dir := initTestRepo(t)
	mustRun(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("change"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "second commit")
	mustRun(t, dir, "tag", "-a", "v1.2.0", "-m", "v1.2.0")

	v, err := LatestSemverTag(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "1.2.0" {
		t.Errorf("got %q, want %q", v.String(), "1.2.0")
	}
}

func TestCreateTag(t *testing.T) {
	dir := initTestRepo(t)
	err := CreateTag(dir, "v2.0.0", "Release 2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	v, err := LatestSemverTag(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "2.0.0" {
		t.Errorf("got %q, want %q", v.String(), "2.0.0")
	}
}

func TestLogBetween(t *testing.T) {
	dir := initTestRepo(t)
	mustRun(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "feat: add feature A")

	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "fix: fix bug B")

	commits, err := LogBetween(dir, "v1.0.0", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2", len(commits))
	}
}

func TestLatestSemverTag_WithPrefix(t *testing.T) {
	dir := initTestRepo(t)
	mustRun(t, dir, "tag", "-a", "v1.0.0", "-m", "root v1.0.0")
	mustRun(t, dir, "tag", "-a", "cli/v0.5.0", "-m", "cli v0.5.0")
	mustRun(t, dir, "tag", "-a", "cli/v1.1.0", "-m", "cli v1.1.0")
	mustRun(t, dir, "tag", "-a", "lib/v2.0.0", "-m", "lib v2.0.0")

	// Without prefix: should return v1.0.0 (unprefixed tags only).
	v, err := LatestSemverTag(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "1.0.0" {
		t.Errorf("no prefix: got %q, want %q", v.String(), "1.0.0")
	}

	// With "cli" prefix: should return 1.1.0.
	v, err = LatestSemverTag(dir, "cli")
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "1.1.0" {
		t.Errorf("cli prefix: got %q, want %q", v.String(), "1.1.0")
	}

	// With "lib" prefix: should return 2.0.0.
	v, err = LatestSemverTag(dir, "lib")
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "2.0.0" {
		t.Errorf("lib prefix: got %q, want %q", v.String(), "2.0.0")
	}

	// With unknown prefix: should return 0.0.0.
	v, err = LatestSemverTag(dir, "unknown")
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "0.0.0" {
		t.Errorf("unknown prefix: got %q, want %q", v.String(), "0.0.0")
	}
}

func TestNamespacedTagString(t *testing.T) {
	v, _ := version.Parse("v1.2.3")

	if got := NamespacedTagString("", v); got != "v1.2.3" {
		t.Errorf("empty prefix: got %q, want %q", got, "v1.2.3")
	}
	if got := NamespacedTagString("cli", v); got != "cli/v1.2.3" {
		t.Errorf("cli prefix: got %q, want %q", got, "cli/v1.2.3")
	}
	if got := NamespacedTagString("workflow/sub", v); got != "workflow/sub/v1.2.3" {
		t.Errorf("nested prefix: got %q, want %q", got, "workflow/sub/v1.2.3")
	}
}

func TestLogBetween_WithPathFilter(t *testing.T) {
	dir := initTestRepo(t)
	mustRun(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")

	// Create file in cli/ subdirectory.
	os.MkdirAll(filepath.Join(dir, "cli"), 0755)
	os.WriteFile(filepath.Join(dir, "cli", "main.go"), []byte("package cli"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "feat: add cli module")

	// Create file in lib/ subdirectory.
	os.MkdirAll(filepath.Join(dir, "lib"), 0755)
	os.WriteFile(filepath.Join(dir, "lib", "lib.go"), []byte("package lib"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "feat: add lib module")

	// Create file at root.
	os.WriteFile(filepath.Join(dir, "root.go"), []byte("package main"), 0644)
	mustRun(t, dir, "add", ".")
	mustRun(t, dir, "commit", "-m", "fix: fix root")

	// Without filter: all 3 commits.
	all, err := LogBetween(dir, "v1.0.0", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Errorf("no filter: got %d commits, want 3", len(all))
	}

	// Filter to cli/: should get 1 commit.
	cliCommits, err := LogBetween(dir, "v1.0.0", "HEAD", "cli")
	if err != nil {
		t.Fatal(err)
	}
	if len(cliCommits) != 1 {
		t.Errorf("cli filter: got %d commits, want 1", len(cliCommits))
	}

	// Filter to lib/: should get 1 commit.
	libCommits, err := LogBetween(dir, "v1.0.0", "HEAD", "lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(libCommits) != 1 {
		t.Errorf("lib filter: got %d commits, want 1", len(libCommits))
	}
}

func TestCreateCommit(t *testing.T) {
	dir := initTestRepo(t)
	path := filepath.Join(dir, "new.txt")
	os.WriteFile(path, []byte("new file"), 0644)

	err := CreateCommit(dir, "add new file", "new.txt")
	if err != nil {
		t.Fatal(err)
	}

	// Verify commit was made.
	out, err := run(dir, "log", "--oneline", "-1")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Error("expected commit log output")
	}
}
