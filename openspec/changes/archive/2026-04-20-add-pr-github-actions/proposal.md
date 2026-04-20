## Why

The repository currently has only a release workflow (`release.yml`) and no automated quality checks for pull requests. Contributors and maintainers have no pre-merge signal that a PR builds, passes tests, lints cleanly, or keeps `go.mod` tidy, so regressions can land on `main` and only be caught during release. A PR-triggered GitHub Actions workflow closes that gap and gives reviewers a consistent green/red status before merge.

## What Changes

- Add a `.github/workflows/ci.yml` workflow that runs on `pull_request` events against `main` and on `push` to `main`.
- Run the standard Makefile quality targets inside the workflow: `make vet`, `make test`, and `make build` to exercise the Go toolchain the same way contributors do locally.
- Run `golangci-lint` via `make lint` as a separate job so lint failures are distinguishable from test failures.
- Verify `go.mod` / `go.sum` are tidy by running `go mod tidy` and failing if the working tree becomes dirty.
- Cache Go build and module caches via `actions/setup-go`'s built-in caching to keep runs fast.
- Use `concurrency` to cancel superseded runs on the same PR so only the latest commit consumes runners.

No changes to existing release behavior; the new workflow is additive.

## Capabilities

### New Capabilities
- `ci-pr-workflow`: GitHub Actions workflow that validates pull requests by running build, vet, test, lint, and module-tidiness checks.

### Modified Capabilities

None. The existing `ci-release-workflow` spec is untouched; this change adds a separate PR-focused workflow.

## Impact

- **New file**: `.github/workflows/ci.yml`.
- **No code changes** in `cmd/`, `internal/`, or `Makefile` (the workflow consumes existing Makefile targets).
- **Branch protection**: maintainers can later require the new checks on `main`; this proposal does not change repository settings.
- **Runner cost**: one workflow run per PR push, scoped to `ubuntu-latest`; concurrency cancellation limits duplicate runs.
