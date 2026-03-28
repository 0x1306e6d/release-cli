## MODIFIED Requirements

### Requirement: Fixed pipeline step order
The release pipeline SHALL execute steps in the following fixed order: detect → analyze commits → apply bump override (if provided) → bump manifest → propagate → generate changelog → commit → tag → publish → notify. The user SHALL NOT be able to reorder steps. In monorepo mode, each package runs its own pipeline steps (detect through changelog) independently, then commit and tag steps are batched across all packages.

#### Scenario: Full pipeline execution (single-project)
- **WHEN** `release-cli release` is run in single-project mode
- **THEN** steps execute in order: detect, analyze, bump override, bump manifest, propagate, changelog, commit, tag, publish, notify

#### Scenario: Full pipeline execution (monorepo, single package)
- **WHEN** `release-cli release --package cli` is run for a leaf package
- **THEN** steps execute in the same order, scoped to `cli`: detect against `cli/`, analyze commits filtered by `cli/`, bump `cli/` manifest, propagate within `cli/`, per-package changelog, commit, namespaced tag `cli/v<version>`, publish, notify

#### Scenario: Cascading release with batched commit
- **WHEN** `release-cli release --package release-cli` cascades to `cli` and `workflow`
- **THEN** detect/analyze/bump/propagate/changelog run for each package independently
- **AND** one batched commit is created with all changes: "Release release-cli 0.3.0, cli 0.3.0, workflow 0.1.5"
- **AND** three tags are created on that commit: `release-cli/v0.3.0`, `cli/v0.3.0`, `workflow/v0.1.5`
- **AND** publish and notify run per-package after tagging

#### Scenario: Pipeline with bump override
- **WHEN** `release-cli release --bump minor` is run
- **THEN** commits are analyzed for changelog content
- **AND** the bump override replaces the convention-derived bump
- **AND** the pipeline continues with the overridden bump level

### Requirement: Hook commands have access to version variables
Hook commands SHALL have access to release context via environment variables: `RELEASE_VERSION` (new version), `RELEASE_PREV_VERSION` (current version), `RELEASE_PROJECT` (project type). In monorepo mode, hooks SHALL additionally receive `RELEASE_PACKAGE` (package name or relative path) and `RELEASE_PACKAGE_PATH` (package path).

#### Scenario: Hook uses version variable
- **WHEN** a hook command is `echo "Releasing $RELEASE_VERSION"`
- **AND** the new version is `1.4.0`
- **THEN** the command outputs `Releasing 1.4.0`

#### Scenario: Hook uses package variables in monorepo mode
- **WHEN** a hook command is `echo "$RELEASE_PACKAGE at $RELEASE_PACKAGE_PATH"`
- **AND** the package is `cli` at path `cli/`
- **THEN** the command outputs `cli at cli`

### Requirement: SNAPSHOT post-release step
When snapshot mode is enabled, the pipeline SHALL include an additional step after tagging that bumps the manifest to the next development version and creates a "prepare next development iteration" commit. In monorepo cascading mode, snapshot bumps for all packages SHALL be batched into one commit.

#### Scenario: Snapshot step after release (single-project)
- **WHEN** snapshot mode is enabled
- **AND** the release version is `1.4.0`
- **THEN** after the release tag is created, the manifest is bumped to `1.5.0-SNAPSHOT` and a commit "Prepare next development iteration" is created

#### Scenario: Snapshot step after cascading release (monorepo)
- **WHEN** snapshot mode is enabled for packages `release-cli` and `cli`
- **AND** the release versions are `1.4.0` and `0.5.0` respectively
- **THEN** after the release tags are created, both manifests are bumped to their next snapshot versions in one commit "Prepare next development iteration"
