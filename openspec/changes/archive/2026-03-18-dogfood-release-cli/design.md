## Context

release-cli is a fully implemented Go CLI for automating release workflows. It supports project detection, semantic versioning, conventional commits, changelog generation, GitHub Releases publishing, and Slack/webhook notifications. However, the tool has never been used to release itself — there is no `.release.yaml`, no CI workflow, and no build/distribution pipeline. Version is already injected via `ldflags` (`internal/cli/root.go:10`), but nothing populates it at build time.

The project lives at `github.com/0x1306e6d/release-cli` and targets linux/darwin on amd64/arm64.

## Goals / Non-Goals

**Goals:**
- release-cli can release itself with `release-cli release`
- Cross-platform binaries are built and attached to GitHub Releases automatically
- CI runs the full pipeline on tag push or manual dispatch
- The `.release.yaml` serves as the canonical example for Go projects

**Non-Goals:**
- Homebrew tap or package manager distribution (future work)
- Docker image publishing
- Windows support in initial release
- Replacing goreleaser for other projects — this uses a Makefile-only approach

## Decisions

### 1. Makefile for cross-compilation (not goreleaser)

Build matrix is encoded in a Makefile with explicit `GOOS`/`GOARCH` targets. Artifacts are `release-cli-<os>-<arch>` tarballs.

**Why over goreleaser**: Zero external dependencies. The build logic is ~30 lines of Make. release-cli's publish step handles the GitHub Release creation; goreleaser would duplicate that responsibility.

**Alternatives considered**:
- **goreleaser**: Full-featured but adds a dependency and overlaps with release-cli's own publish step. Would undermine the dogfooding goal.
- **Shell script**: Less portable and harder to maintain than Make targets.

### 2. GitHub Actions for CI (not other CI systems)

A single `.github/workflows/release.yml` workflow triggers on `v*` tag pushes and `workflow_dispatch`.

**Why**: Project is hosted on GitHub. Actions is zero-config for GitHub Releases permissions via `GITHUB_TOKEN`.

### 3. release-cli publishes its own GitHub Release

The CI workflow builds binaries first, then runs `release-cli release` with the artifacts configured in `.release.yaml`. release-cli creates the GitHub Release and attaches the tarballs.

**Why**: This is the whole point of dogfooding. The pipeline's publish step (GitHub Releases with artifact upload) gets real-world validation.

### 4. Conventional Commits convention

The project will use Conventional Commits (`feat:`, `fix:`, `chore:`, etc.) for commit analysis and changelog generation.

**Why**: Already the most common convention. Aligns with the default `commits.convention: conventional` in release-cli.

### 5. Version from git tags (Go project standard)

Go projects don't have a manifest file for version. release-cli's Go detector already reads version from the latest semver git tag.

**Why**: This is the standard Go versioning approach and is already implemented in the detector.

## Risks / Trade-offs

- **[Bootstrap problem]** The first release requires a pre-built `release-cli` binary since the tool doesn't exist as a release yet. → **Mitigation**: CI workflow builds from source first (`go build`), then uses the freshly built binary to run `release-cli release`. The Makefile's `build` target produces the local binary.

- **[CI token permissions]** GitHub Actions `GITHUB_TOKEN` needs `contents: write` to create releases. → **Mitigation**: Explicitly set permissions in the workflow file.

- **[Tag-triggered loop]** release-cli creates a git tag, which could re-trigger the workflow. → **Mitigation**: The workflow triggers on `v*` tags pushed externally (by a human or `release-cli release` run locally). The CI workflow itself does not push tags — it publishes the release for an existing tag. Alternatively, use `workflow_dispatch` only.
