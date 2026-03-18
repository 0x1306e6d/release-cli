## Why

Release workflows are inconsistent across projects and ecosystems. Teams forget steps, skip changelogs, tag wrong commits, or do releases differently in every repo. release-cli solves this by providing a single, opinionated CLI tool that detects your project type and executes a consistent release pipeline — written in Go, language agnostic, open source.

## What Changes

- New CLI tool `release-cli` built in Go, distributed as a single binary
- Auto-detection of project type (language + build tool) from manifest files
- Configuration-driven release workflow via `.release.yaml` with strong defaults
- Manifest-based versioning (source of truth in project files, not git tags) with git-tag fallback for ecosystems without version files (e.g., Go)
- Git tags always created on every release regardless of versioning strategy
- SNAPSHOT/pre-release version support with ecosystem-aware suffixes (e.g., `-SNAPSHOT` for Java, `.dev0` for Python)
- Configurable commit convention parsing (Conventional Commits default) to determine bump type
- Version propagation to secondary files (Dockerfile labels, version constants, etc.)
- Changelog generation from commit history
- Lifecycle hooks (pre-bump, post-bump, pre-publish, post-publish) for custom commands
- Publish integrations (GitHub Releases as default, ecosystem registries)
- Notification integrations (Slack, webhooks)

## Capabilities

### New Capabilities
- `project-detection`: Detect project type from manifest files. Support general (`java`) and specific (`java-gradle`) identifiers. Each detector knows how to read/write versions for its ecosystem.
- `version-management`: Read current version from manifest or git tags, calculate next version from commit analysis, bump manifest, create git tags. Support semver and calver schemes. Handle SNAPSHOT/pre-release lifecycle.
- `version-propagation`: Update secondary files (Dockerfile, version constants, config files) when version bumps. Support structured (YAML/JSON/TOML field paths) and pattern-based (regex/template) propagation targets.
- `commit-analysis`: Parse commit history since last release to determine bump type. Conventional Commits as default, configurable convention with custom major/minor/patch mappings.
- `changelog-generation`: Generate changelog entries from parsed commits grouped by type. Write to CHANGELOG.md with customizable templates.
- `release-pipeline`: Orchestrate the release flow: detect → analyze → bump → propagate → changelog → commit → tag → publish → notify. Lifecycle hooks at each stage.
- `release-config`: Parse and validate `.release.yaml` configuration. Support `release-cli init` to auto-detect and generate config. Minimal required config (`project:` field only).
- `publish-integration`: Publish releases to GitHub Releases (default), with ecosystem-specific registries (npm, PyPI, Maven Central, crates.io). Support artifact uploads.
- `notify-integration`: Send release notifications to Slack, webhooks, and other channels.

### Modified Capabilities

(none — greenfield project)

## Impact

- New Go module with cobra CLI framework
- Dependencies: semver library, git operations, GitHub API client, YAML parser
- New `.release.yaml` config file convention for adopting projects
- Detector interface designed for community contributions via PRs
- Architecture must keep monorepo support feasible for future addition without breaking changes
