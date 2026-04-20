## Context

`release-cli` is a Go CLI that already ships a `release.yml` workflow for tag-triggered releases, and the Makefile exposes the standard developer targets (`build`, `test`, `lint`, `vet`). There is no PR-triggered workflow, so quality signals only surface at release time. Contributors run checks locally but nothing enforces them on incoming PRs, and `main` can drift out of green between releases. The change is small and self-contained -- one new YAML file under `.github/workflows/` -- but a short design doc is useful to pin down job layout, trigger scope, and caching strategy before we write the file.

## Goals / Non-Goals

**Goals:**
- Give every pull request a single, deterministic pass/fail signal built from `make vet`, `make test`, `make build`, and `make lint`.
- Detect drift in `go.mod` / `go.sum` on PRs, since an untidy module graph is a common preventable regression.
- Keep feedback fast: leverage `actions/setup-go` caching and cancel stale runs so contributors see results within a few minutes of pushing.
- Reuse the Makefile targets already used locally so CI and developer workflows never diverge.

**Non-Goals:**
- Cross-platform CI (Windows/macOS runners). Linux is sufficient for the Go code here; cross-platform binaries are already produced by `release.yml`'s `release-artifacts` step.
- Multi-version Go matrix. The repo pins the toolchain via `go.mod`, so a single Go version from `go-version-file` is enough.
- Branch-protection configuration. That is a repo-admin action performed after the workflow ships.
- Security scanning, coverage uploads, or release-preview steps -- deliberately out of scope to keep this change minimal.

## Decisions

**Decision 1: One workflow file, multiple jobs.**
Put everything in `.github/workflows/ci.yml` with separate jobs for `build-test`, `lint`, and `tidy`. Rationale: parallel jobs give independent status checks (a lint-only failure does not hide a test failure), and a single file keeps the workflow easy to find next to `release.yml`. *Alternative considered*: one job with sequential steps -- rejected because a failing early step masks later problems.

**Decision 2: Triggers are `pull_request` (against `main`) and `push` to `main`.**
PRs get pre-merge validation; pushes to `main` guard against direct commits or merge-queue merges that skip the PR path. Rationale: we want `main` to stay green under all merge strategies. *Alternative considered*: `pull_request` only -- rejected because direct pushes and merge-queue merges would go unchecked.

**Decision 3: Use `actions/setup-go@v5` with `go-version-file: go.mod` and its built-in cache.**
Rationale: matches the existing `release.yml` pattern for consistency, avoids a separate `actions/cache` step, and auto-tracks the Go toolchain declared in `go.mod`. *Alternative considered*: explicit `actions/cache` on `~/go/pkg/mod` and `~/.cache/go-build` -- rejected as unnecessary duplication of what `setup-go` already does.

**Decision 4: Tidiness check uses `go mod tidy` + `git diff --exit-code`.**
Run `go mod tidy` in CI, then fail if `go.mod` or `go.sum` changes. Rationale: simple, no extra tooling, and the failure message points directly at the offending file. *Alternative considered*: `go mod tidy -diff` (Go 1.23+) -- viable but less portable if the repo's Go version shifts; defer unless we need it.

**Decision 5: `concurrency` group keyed on workflow + ref with `cancel-in-progress: true` for PRs, `false` for `main`.**
Rationale: cancelling superseded PR runs saves runner time, while preserving `main` runs ensures every landed commit is validated end-to-end. *Alternative considered*: cancel everywhere -- rejected because losing a `main` run leaves a gap in the history of green commits.

**Decision 6: Pin actions by major version (`@v4`, `@v5`), matching `release.yml`.**
Rationale: consistency with the existing workflow and the maintainer's revealed preference. SHA-pinning is stricter but adds maintenance overhead disproportionate to a public OSS repo.

## Risks / Trade-offs

- **Risk**: Flaky tests block PRs. → *Mitigation*: The current suite is small and deterministic; revisit with `-count=1` or retry wrappers only if flake shows up.
- **Risk**: `golangci-lint` version drift between contributors and CI. → *Mitigation*: Install `golangci-lint` via the official `golangci/golangci-lint-action` (pinned at `@v6`) instead of relying on whatever `make lint` finds on the runner. The initial workflow uses `version: latest` for the lint tool itself so the first run can discover a version compatible with the repo's Go toolchain; once that run is green, pin `version:` to that exact tag so CI is deterministic and the action's lint-result cache is effective (`latest` defeats the cache key).
- **Risk**: Tidy check false-positives if contributors run an older Go version locally. → *Mitigation*: The failure message is actionable (`run 'go mod tidy'`), and `setup-go` uses the repo's declared version so CI is the source of truth.
- **Trade-off**: Three parallel jobs use slightly more runner minutes than one sequential job, but the faster feedback and clearer status checks are worth it for a small repo.

## Migration Plan

1. Land the new workflow file on a feature branch; the first PR triggers the workflow against itself, providing a live smoke test.
2. Once merged, maintainers can optionally enable branch protection on `main` requiring the three checks (`build-test`, `lint`, `tidy`) to pass before merge. This is a repo-settings change, not part of this code change.
3. Rollback is trivial: delete `.github/workflows/ci.yml`. No downstream consumers depend on the workflow.

## Open Questions

- None blocking implementation. If future needs arise (Go version matrix, coverage reporting, release-preview artifacts), file a separate change.
