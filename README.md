# release-cli

Automate your release workflow from the command line.

release-cli detects your project type, analyzes commits, bumps versions, generates changelogs, creates tags, publishes GitHub Releases, and sends notifications -- all from a single command.

## Features

- **Multi-ecosystem support** -- Go, Node.js, Python, Rust, Java (Gradle/Maven), Dart, Helm
- **Commit analysis** -- Conventional Commits, Angular, or custom conventions
- **Semantic versioning** -- Automatic bump calculation from commit history
- **Changelog generation** -- Grouped or flat changelogs with custom template support
- **Version propagation** -- Update version across Dockerfiles, config files, and more
- **GitHub Releases** -- Publish releases with artifact uploads
- **Lifecycle hooks** -- Run scripts at pre-bump, post-bump, pre-publish, post-publish
- **Snapshot mode** -- Development versions (e.g., `-SNAPSHOT`, `.dev0`)
- **Dry-run mode** -- Preview the entire release without executing

## Installation

### From source

```bash
go install github.com/0x1306e6d/release-cli/cmd/release-cli@latest
```

### Build locally

```bash
git clone https://github.com/0x1306e6d/release-cli.git
cd release-cli
make install
```

## Quick Start

Initialize a config file in your project:

```bash
release-cli init
```

This creates `.release.yaml` with auto-detected project type and sensible defaults.

Preview what will happen:

```bash
release-cli status
```

Run the release:

```bash
release-cli release
```

## Commands

### `release`

Executes the full release pipeline: bump version, generate changelog, commit, tag, push, and publish.

```bash
release-cli release [flags]
```

| Flag     | Description                                        |
| -------- | -------------------------------------------------- |
| `--bump` | Override bump level (`major`, `minor`, or `patch`) |

### `init`

Generates a `.release.yaml` config file with auto-detected project type.

```bash
release-cli init
```

### `status`

Shows the current version, last release tag, commits since last release, and a preview of the next version.

```bash
release-cli status
```

### Global Flags

| Flag        | Description                       |
| ----------- | --------------------------------- |
| `--dry-run` | Preview actions without executing |
| `--verbose` | Enable verbose output             |
| `--version` | Show version                      |

## Configuration

release-cli is configured via `.release.yaml` in the project root. Only the `project` field is required.

### Minimal Example

```yaml
project: go
```

### Full Reference

```yaml
project: go # Required: go, node, python, rust, java-gradle, java-maven, dart, helm

version:
  scheme: semver # Version scheme (default: semver)
  snapshot: false # Enable snapshot versions
  manifest: go.mod # Override manifest file
  field: version # Override version field name
  pattern: 'version = "(.+)"' # Override version regex pattern

changes:
  commits:
    convention: conventional # Convention: conventional, angular, custom
    types: # Custom type mappings (only for custom convention)
      major: [breaking]
      minor: [feat, feature]
      patch: [fix, bugfix]

changelog:
  enabled: true # Generate changelog (default: true)
  file: CHANGELOG.md # Changelog file path (default: CHANGELOG.md)
  template: | # Custom Go template (optional)
    ## {{ .Version }} ({{ .Date }})
    {{ range .Groups }}...

propagate: # Propagate version to other files
  - file: Dockerfile
    type: docker-label # Built-in type
  - file: config.json
    field: app.version # JSON/YAML field path
  - file: version.txt
    pattern: "VERSION=(.+)" # Regex with capture group

hooks:
  pre-bump: ./scripts/validate.sh
  post-bump: ./scripts/build.sh
  pre-publish: ./scripts/docs.sh
  post-publish: ./scripts/notify.sh

publish:
  github:
    enabled: true # Default: true
    draft: false # Create as draft release
    artifacts: # Glob patterns for upload
      - dist/*.tar.gz

notify:
  slack:
    webhook: ${SLACK_WEBHOOK_URL}
    channel: "#releases"
  webhook:
    url: ${WEBHOOK_URL}
```

Environment variables can be referenced with `${VAR_NAME}` syntax in string values.

### Behavior Without `changes.commits`

When the `changes` section is omitted, all commits are accepted as patch-level changes and the changelog is rendered as a flat list. This is useful for projects that don't follow a structured commit convention.

## Supported Projects

| Type          | Detection File                      | Version Source      |
| ------------- | ----------------------------------- | ------------------- |
| `go`          | `go.mod`                            | Git tags            |
| `node`        | `package.json`                      | `version` field     |
| `python`      | `pyproject.toml`                    | `version` field     |
| `rust`        | `Cargo.toml`                        | `version` field     |
| `java-gradle` | `build.gradle` / `build.gradle.kts` | `gradle.properties` |
| `java-maven`  | `pom.xml`                           | `<version>` element |
| `dart`        | `pubspec.yaml`                      | `version` field     |
| `helm`        | `Chart.yaml`                        | `version` field     |

## Commit Conventions

### Conventional Commits

```
feat(auth): add OAuth2 support
fix: resolve null pointer in parser
feat!: redesign API response format
```

- `feat` -> minor
- `fix` -> patch
- `!` or `BREAKING CHANGE` footer -> major

### Angular

```
feat(core): add new endpoint
fix(auth): handle expired tokens
perf(db): optimize query execution
```

- `feat` -> minor
- `fix`, `perf` -> patch
- Breaking changes -> major

### Custom

Define your own type-to-bump mappings:

```yaml
changes:
  commits:
    convention: custom
    types:
      major: [breaking]
      minor: [feat, feature]
      patch: [fix, bugfix]
```

## CI/CD

### GitHub Actions

```yaml
name: Release

on:
  push:
    tags:
      - "v*"
  workflow_dispatch:
    inputs:
      bump:
        description: "Version bump level"
        required: false
        type: choice
        options:
          - ""
          - patch
          - minor
          - major

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install release-cli
        run: go install github.com/0x1306e6d/release-cli/cmd/release-cli@latest

      - name: Configure git identity
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

      - name: Release
        run: release-cli release ${{ inputs.bump && format('--bump {0}', inputs.bump) || '' }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Environment Variables

| Variable            | Description                                               |
| ------------------- | --------------------------------------------------------- |
| `GITHUB_TOKEN`      | Required for GitHub Release publishing                    |
| `SLACK_WEBHOOK_URL` | Slack incoming webhook URL (if using Slack notifications) |
| `WEBHOOK_URL`       | Generic webhook URL (if using webhook notifications)      |

## Lifecycle Hooks

Hooks run shell commands at key stages of the release pipeline. The following variables are injected:

| Variable           | Description                |
| ------------------ | -------------------------- |
| `VERSION`          | New version being released |
| `PREVIOUS_VERSION` | Previous version           |
| `PROJECT`          | Project type               |

A non-zero exit code from any hook aborts the release.
