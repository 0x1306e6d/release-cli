## MODIFIED Requirements

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
