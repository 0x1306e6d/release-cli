## ADDED Requirements

### Requirement: Config file format is YAML
The system SHALL read configuration from `.release.yaml` in the project root directory.

#### Scenario: Config file found
- **WHEN** `.release.yaml` exists in the project root
- **THEN** the system loads and parses it as YAML

#### Scenario: Config file not found
- **WHEN** `.release.yaml` does not exist
- **THEN** the system SHALL report an error and suggest running `release-cli init`

### Requirement: Only project field is required
The `project` field SHALL be the only required field in `.release.yaml`. All other fields SHALL have sensible defaults. When `changes` is absent, the system SHALL accept all commits without categorization.

#### Scenario: Minimal config
- **WHEN** the config contains only `project: go`
- **THEN** the system operates with default values: semver, no categorization (all commits accepted, flat changelog, patch bump), changelog enabled, GitHub publish enabled

#### Scenario: Missing project field
- **WHEN** the config file exists but has no `project` field
- **THEN** the system reports a validation error indicating `project` is required

### Requirement: Config sections
The config SHALL support the following top-level sections: `project`, `version`, `changes`, `changelog`, `propagate`, `hooks`, `publish`, `notify`. The `changes` section SHALL contain source-specific sub-sections (e.g., `changes.commits`).

#### Scenario: All sections present
- **WHEN** the config includes all sections with valid values
- **THEN** the system parses and applies all configured values

#### Scenario: Unknown section
- **WHEN** the config includes an unrecognized top-level key (including a top-level `categorize`)
- **THEN** the system SHALL report a warning about the unknown key

### Requirement: Environment variable references in config
The system SHALL support `${ENV_VAR}` syntax in string values to reference environment variables.

#### Scenario: Resolve environment variable
- **WHEN** the config contains `webhook: ${SLACK_WEBHOOK_URL}`
- **AND** the environment variable `SLACK_WEBHOOK_URL` is set to `https://hooks.slack.com/xxx`
- **THEN** the value resolves to `https://hooks.slack.com/xxx`

#### Scenario: Missing environment variable
- **WHEN** the config references `${UNDEFINED_VAR}` and the variable is not set
- **THEN** the system SHALL report an error indicating the missing variable

### Requirement: Config validation
The system SHALL validate the config after parsing, checking for: valid project identifier, valid versioning scheme, valid commit convention, valid propagation targets, and valid publish/notify configurations.

#### Scenario: Invalid project identifier
- **WHEN** the config specifies `project: unknown-lang`
- **THEN** the system reports an error listing the valid project identifiers

### Requirement: Init command generates config
The `release-cli init` command SHALL detect the project type and generate a `.release.yaml` with all sections — required fields filled in and optional fields commented out with descriptions. The `changes.commits` section SHALL be included as a commented example.

#### Scenario: Init for a Gradle project
- **WHEN** `release-cli init` is run in a directory with `build.gradle`
- **THEN** a `.release.yaml` is generated with `project: java-gradle`, version scheme, `changelog` section, and `changes.commits` as a commented option

#### Scenario: Init does not overwrite existing config
- **WHEN** `release-cli init` is run and `.release.yaml` already exists
- **THEN** the system SHALL report that config already exists and suggest editing it manually

### Requirement: Init auto-detects snapshot mode
The `release-cli init` command SHALL detect if the current manifest version contains a pre-release suffix (e.g., `-SNAPSHOT`) and enable snapshot mode automatically in the generated config.

#### Scenario: Snapshot detected during init
- **WHEN** `release-cli init` is run and `gradle.properties` contains `version=1.4.0-SNAPSHOT`
- **THEN** the generated config includes `version.snapshot: true`

### Requirement: Configuration documentation includes Go project example
The release-config spec SHALL reference the project's own `.release.yaml` as the canonical Go project example.

#### Scenario: Go project example exists
- **WHEN** a user reads the project documentation for Go project configuration
- **THEN** they SHALL be able to reference `.release.yaml` in the repository root as a working example

#### Scenario: Example covers all pipeline stages
- **WHEN** the `.release.yaml` example is reviewed
- **THEN** it SHALL demonstrate project detection, commit convention, changelog, and publish configuration

### Requirement: Configurable commit convention is under changes
The commit convention SHALL be configured under `changes.commits.convention`. Supported values remain `conventional`, `angular`, `freeform`, and `custom`.

#### Scenario: Convention under changes
- **WHEN** the config specifies `changes.commits.convention: angular`
- **THEN** the system uses the Angular convention for commit parsing and changelog grouping

#### Scenario: No changes section
- **WHEN** the config does not include a `changes` section
- **THEN** the system SHALL accept all commits, render a flat changelog, and default to patch bump
