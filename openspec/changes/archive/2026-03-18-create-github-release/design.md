## Context

The release pipeline (`internal/pipeline/pipeline.go`) has a placeholder at step 13 where publishing should happen. The `GitHubPublisher` in `internal/publish/github.go` is fully implemented — it can create GitHub Releases via the API and upload artifacts. The `PublishConfig` in `internal/config/config.go` already has fields for `enabled`, `draft`, and `artifacts`. The CI workflow already passes `GITHUB_TOKEN` as an environment variable.

The missing piece is glue code: the pipeline needs to construct a `GitHubPublisher` from the config + environment, derive the repo owner/name from the git remote, and call `Publish()`.

## Goals / Non-Goals

**Goals:**
- Wire `GitHubPublisher` into the pipeline so GitHub Releases are created automatically
- Derive repository owner and name from the git remote URL (no new config fields)
- Support dry-run mode for the publish step
- Handle the case where `GITHUB_TOKEN` is not set (skip publish with a warning)

**Non-Goals:**
- Adding new publish targets (npm, PyPI, etc.)
- Wiring the notify step (separate change)
- Changing the `GitHubPublisher` API or adding new config fields

## Decisions

### 1. Derive owner/repo from git remote URL

**Decision**: Add a `RemoteOwnerRepo(dir string)` function in `internal/git/` that parses the `origin` remote URL to extract owner and repo name.

**Rationale**: The owner/repo information is already available from the git remote. Requiring users to configure it separately would be redundant and error-prone. Both HTTPS (`https://github.com/owner/repo.git`) and SSH (`git@github.com:owner/repo.git`) URL formats must be handled.

**Alternatives considered**:
- Add `owner` and `repo` config fields — rejected because it duplicates what git already knows and adds configuration burden.
- Use `GITHUB_REPOSITORY` env var (set by GitHub Actions) — rejected because it ties the tool to GitHub Actions; the git remote approach works everywhere.

### 2. Skip publish gracefully when token is missing

**Decision**: If `GITHUB_TOKEN` is not set and publish is enabled, log a warning and skip the publish step rather than failing the pipeline.

**Rationale**: Users may run `release-cli release` locally during development without a GitHub token. Failing hard would break the local workflow. The warning makes it clear that the GitHub Release was not created.

### 3. Pass changelog content through the pipeline

**Decision**: Pass the `changelogContent` string (already computed at step 9) to the publish step as the release body.

**Rationale**: The changelog content is already generated and available in the pipeline's local scope. No additional plumbing is needed — just pass it to `ReleaseInfo.ChangelogBody`.

## Risks / Trade-offs

- **[Risk] Remote URL parsing edge cases** → Mitigation: Support both HTTPS and SSH formats, strip `.git` suffix. If parsing fails, skip publish with a warning rather than crashing.
- **[Risk] Token permissions insufficient** → Mitigation: The CI workflow already has `contents: write`. The error from the GitHub API will be descriptive if permissions are wrong.
