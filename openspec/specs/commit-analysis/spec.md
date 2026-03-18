## ADDED Requirements

### Requirement: Parse commits since last release
The system SHALL parse all commits between the last release tag and HEAD to determine the bump type and generate changelog content.

#### Scenario: Commits since last tag
- **WHEN** the last release tag is `v1.2.3` and there are 5 commits since
- **THEN** all 5 commits are parsed according to the configured convention

#### Scenario: No previous release tag
- **WHEN** no release tags exist in the repository
- **THEN** all commits from the initial commit to HEAD are parsed

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
The system SHALL allow configuring the commit convention in `.release.yaml`. Supported values SHALL include `conventional`, `angular`, `freeform`, and `custom`. A `custom` option SHALL allow defining arbitrary type-to-bump mappings.

#### Scenario: Custom commit convention
- **WHEN** the config specifies `commits.convention: custom` with `minor: ["feature", "enhancement"]`
- **AND** a commit message starts with `feature: add dark mode`
- **THEN** the determined bump type is `minor`

#### Scenario: Angular convention
- **WHEN** the config specifies `commits.convention: angular`
- **AND** commits follow Angular commit format
- **THEN** the system parses them according to Angular conventions

#### Scenario: Freeform convention
- **WHEN** the config specifies `commits.convention: freeform`
- **AND** commits use plain English messages
- **THEN** every commit is treated as a releasable patch-level change

### Requirement: Non-conforming commits are ignored for bump calculation
Commits that do not match the configured convention SHALL be ignored when determining the bump type but MAY be included in the changelog under an "Other" category.

#### Scenario: Non-conventional commit
- **WHEN** a commit message is `updated readme`
- **THEN** it does not contribute to the bump type calculation
- **AND** it MAY appear in the changelog under "Other"

### Requirement: No releasable commits detected
When no commits since the last release match the configured convention for any bump type, the system SHALL report that no release is needed and exit without error.

#### Scenario: No releasable changes
- **WHEN** all commits since the last release are non-conventional (e.g., `update docs`, `cleanup`)
- **THEN** the system reports "No releasable changes found" and exits with code 0
