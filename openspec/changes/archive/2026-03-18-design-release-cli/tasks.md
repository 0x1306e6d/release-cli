## 1. Project Scaffolding

- [x] 1.1 Initialize Go module (`go mod init`), add cobra dependency, create `cmd/release-cli/main.go` entry point
- [x] 1.2 Set up project directory structure (`internal/cli`, `internal/config`, `internal/detector`, `internal/version`, `internal/commits`, `internal/changelog`, `internal/propagate`, `internal/pipeline`, `internal/publish`, `internal/notify`, `internal/git`)
- [x] 1.3 Implement root cobra command with `--version` flag and global flags (`--dry-run`, `--verbose`)

## 2. Configuration

- [x] 2.1 Define config struct matching `.release.yaml` schema (project, version, commits, changelog, propagate, hooks, publish, notify sections)
- [x] 2.2 Implement YAML config parser with validation (required `project` field, valid identifiers, valid enum values)
- [x] 2.3 Implement environment variable resolution (`${ENV_VAR}` syntax) in config string values
- [x] 2.4 Implement config defaults (semver, conventional commits, changelog enabled, GitHub publish enabled)
- [x] 2.5 Add warning for unknown config keys

## 3. Detector Framework

- [x] 3.1 Define `Detector` interface (`Name`, `Aliases`, `Detect`, `ReadVersion`, `WriteVersion`, `DefaultPublishTargets`, `SnapshotSuffix`)
- [x] 3.2 Implement detector registry with exact-name lookup and general-identifier resolution
- [x] 3.3 Handle ambiguous general identifier detection (multiple build tools found → error with suggestions)

## 4. Version Management

- [x] 4.1 Implement semver parsing and bumping (major, minor, patch) with pre-release suffix support
- [x] 4.2 Implement version reading from git tags (for tag-based ecosystems) with `v` prefix handling
- [x] 4.3 Implement SNAPSHOT/pre-release stripping on release and post-release SNAPSHOT bump with ecosystem-aware suffixes
- [x] 4.4 Implement manifest override support (`version.manifest` and `version.field`/`version.pattern` config)

## 5. Built-in Detectors (v1)

- [x] 5.1 Implement Go detector (version from git tags, no manifest write)
- [x] 5.2 Implement Node detector (read/write `version` in `package.json`)
- [x] 5.3 Implement Python detector (read/write `version` in `pyproject.toml` under `[project]`)
- [x] 5.4 Implement Rust detector (read/write `version` in `Cargo.toml` under `[package]`)
- [x] 5.5 Implement Java-Gradle detector (read/write `version` in `gradle.properties`)
- [x] 5.6 Implement Java-Maven detector (read/write `<version>` in `pom.xml`)
- [x] 5.7 Implement Dart detector (read/write `version` in `pubspec.yaml`)
- [x] 5.8 Implement Helm detector (read/write `version` in `Chart.yaml`)

## 6. Git Operations

- [x] 6.1 Implement git tag listing and latest semver tag resolution
- [x] 6.2 Implement git commit log retrieval between two refs (tag..HEAD)
- [x] 6.3 Implement git commit creation (staging files, creating commit with message)
- [x] 6.4 Implement git tag creation (`v<version>` format)

## 7. Commit Analysis

- [x] 7.1 Implement Conventional Commits parser (type, scope, breaking change detection from `!` suffix and `BREAKING CHANGE` footer)
- [x] 7.2 Implement bump type resolution (highest wins: major > minor > patch)
- [x] 7.3 Implement Angular commit convention parser
- [x] 7.4 Implement custom convention support (user-defined type-to-bump mappings from config)
- [x] 7.5 Handle "no releasable changes" case (exit 0 with message)

## 8. Version Propagation

- [x] 8.1 Implement structured field propagation for YAML, JSON, and TOML files (parse, update field, write back)
- [x] 8.2 Implement pattern-based propagation (`{{.Version}}` template matching and replacement in arbitrary files)
- [x] 8.3 Implement built-in propagation types (`docker-label` and others) that expand to predefined patterns
- [x] 8.4 Add error reporting when pattern is not found in target file

## 9. Changelog Generation

- [x] 9.1 Implement default changelog template (grouped by commit type: Breaking Changes, Features, Bug Fixes, Other)
- [x] 9.2 Implement changelog file writing (prepend new entry, preserve existing content, create file if missing)
- [x] 9.3 Implement custom Go template support for changelog rendering
- [x] 9.4 Support `changelog.enabled: false` to skip generation

## 10. Release Pipeline

- [x] 10.1 Implement pipeline orchestrator with fixed step order (detect → analyze → bump → propagate → changelog → commit → tag → publish → notify)
- [x] 10.2 Implement step skipping based on config (disabled changelog, no publish targets, no notify targets)
- [x] 10.3 Implement lifecycle hooks execution (pre-bump, post-bump, pre-publish, post-publish) with environment variables (`RELEASE_VERSION`, `RELEASE_PREV_VERSION`, `RELEASE_PROJECT`)
- [x] 10.4 Implement hook failure handling (abort pipeline, report error with command output)
- [x] 10.5 Implement SNAPSHOT post-release step (bump to next dev version, create "Prepare next development iteration" commit)
- [x] 10.6 Implement `--dry-run` mode (preview all steps without executing)
- [x] 10.7 Implement progress reporting (`✓ Bumped version: 1.3.0 → 1.4.0`, etc.)

## 11. Publish Integration

- [x] 11.1 Implement GitHub Release creation (create release from tag, set body from changelog entry)
- [x] 11.2 Implement GitHub Release artifact upload (glob pattern matching, file upload)
- [x] 11.3 Implement GitHub Release draft mode
- [x] 11.4 Implement publish target enable/disable (`enabled: false`)

## 12. Notify Integration

- [x] 12.1 Implement Slack webhook notification (POST JSON with version and changelog summary)
- [x] 12.2 Implement generic webhook notification (POST JSON payload with release metadata)
- [x] 12.3 Ensure notification failures are warnings, not errors (release already completed)

## 13. CLI Commands

- [x] 13.1 Implement `release-cli init` command (detect project type, auto-detect snapshot, generate `.release.yaml` with commented-out optional sections)
- [x] 13.2 Implement `release-cli release` command (load config, run pipeline)
- [x] 13.3 Implement `release-cli status` command (show current version, last release, commits since, next version preview)
- [x] 13.4 Prevent `init` from overwriting existing config

## 14. Testing

- [x] 14.1 Add unit tests for each detector (version read/write, detection logic)
- [x] 14.2 Add unit tests for semver parsing, bumping, and SNAPSHOT handling
- [x] 14.3 Add unit tests for commit parsing (conventional, angular, custom)
- [x] 14.4 Add unit tests for config parsing and validation
- [x] 14.5 Add integration tests for the full release pipeline using a temporary git repo
