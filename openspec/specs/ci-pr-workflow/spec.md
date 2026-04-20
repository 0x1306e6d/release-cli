## ADDED Requirements

### Requirement: GitHub Actions workflow for pull requests
The repository SHALL include a `.github/workflows/ci.yml` workflow that validates changes before they merge to `main`.

#### Scenario: Workflow triggers on pull requests to main
- **WHEN** a pull request is opened, synchronized, or reopened against the `main` branch
- **THEN** the CI workflow SHALL be triggered

#### Scenario: Workflow triggers on pushes to main
- **WHEN** a commit is pushed directly to the `main` branch
- **THEN** the CI workflow SHALL be triggered to validate the resulting tree

#### Scenario: Workflow does not run on unrelated tag pushes
- **WHEN** a tag matching `v*` is pushed
- **THEN** the CI workflow SHALL NOT be triggered, leaving tag handling to `release.yml`

### Requirement: Workflow runs build and test checks
The workflow SHALL execute the Go build and test targets exposed by the repository Makefile.

#### Scenario: Build succeeds
- **WHEN** the `build-test` job runs
- **THEN** it SHALL run `make build` and the step SHALL fail the job on non-zero exit

#### Scenario: Unit tests run
- **WHEN** the `build-test` job runs
- **THEN** it SHALL run `make test` and the step SHALL fail the job on any failing test

#### Scenario: Go vet runs
- **WHEN** the `build-test` job runs
- **THEN** it SHALL run `make vet` and the step SHALL fail the job on any reported issue

### Requirement: Workflow runs lint checks
The workflow SHALL run `golangci-lint` in a dedicated job so lint failures are reported as a distinct status check.

#### Scenario: Lint job runs golangci-lint
- **WHEN** the `lint` job runs
- **THEN** it SHALL execute `golangci-lint` against the repository and fail the job on any reported issue

#### Scenario: Lint job uses a pinned golangci-lint version
- **WHEN** the `lint` job installs `golangci-lint`
- **THEN** it SHALL use a pinned major version of `golangci/golangci-lint-action` so CI results are reproducible

### Requirement: Workflow verifies module tidiness
The workflow SHALL detect drift in `go.mod` and `go.sum` caused by forgotten `go mod tidy` runs.

#### Scenario: Tidy check passes when go.mod is clean
- **WHEN** `go mod tidy` produces no changes to `go.mod` or `go.sum`
- **THEN** the `tidy` job SHALL succeed

#### Scenario: Tidy check fails on drift
- **WHEN** `go mod tidy` produces changes to `go.mod` or `go.sum`
- **THEN** the `tidy` job SHALL fail with a message instructing the contributor to run `go mod tidy` locally

### Requirement: Workflow uses cached Go toolchain
The workflow SHALL use `actions/setup-go@v5` with `go-version-file: go.mod` and its built-in module/build cache in every job that invokes Go.

#### Scenario: setup-go configures the toolchain from go.mod
- **WHEN** any job that runs Go tooling starts
- **THEN** it SHALL use `actions/setup-go@v5` with `go-version-file: go.mod`

#### Scenario: Go caches are reused across runs
- **WHEN** a job runs after a previous successful run on the same ref
- **THEN** `actions/setup-go` SHALL restore the module cache and build cache without an explicit `actions/cache` step

### Requirement: Workflow cancels stale pull request runs
The workflow SHALL use a `concurrency` group so that superseded runs on the same pull request are cancelled, while runs on `main` are preserved.

#### Scenario: Superseded PR run is cancelled
- **WHEN** a new commit is pushed to a pull request while an earlier run for the same PR is still in progress
- **THEN** the earlier run SHALL be cancelled and only the latest commit's run SHALL continue

#### Scenario: Main branch runs are never cancelled
- **WHEN** a new commit lands on `main` while a previous `main` run is still executing
- **THEN** both runs SHALL complete, preserving a validated history of `main`

### Requirement: Workflow declares least-privilege permissions
The workflow SHALL declare explicit `permissions` so it runs with the minimum scope required for read-only validation.

#### Scenario: Default contents read permission
- **WHEN** the workflow executes
- **THEN** the top-level `permissions` block SHALL set `contents: read` and SHALL NOT request write scopes
