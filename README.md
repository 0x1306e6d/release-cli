# release-cli

Automate your release workflow from the command line.

release-cli detects your project type, analyzes commits, bumps versions, generates changelogs, creates tags, and publishes GitHub Releases -- all from a single command.

## Features

- **Monorepo support** -- Hierarchical multi-package releases with cascading, batched commits, and namespaced tags
- **Multi-ecosystem support** -- Go, Node.js, Python, Rust, Java (Gradle/Maven), Dart, Helm
- **Commit analysis** -- Conventional Commits, Angular, or custom conventions
- **Semantic versioning** -- Automatic bump calculation from commit history
- **Changelog generation** -- Grouped or flat changelogs with custom template support
- **Version propagation** -- Update version across Dockerfiles, config files, and more
- **GitHub Releases** -- Publish releases with artifact uploads
- **Lifecycle hooks** -- Run scripts at pre-bump, post-bump, pre-publish, post-publish
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

### `release-cli init`

Generates a `.release.yaml` config file with auto-detected project type.

### `release-cli status`

Shows the current version, last release tag, and commits since last release. In monorepo mode, displays a summary table for all packages.

| Flag        | Description                                   |
| ----------- | --------------------------------------------- |
| `--package` | Show status for a specific package (monorepo) |

### `release-cli release`

Executes the full release pipeline: bump version, generate changelog, commit, tag, push, and publish.

| Flag        | Description                                  |
| ----------- | -------------------------------------------- |
| `--bump`    | Override bump level (major, minor, or patch) |
| `--package` | Release specific package(s) in monorepo mode |
| `--all`     | Release all packages (monorepo mode)         |

### Global Flags

| Flag        | Description                       |
| ----------- | --------------------------------- |
| `--dry-run` | Preview actions without executing |
| `--verbose` | Enable verbose output             |
| `--version` | Show version                      |

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

## Configuration

release-cli is configured via `.release.yaml` in the project root. Only the `project` field is required.

### Minimal Example

```yaml
project: go
```

### Full Reference

```yaml
project: go # Required: go, node, python, rust, java-gradle, java-maven, dart, helm
name: my-project # Tag prefix (required when modules is declared, see Monorepo section)
modules: # Child package paths (see Monorepo section)
  - cli
  - lib

version:
  scheme: semver # Version scheme (default: semver)
  snapshot: false # Enable snapshot versions (e.g., -SNAPSHOT for Java, .dev0 for Python)
  manifest: go.mod # Override manifest file
  field: version # Override version field name
  pattern: 'version = "(.+)"' # Override version regex pattern

changes:
  commits:
    convention: custom # Convention: conventional, angular, custom
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
```

Environment variables can be referenced with `${VAR_NAME}` syntax in string values.

### Behavior Without `changes.commits`

When the `changes` section is omitted, all commits are accepted as patch-level changes and the changelog is rendered as a flat list. This is useful for projects that don't follow a structured commit convention.

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

## Monorepo

release-cli supports hierarchical multi-package repositories out of the box. Each package gets independent versioning, changelogs, and releases -- with optional cascading when releasing a parent.

### Configuration

Each package gets its own `.release.yaml`. A parent declares its children via the `modules` field.

```yaml
# .release.yaml (root)
name: my-project
project: go
modules:
  - cli
  - workflow
```

```yaml
# cli/.release.yaml
project: node
```

```yaml
# workflow/.release.yaml
project: go
modules:
  - sub
```

```yaml
# workflow/sub/.release.yaml
project: go
```

Key rules:

- `name` is required when `modules` is declared -- it becomes the git tag prefix (e.g., `my-project/v1.2.0`)
- Child packages use their relative path as the tag prefix (e.g., `cli/v1.0.0`, `workflow/sub/v0.3.0`)
- Each `.release.yaml` is fully independent -- no config inheritance between parent and child
- Selecting a parent for release cascades to all its descendants

### Releasing

```bash
release-cli release --package cli                    # Release a single package
release-cli release --package cli --package workflow  # Release multiple packages
release-cli release --all                            # Release the entire tree
```

When a parent is selected, all descendants are force-released (even those without new commits get a patch bump). A cascading release produces one batched commit with one tag per package:

```
commit: "Release my-project 0.3.0, cli 0.3.0, workflow 0.1.5"
  -> tag: my-project/v0.3.0
  -> tag: cli/v0.3.0
  -> tag: workflow/v0.1.5
```

Commit analysis is path-scoped: only commits touching files under a package's directory contribute to its bump and changelog. Parent packages see all commits under their path, including children's.

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

### Monorepo Release

```yaml
name: Release Package

on:
  workflow_dispatch:
    inputs:
      package:
        description: "Package to release (or 'all')"
        required: true
        type: string
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
        run: |
          if [ "${{ inputs.package }}" = "all" ]; then
            release-cli release --all ${{ inputs.bump && format('--bump {0}', inputs.bump) || '' }}
          else
            release-cli release --package ${{ inputs.package }} ${{ inputs.bump && format('--bump {0}', inputs.bump) || '' }}
          fi
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Lifecycle Hooks

Hooks run shell commands at key stages of the release pipeline. The following environment variables are injected:

| Variable               | Description                                        |
| ---------------------- | -------------------------------------------------- |
| `RELEASE_VERSION`      | New version being released                         |
| `RELEASE_PREV_VERSION` | Previous version                                   |
| `RELEASE_PROJECT`      | Project type                                       |
| `RELEASE_PACKAGE`      | Package name (monorepo only)                       |
| `RELEASE_PACKAGE_PATH` | Package path relative to repo root (monorepo only) |

A non-zero exit code from any hook aborts the release.

## Environment Variables

| Variable       | Description                            |
| -------------- | -------------------------------------- |
| `GITHUB_TOKEN` | Required for GitHub Release publishing |
