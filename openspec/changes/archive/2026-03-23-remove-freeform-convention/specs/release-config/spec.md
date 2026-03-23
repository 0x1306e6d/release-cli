## MODIFIED Requirements

### Requirement: Configurable commit convention is under changes
The commit convention SHALL be configured under `changes.commits.convention`. Supported values SHALL include `conventional`, `angular`, and `custom`. A `custom` option SHALL allow defining arbitrary type-to-bump mappings. When `changes.commits` is absent, the system SHALL accept all commits, render a flat changelog, and default to patch bump.

#### Scenario: Convention under changes
- **WHEN** the config specifies `changes.commits.convention: angular`
- **THEN** the system uses the Angular convention for commit parsing and changelog grouping

#### Scenario: No changes section
- **WHEN** the config does not include a `changes` section
- **THEN** the system SHALL accept all commits, render a flat changelog, and default to patch bump

#### Scenario: Freeform convention rejected
- **WHEN** the config specifies `changes.commits.convention: freeform`
- **THEN** the system SHALL report a validation error listing valid conventions (`conventional`, `angular`, `custom`)

### Requirement: Init command generates config
The `release-cli init` command SHALL detect the project type and generate a `.release.yaml` with all sections — required fields filled in and optional fields commented out with descriptions. The `changes.commits` section SHALL be included as a commented example listing `conventional`, `angular`, and `custom` as valid conventions.

#### Scenario: Init for a Gradle project
- **WHEN** `release-cli init` is run in a directory with `build.gradle`
- **THEN** a `.release.yaml` is generated with `project: java-gradle`, version scheme, `changelog` section, and `changes.commits` as a commented option listing `conventional`, `angular`, `custom`
