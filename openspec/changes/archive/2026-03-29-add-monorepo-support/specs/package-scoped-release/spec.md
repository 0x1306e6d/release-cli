## ADDED Requirements

### Requirement: Path-based commit filtering
In monorepo mode, the system SHALL filter commits per package by only including commits that touch files under the package's directory. The filtering SHALL use `git log -- <path>` semantics. A parent package sees all commits under its path, including those in children's directories.

#### Scenario: Commits scoped to child package
- **WHEN** package `cli` is at path `cli/`
- **AND** there are 10 commits since the last release, 3 of which touch files under `cli/`
- **THEN** only the 3 relevant commits are analyzed for the `cli` package

#### Scenario: Parent sees all commits including children's
- **WHEN** the root package is at path `.`
- **AND** there are 10 commits since the last release: 3 touch `cli/`, 2 touch `workflow/`, 5 touch root-level files
- **THEN** all 10 commits are included in the root package's analysis

#### Scenario: Commit touches multiple packages
- **WHEN** a single commit modifies files in both `cli/` and `workflow/`
- **THEN** the commit contributes to both the `cli` and `workflow` package analyses independently

### Requirement: Namespaced git tags per package
In monorepo mode, git tags SHALL use the following format:
- Root parent: `<name>/v<version>` (e.g., `release-cli/v1.2.0`)
- Child package: `<relative-path>/v<version>` (e.g., `cli/v1.0.0`, `workflow/sub/v0.3.0`)

The system SHALL read the latest tag matching the package prefix to determine the current version.

#### Scenario: Tag creation for root package
- **WHEN** the root package with `name: release-cli` is released with version `1.2.0`
- **THEN** the system creates git tag `release-cli/v1.2.0`

#### Scenario: Tag creation for child package
- **WHEN** child package at `cli/` is released with version `0.5.0`
- **THEN** the system creates git tag `cli/v0.5.0`

#### Scenario: Tag creation for deeply nested package
- **WHEN** child package at `workflow/sub/` is released with version `1.0.0`
- **THEN** the system creates git tag `workflow/sub/v1.0.0`

#### Scenario: Reading latest version from namespaced tags
- **WHEN** the repository has tags `cli/v1.0.0`, `cli/v1.1.0`, `workflow/v0.2.0`
- **AND** the system reads the latest version for package `cli`
- **THEN** it returns `1.1.0` (from tag `cli/v1.1.0`)

#### Scenario: No previous tag for package
- **WHEN** package `cli` has no existing namespaced tags
- **THEN** all commits touching `cli/` from the initial commit are included

### Requirement: Hierarchical cascading forces all descendants
When a parent package is selected for release, the system SHALL force-release all its descendants, even those without new commits since their last tag. Packages without commits SHALL receive a patch bump by default (unless `--bump` overrides).

#### Scenario: Cascading from root
- **WHEN** `release-cli release --package release-cli` is run
- **AND** the root has modules `cli` and `workflow`
- **THEN** the system releases `release-cli`, `cli`, and `workflow` — all three

#### Scenario: Cascading from intermediate parent
- **WHEN** `release-cli release --package workflow` is run
- **AND** `workflow` has child `sub`
- **THEN** the system releases `workflow` and `sub`, but not the root or `cli`

#### Scenario: Force release of unchanged descendant
- **WHEN** the root is selected for release
- **AND** `cli` has no commits since its last tag
- **THEN** `cli` is still released with a patch bump

### Requirement: Batched commit with multiple tags
A cascading release SHALL produce one git commit containing all version bumps and changelogs for all affected packages, with one tag per package pointing to that commit.

#### Scenario: Batched cascading release
- **WHEN** the root cascades to `cli` and `workflow`
- **THEN** the system creates one commit: "Release release-cli 0.3.0, cli 0.3.0, workflow 0.1.5"
- **AND** creates three tags on that commit: `release-cli/v0.3.0`, `cli/v0.3.0`, `workflow/v0.1.5`

#### Scenario: Single leaf package release
- **WHEN** `release-cli release --package cli` is run and `cli` has no children
- **THEN** the system creates one commit: "Release cli 0.3.0"
- **AND** creates one tag: `cli/v0.3.0`

