## MODIFIED Requirements

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
