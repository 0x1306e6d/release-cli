## MODIFIED Requirements

### Requirement: Config sections
The config SHALL support the following top-level sections: `project`, `version`, `changes`, `changelog`, `propagate`, `hooks`, `publish`, `notify`. The `changes` section SHALL contain source-specific sub-sections (e.g., `changes.commits`).

#### Scenario: All sections present
- **WHEN** the config includes all sections with valid values
- **THEN** the system parses and applies all configured values

#### Scenario: Unknown section
- **WHEN** the config includes an unrecognized top-level key (including a top-level `categorize`)
- **THEN** the system SHALL report a warning about the unknown key

### Requirement: Only project field is required
The `project` field SHALL be the only required field in `.release.yaml`. All other fields SHALL have sensible defaults. When `changes` is absent, the system SHALL accept all commits without categorization.

#### Scenario: Minimal config
- **WHEN** the config contains only `project: go`
- **THEN** the system operates with default values: semver, no categorization (all commits accepted, flat changelog, patch bump), changelog enabled, GitHub publish enabled

#### Scenario: Missing project field
- **WHEN** the config file exists but has no `project` field
- **THEN** the system reports a validation error indicating `project` is required

### Requirement: Init command generates config
The `release-cli init` command SHALL detect the project type and generate a `.release.yaml` with all sections — required fields filled in and optional fields commented out with descriptions. The `changes.commits` section SHALL be included as a commented example.

#### Scenario: Init for a Gradle project
- **WHEN** `release-cli init` is run in a directory with `build.gradle`
- **THEN** a `.release.yaml` is generated with `project: java-gradle`, version scheme, `changelog` section, and `changes.commits` as a commented option

#### Scenario: Init does not overwrite existing config
- **WHEN** `release-cli init` is run and `.release.yaml` already exists
- **THEN** the system SHALL report that config already exists and suggest editing it manually

### Requirement: Configurable commit convention is under changes
The commit convention SHALL be configured under `changes.commits.convention`. Supported values remain `conventional`, `angular`, `freeform`, and `custom`.

#### Scenario: Convention under changes
- **WHEN** the config specifies `changes.commits.convention: angular`
- **THEN** the system uses the Angular convention for commit parsing and changelog grouping

#### Scenario: No changes section
- **WHEN** the config does not include a `changes` section
- **THEN** the system SHALL accept all commits, render a flat changelog, and default to patch bump

## REMOVED Requirements

### Requirement: Config sections
**Reason**: The top-level `categorize` key is replaced by `changes.commits`. The set of valid top-level sections no longer includes `categorize`.
**Migration**: Move `categorize:` content under `changes.commits:` in `.release.yaml`.
