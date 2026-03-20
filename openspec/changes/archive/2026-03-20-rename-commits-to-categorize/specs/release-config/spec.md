## MODIFIED Requirements

### Requirement: Config sections
The config SHALL support the following top-level sections: `project`, `version`, `categorize`, `changelog`, `propagate`, `hooks`, `publish`, `notify`.

#### Scenario: All sections present
- **WHEN** the config includes all sections with valid values
- **THEN** the system parses and applies all configured values

#### Scenario: Unknown section
- **WHEN** the config includes an unrecognized top-level key
- **THEN** the system SHALL report a warning about the unknown key

### Requirement: Init command generates config
The `release-cli init` command SHALL detect the project type and generate a `.release.yaml` with all sections — required fields filled in and optional fields commented out with descriptions. The commit categorization section SHALL use the `categorize` key.

#### Scenario: Init for a Gradle project
- **WHEN** `release-cli init` is run in a directory with `build.gradle`
- **THEN** a `.release.yaml` is generated with `project: java-gradle`, version scheme, `categorize` section for commit convention, changelog, and publish sections (with optional sections commented out)

#### Scenario: Init does not overwrite existing config
- **WHEN** `release-cli init` is run and `.release.yaml` already exists
- **THEN** the system SHALL report that config already exists and suggest editing it manually
