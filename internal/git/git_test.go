package git

import (
	"os"
	"path/filepath"
	"testing"
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
