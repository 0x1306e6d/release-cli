## ADDED Requirements

### Requirement: GitHub Actions workflow for automated releases
The repository SHALL include a `.github/workflows/release.yml` workflow that automates the build-and-release pipeline.

#### Scenario: Workflow triggers on version tags
- **WHEN** a tag matching `v*` is pushed to the repository
- **THEN** the release workflow SHALL be triggered

#### Scenario: Workflow supports manual dispatch
- **WHEN** a user triggers the workflow via `workflow_dispatch`
- **THEN** the release workflow SHALL execute the same pipeline as a tag-triggered run

#### Scenario: Workflow builds cross-platform binaries
- **WHEN** the release workflow runs
- **THEN** it SHALL execute the Makefile `release-artifacts` target to produce binaries for all supported platforms

#### Scenario: Workflow runs release-cli release
- **WHEN** binaries have been built
- **THEN** the workflow SHALL run `release-cli release` using the freshly built binary to create the GitHub Release and attach artifacts

### Requirement: Workflow has correct permissions
The workflow SHALL declare explicit permissions required for creating GitHub Releases.

#### Scenario: Contents write permission
- **WHEN** the workflow executes
- **THEN** the `permissions` block SHALL include `contents: write`

### Requirement: Workflow builds release-cli from source
The workflow SHALL build release-cli from the repository source before using it, solving the bootstrap problem.

#### Scenario: Bootstrap build
- **WHEN** the workflow starts
- **THEN** it SHALL run `go build` to produce a local `release-cli` binary before running `release-cli release`
