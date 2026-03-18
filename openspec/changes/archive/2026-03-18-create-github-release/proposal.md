## Why

After running `release-cli release`, a git tag and commit are pushed to the remote, but no GitHub Release is created. The `GitHubPublisher` implementation already exists in `internal/publish/github.go` but the pipeline never calls it — the publish step at pipeline.go:150 is a placeholder. This means release artifacts (cross-platform binaries) are never uploaded to GitHub Releases, and users have no visible release page on GitHub.

## What Changes

- Wire the existing `GitHubPublisher` into the pipeline's publish step so that a GitHub Release is created with the changelog body and configured artifacts are uploaded.
- Parse the git remote URL to derive repository owner and name automatically.
- Read `GITHUB_TOKEN` from the environment to authenticate with the GitHub API.
- Add dry-run reporting for the publish step.

## Capabilities

### New Capabilities

_(none — this wires an existing, already-implemented capability)_

### Modified Capabilities

- `publish-integration`: The publish step transitions from placeholder to active. The pipeline now creates a GitHub Release and uploads artifacts when `publish.github.enabled` is true (the default).

## Impact

- **Code**: `internal/pipeline/pipeline.go` — replace publish placeholder with `GitHubPublisher` call. New helper to parse owner/repo from git remote in `internal/git/`.
- **CI**: No workflow changes needed — `GITHUB_TOKEN` is already passed as an env var.
- **Config**: No schema changes — `publish.github` config is already defined and defaults to enabled.
- **Dependencies**: No new external dependencies.
