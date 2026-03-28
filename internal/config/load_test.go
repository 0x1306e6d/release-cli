package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MinimalConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "project: go\n")

	cfg, warnings, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	if cfg.Project != "go" {
		t.Errorf("project = %q, want %q", cfg.Project, "go")
	}
	// Check defaults
	if cfg.Version.Scheme != "semver" {
		t.Errorf("version.scheme = %q, want %q", cfg.Version.Scheme, "semver")
	}
	if cfg.Changes.Commits != nil {
		t.Error("changes.commits should be nil when not configured")
	}
	if cfg.Changelog.Enabled == nil || !*cfg.Changelog.Enabled {
		t.Error("changelog.enabled should default to true")
	}
	if cfg.Changelog.File != "CHANGELOG.md" {
		t.Errorf("changelog.file = %q, want %q", cfg.Changelog.File, "CHANGELOG.md")
	}
	if cfg.Publish.GitHub.Enabled == nil || !*cfg.Publish.GitHub.Enabled {
		t.Error("publish.github.enabled should default to true")
	}
}

func TestLoad_FullConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `project: node
version:
  scheme: semver
  snapshot: true
changes:
  commits:
    convention: angular
changelog:
  enabled: false
  file: HISTORY.md
propagate:
  - file: Dockerfile
    type: docker-label
hooks:
  pre-bump: "npm test"
  post-bump: "npm run build"
publish:
  github:
    enabled: true
    draft: true
    artifacts:
      - dist/*.tar.gz
`)

	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project != "node" {
		t.Errorf("project = %q, want %q", cfg.Project, "node")
	}
	if !cfg.Version.Snapshot {
		t.Error("version.snapshot should be true")
	}
	if cfg.Changes.Commits == nil {
		t.Fatal("changes.commits should not be nil")
	}
	if cfg.Changes.Commits.Convention != "angular" {
		t.Errorf("changes.commits.convention = %q, want %q", cfg.Changes.Commits.Convention, "angular")
	}
	if cfg.Changelog.Enabled == nil || *cfg.Changelog.Enabled {
		t.Error("changelog.enabled should be false")
	}
	if cfg.Changelog.File != "HISTORY.md" {
		t.Errorf("changelog.file = %q, want %q", cfg.Changelog.File, "HISTORY.md")
	}
	if len(cfg.Propagate) != 1 {
		t.Fatalf("propagate count = %d, want 1", len(cfg.Propagate))
	}
	if cfg.Propagate[0].Type != "docker-label" {
		t.Errorf("propagate[0].type = %q, want %q", cfg.Propagate[0].Type, "docker-label")
	}
	if cfg.Hooks.PreBump != "npm test" {
		t.Errorf("hooks.pre-bump = %q, want %q", cfg.Hooks.PreBump, "npm test")
	}
	if !cfg.Publish.GitHub.Draft {
		t.Error("publish.github.draft should be true")
	}
	if len(cfg.Publish.GitHub.Artifacts) != 1 {
		t.Fatalf("artifacts count = %d, want 1", len(cfg.Publish.GitHub.Artifacts))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestLoad_MissingProject(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "version:\n  scheme: semver\n")

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error for missing project")
	}
}

func TestLoad_EnvVarResolution(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TEST_WEBHOOK_URL", "https://hooks.slack.com/test")
	writeConfig(t, dir, `project: go
notify:
  slack:
    webhook: ${TEST_WEBHOOK_URL}
`)

	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Notify.Slack == nil {
		t.Fatal("notify.slack should not be nil")
	}
	if cfg.Notify.Slack.Webhook != "https://hooks.slack.com/test" {
		t.Errorf("webhook = %q, want %q", cfg.Notify.Slack.Webhook, "https://hooks.slack.com/test")
	}
}

func TestLoad_UndefinedEnvVar(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `project: go
notify:
  slack:
    webhook: ${UNDEFINED_VAR_12345}
`)

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for undefined env var")
	}
}

func TestLoad_UnknownKeys(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "project: go\nunknown_key: value\n")

	_, warnings, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
}

func TestLoad_InvalidConvention(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "project: go\nchanges:\n  commits:\n    convention: invalid\n")

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error for invalid convention")
	}
}

func TestLoad_CustomConventionRequiresTypes(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "project: go\nchanges:\n  commits:\n    convention: custom\n")

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error for custom convention without types")
	}
}

func TestLoad_PropagateRequiresStrategy(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "project: go\npropagate:\n  - file: Dockerfile\n")

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error for propagate without type/field/pattern")
	}
}

func TestLoad_MonorepoWithNameAndModules(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `name: my-project
project: go
modules:
  - cli
  - workflow
`)

	cfg, warnings, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if cfg.Name != "my-project" {
		t.Errorf("name = %q, want %q", cfg.Name, "my-project")
	}
	if len(cfg.Modules) != 2 {
		t.Fatalf("modules count = %d, want 2", len(cfg.Modules))
	}
	if cfg.Modules[0] != "cli" || cfg.Modules[1] != "workflow" {
		t.Errorf("modules = %v, want [cli workflow]", cfg.Modules)
	}
	if !cfg.IsMonorepo() {
		t.Error("IsMonorepo() should return true when modules are declared")
	}
}

func TestLoad_ModulesWithoutName(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `project: go
modules:
  - cli
`)

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected validation error for modules without name")
	}
}

func TestLoad_SingleProjectIsNotMonorepo(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "project: go\n")

	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IsMonorepo() {
		t.Error("IsMonorepo() should return false for single-project config")
	}
}

func TestLoad_NameIgnoredWithoutModules(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "name: foo\nproject: go\n")

	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "foo" {
		t.Errorf("name = %q, want %q", cfg.Name, "foo")
	}
	if cfg.IsMonorepo() {
		t.Error("IsMonorepo() should return false when no modules declared")
	}
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}
