## MODIFIED Requirements

### Requirement: Parse commits since last release
The system SHALL parse all commits between the last release tag and HEAD to determine the bump type and generate changelog content. In monorepo mode, the system SHALL filter commits to only those that touch files under the package's path directory, and SHALL use the package's namespaced tag (e.g., `cli/v1.2.3`) as the starting point. Parent packages see all commits under their path, including children's.

#### Scenario: Commits since last tag (single-project)
- **WHEN** the last release tag is `v1.2.3` and there are 5 commits since
- **THEN** all 5 commits are parsed according to the configured convention

#### Scenario: Commits since last tag (child package)
- **WHEN** package `cli` has last tag `cli/v1.2.3`
- **AND** there are 10 commits since that tag, 3 of which touch files under `cli/`
- **THEN** only the 3 path-matching commits are parsed for the `cli` package

#### Scenario: Commits since last tag (parent package)
- **WHEN** root package `release-cli` has last tag `release-cli/v1.0.0`
- **AND** there are 10 commits since that tag touching various paths including `cli/` and `workflow/`
- **THEN** all 10 commits are included in the root package's analysis

#### Scenario: No previous release tag (single-project)
- **WHEN** no release tags exist in the repository
- **THEN** all commits from the initial commit to HEAD are parsed

#### Scenario: No previous release tag (monorepo)
- **WHEN** package `cli` has no existing namespaced tags
- **THEN** all commits from the initial commit that touch files under `cli/` are parsed

### Requirement: No releasable commits detected
When no commits since the last release match the configured convention for any bump type, the system SHALL report that no release is needed and exit without error. When `--bump` is provided, this check SHALL be skipped. In cascading mode, unchanged descendants SHALL receive a patch bump (force release).

#### Scenario: No releasable changes (single-project)
- **WHEN** all commits since the last release are non-conventional (e.g., `update docs`, `cleanup`)
- **AND** no `--bump` flag is provided
- **THEN** the system reports "No releasable changes found" and exits with code 0

#### Scenario: Force release of unchanged descendant in cascade
- **WHEN** a parent package is selected for release
- **AND** child package `cli` has no commits since its last tag
- **THEN** `cli` is released with a patch bump

#### Scenario: No releasable changes but bump override provided
- **WHEN** all commits since the last release are non-conventional
- **AND** `--bump patch` is provided
- **THEN** the system proceeds with the release using patch bump