### Requirement: CLI --package flag for targeted release
The `release` command SHALL accept `--package <name>` flags, allowing multiple values, to release specific packages in monorepo mode. If a selected package has children, the release cascades.

#### Scenario: Release specific leaf package
- **WHEN** `release-cli release --package cli` is run
- **THEN** the system runs the release pipeline only for the `cli` package

#### Scenario: Release multiple packages
- **WHEN** `release-cli release --package cli --package workflow` is run
- **THEN** the system releases both `cli` and `workflow` (and `workflow`'s descendants if any)

#### Scenario: Unknown package name
- **WHEN** `release-cli release --package nonexistent` is run
- **THEN** the system SHALL report an error: `package "nonexistent" not found in config`

#### Scenario: --package flag in single-project mode
- **WHEN** `release-cli release --package cli` is run in single-project mode
- **THEN** the system SHALL report an error: `--package flag requires monorepo mode (use "modules" in config)`

### Requirement: CLI --all flag as root selection
The `release` command SHALL accept an `--all` flag, which is equivalent to selecting the root package. This cascades to all packages in the tree.

#### Scenario: --all releases entire tree
- **WHEN** `release-cli release --all` is run
- **THEN** the system releases the root and all its descendants

#### Scenario: --all flag in single-project mode
- **WHEN** `release-cli release --all` is run in single-project mode
- **THEN** the system SHALL report an error: `--all flag requires monorepo mode (use "modules" in config)`

### Requirement: Monorepo release requires explicit target
In monorepo mode, `release-cli release` without `--package` or `--all` SHALL report an error asking the user to specify a target.

#### Scenario: Release without target in monorepo mode
- **WHEN** `release-cli release` is run in monorepo mode without `--package` or `--all`
- **THEN** the system SHALL report an error: `monorepo mode requires --package <name> or --all`

### Requirement: Status command shows package information
In monorepo mode, `release-cli status` without `--package` SHALL display a summary table of all packages in the tree showing: package name, current version, pending commits, and next bump type. With `--package`, it SHALL show detailed status for that package only.

#### Scenario: Status summary for all packages
- **WHEN** `release-cli status` is run in monorepo mode
- **THEN** the system displays a table with columns: Package, Version, Pending Commits, Next Bump

#### Scenario: Status for specific package
- **WHEN** `release-cli status --package cli` is run
- **THEN** the system displays detailed status for the `cli` package (same as single-project status output)

### Requirement: Per-package pipeline execution
Each package in a release SHALL run its own pipeline instance with operations scoped to the package: detector resolved against package path, version read from package manifest or namespaced tags, commits filtered by path, changelog written relative to package path. The commit and tag steps are batched across all packages in a cascading release.

#### Scenario: Pipeline scoped to package path
- **WHEN** the release pipeline runs for package `cli` with path `cli/` and `project: go`
- **THEN** the detector checks `cli/` for `go.mod`
- **AND** version is read from tags prefixed with `cli/`
- **AND** commits are filtered to `cli/`
- **AND** changelog is written to `cli/CHANGELOG.md`

#### Scenario: Pipeline execution order in cascade
- **WHEN** the root cascades to `cli` and `workflow`
- **THEN** each package's detect/analyze/bump/propagate/changelog steps run independently
- **AND** then one batched commit is created with all changes
- **AND** then one tag per package is created on that commit

### Requirement: Per-package hook environment variables
In monorepo mode, hook commands SHALL have access to additional environment variables: `RELEASE_PACKAGE` (package name or relative path) and `RELEASE_PACKAGE_PATH` (package path), in addition to the existing `RELEASE_VERSION`, `RELEASE_PREV_VERSION`, and `RELEASE_PROJECT`.

#### Scenario: Hook accesses package variables
- **WHEN** a hook command is `echo "Releasing $RELEASE_PACKAGE at $RELEASE_PACKAGE_PATH"`
- **AND** the package is at path `cli/`
- **THEN** the command outputs `Releasing cli at cli`

### Requirement: Dry-run with --package and --all
The `--dry-run` flag SHALL work with both `--package` and `--all` in monorepo mode, showing what would happen for the targeted package(s) including cascading.

#### Scenario: Dry-run for cascading release
- **WHEN** `release-cli release --package release-cli --dry-run` is run
- **THEN** the system shows a preview of the release for `release-cli` and all its descendants
