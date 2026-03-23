## MODIFIED Requirements

### Requirement: Configurable commit convention
The system SHALL allow configuring the commit convention in `.release.yaml` under the `changes.commits` key. Supported values SHALL include `conventional`, `angular`, `freeform`, and `custom`. A `custom` option SHALL allow defining arbitrary type-to-bump mappings. When `changes.commits` is absent, the system SHALL use freeform behavior internally (accept all commits, patch bump).

#### Scenario: Custom commit convention
- **WHEN** the config specifies `changes.commits.convention: custom` with `minor: ["feature", "enhancement"]`
- **AND** a commit message starts with `feature: add dark mode`
- **THEN** the determined bump type is `minor`

#### Scenario: Angular convention
- **WHEN** the config specifies `changes.commits.convention: angular`
- **AND** commits follow Angular commit format
- **THEN** the system parses them according to Angular conventions

#### Scenario: Freeform convention
- **WHEN** the config specifies `changes.commits.convention: freeform`
- **AND** commits use plain English messages
- **THEN** every commit is treated as a releasable patch-level change

#### Scenario: No changes section defaults to freeform behavior
- **WHEN** the config does not include a `changes` section
- **THEN** the system SHALL accept all commits as releasable with patch bump (freeform behavior)

### Requirement: No releasable commits detected
When no commits since the last release match the configured convention for any bump type, the system SHALL report that no release is needed and exit without error. When `--bump` is provided, this check SHALL be skipped (the user explicitly requested a release).

#### Scenario: No releasable changes
- **WHEN** all commits since the last release are non-conventional (e.g., `update docs`, `cleanup`)
- **AND** no `--bump` flag is provided
- **THEN** the system reports "No releasable changes found" and exits with code 0

#### Scenario: No releasable changes but bump override provided
- **WHEN** all commits since the last release are non-conventional
- **AND** `--bump patch` is provided
- **THEN** the system proceeds with the release using patch bump
