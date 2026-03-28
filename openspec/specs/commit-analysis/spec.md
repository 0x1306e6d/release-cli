## ADDED Requirements

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

### Requirement: Conventional Commits as default convention
The system SHALL support Conventional Commits as the default commit convention, mapping commit types to bump levels:
- `feat` → minor bump
- `fix` → patch bump
- `BREAKING CHANGE` footer or `!` suffix on type → major bump

#### Scenario: Feature commit triggers minor bump
- **WHEN** commits since last release include `feat: add user export`
- **THEN** the determined bump type is `minor`

#### Scenario: Fix commit triggers patch bump
- **WHEN** commits since last release include only `fix: correct null pointer`
- **THEN** the determined bump type is `patch`

#### Scenario: Breaking change triggers major bump
- **WHEN** commits since last release include `feat!: redesign API`
- **THEN** the determined bump type is `major`

#### Scenario: Breaking change in footer triggers major bump
- **WHEN** a commit message contains a `BREAKING CHANGE:` footer
- **THEN** the determined bump type is `major`

### Requirement: Highest bump type wins
When multiple commits have different bump implications, the system SHALL use the highest bump type (major > minor > patch).

#### Scenario: Mixed commits
- **WHEN** commits include both `fix: typo` and `feat: new feature`
- **THEN** the determined bump type is `minor` (higher than patch)

### Requirement: Configurable commit convention
The system SHALL allow configuring the commit convention in `.release.yaml` under the `changes.commits` key. Supported values SHALL include `conventional`, `angular`, and `custom`. A `custom` option SHALL allow defining arbitrary type-to-bump mappings. When `changes.commits` is absent, the system SHALL accept all commits as releasable with patch bump and flat changelog.

#### Scenario: Custom commit convention
- **WHEN** the config specifies `changes.commits.convention: custom` with `minor: ["feature", "enhancement"]`
- **AND** a commit message starts with `feature: add dark mode`
- **THEN** the determined bump type is `minor`

#### Scenario: Angular convention
- **WHEN** the config specifies `changes.commits.convention: angular`
- **AND** commits follow Angular commit format
- **THEN** the system parses them according to Angular conventions

#### Scenario: No changes section defaults to accept-all behavior
- **WHEN** the config does not include a `changes` section
- **THEN** the system SHALL accept all commits as releasable with patch bump

### Requirement: Non-conforming commits are ignored for bump calculation
Commits that do not match the configured convention SHALL be ignored when determining the bump type but MAY be included in the changelog under an "Other" category.

#### Scenario: Non-conventional commit
- **WHEN** a commit message is `updated readme`
- **THEN** it does not contribute to the bump type calculation
- **AND** it MAY appear in the changelog under "Other"

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
