## MODIFIED Requirements

### Requirement: Configurable commit convention
The system SHALL allow configuring the commit convention in `.release.yaml` under the `categorize` key. Supported values SHALL include `conventional`, `angular`, `freeform`, and `custom`. A `custom` option SHALL allow defining arbitrary type-to-bump mappings.

#### Scenario: Custom commit convention
- **WHEN** the config specifies `categorize.convention: custom` with `minor: ["feature", "enhancement"]`
- **AND** a commit message starts with `feature: add dark mode`
- **THEN** the determined bump type is `minor`

#### Scenario: Angular convention
- **WHEN** the config specifies `categorize.convention: angular`
- **AND** commits follow Angular commit format
- **THEN** the system parses them according to Angular conventions

#### Scenario: Freeform convention
- **WHEN** the config specifies `categorize.convention: freeform`
- **AND** commits use plain English messages
- **THEN** every commit is treated as a releasable patch-level change

#### Scenario: Default convention when categorize section is absent
- **WHEN** the config does not include a `categorize` section
- **THEN** the system SHALL default to `conventional` commit convention
