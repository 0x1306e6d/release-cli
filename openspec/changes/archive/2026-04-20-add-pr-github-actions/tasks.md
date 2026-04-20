## 1. Scaffold workflow file

- [x] 1.1 Create `.github/workflows/ci.yml` with `name: CI` and top-level `permissions: contents: read`.
- [x] 1.2 Configure triggers: `pull_request` with `branches: [main]` and `push` with `branches: [main]`.
- [x] 1.3 Add a `concurrency` block keyed on `${{ github.workflow }}-${{ github.ref }}` with `cancel-in-progress: ${{ github.event_name == 'pull_request' }}`.

## 2. Build-and-test job

- [x] 2.1 Define job `build-test` on `ubuntu-latest` with `actions/checkout@v4` and `actions/setup-go@v5` using `go-version-file: go.mod`.
- [x] 2.2 Add steps that run `make vet`, `make test`, and `make build` in that order, each as its own named step.
- [x] 2.3 Confirm `setup-go` caching is enabled (default behavior) so module and build caches are reused.

## 3. Lint job

- [x] 3.1 Define job `lint` on `ubuntu-latest` with `actions/checkout@v4` and `actions/setup-go@v5` using `go-version-file: go.mod`.
- [x] 3.2 Use `golangci/golangci-lint-action@v6` pinned to a major version, with `version: latest` or a pinned `golangci-lint` version consistent with local development.
- [x] 3.3 Ensure the action fails the job on any lint issue (default behavior of the action).

## 4. Tidy job

- [x] 4.1 Define job `tidy` on `ubuntu-latest` with `actions/checkout@v4` and `actions/setup-go@v5` using `go-version-file: go.mod`.
- [x] 4.2 Add a step that runs `go mod tidy`.
- [x] 4.3 Add a step that runs `git diff --exit-code -- go.mod go.sum` and prints an instructive message (`run 'go mod tidy' and commit the result`) when the diff is non-empty.

## 5. Validate workflow

- [x] 5.1 Run `openspec validate add-pr-github-actions --strict` and resolve any findings.
- [x] 5.2 Run `yamllint .github/workflows/ci.yml` (or equivalent schema lint) locally to catch syntax issues before pushing.
- [ ] 5.3 Open a draft PR so the new workflow validates itself; confirm all three jobs pass on a clean tree and fail when intentionally broken (e.g., introduce a failing test, revert).
- [ ] 5.4 After the first green run, replace `version: latest` in the `lint` job with the exact `golangci-lint` tag resolved by that run, so CI is deterministic and the action's lint-result cache is effective.

## 6. Documentation

- [x] 6.1 Add a short CI section (or sentence) to `README.md` noting that PRs are validated by the CI workflow, so contributors know what to expect.
- [x] 6.2 Leave branch-protection rule changes to the repo maintainer; do not attempt to change settings from the workflow.
