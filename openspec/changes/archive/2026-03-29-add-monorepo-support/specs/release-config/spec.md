## MODIFIED Requirements

### Requirement: Only project field is required
The config SHALL require a `project` field. Optionally, it MAY include a `modules` list (making it a parent in a monorepo) and a `name` field (required when `modules` is present). All other fields SHALL have sensible defaults. When `changes` is absent, the system SHALL accept all commits without categorization.

#### Scenario: Minimal single-project config
- **WHEN** the config contains only `project: go`
- **THEN** the system operates in single-project mode with default values: semver, no categorization (all commits accepted, flat changelog, patch bump), changelog enabled, GitHub publish enabled

#### Scenario: Minimal monorepo parent config
- **WHEN** the config contains `name: my-project`, `project: go`, and `modules: [cli, lib]`
- **THEN** the system operates in monorepo mode, loading child configs from `cli/.release.yaml` and `lib/.release.yaml`

#### Scenario: Missing project field
- **WHEN** the config file exists but has no `project` field
- **THEN** the system reports a validation error indicating `project` is required

#### Scenario: Modules without name
- **WHEN** the config has `project: go` and `modules: [cli]` but no `name`
- **THEN** the system reports a validation error: `"name" is required when "modules" is declared`

### Requirement: Config sections
The config SHALL support the following top-level sections: `project`, `name`, `version`, `changes`, `changelog`, `propagate`, `hooks`, `publish`, `notify`, `modules`. The `changes` section SHALL contain source-specific sub-sections (e.g., `changes.commits`). The `modules` section SHALL be a list of relative directory paths.

#### Scenario: All sections present
- **WHEN** the config includes all sections with valid values
- **THEN** the system parses and applies all configured values

#### Scenario: Unknown section
- **WHEN** the config includes an unrecognized top-level key (including a top-level `categorize`)
- **THEN** the system SHALL report a warning about the unknown key

### Requirement: Config validation
The system SHALL validate the config after parsing, checking for: valid project identifier, valid versioning scheme, valid commit convention, valid propagation targets, valid publish/notify configurations. When `modules` is present, additionally validate: `name` is set, each child directory exists, each child has a valid `.release.yaml`, no circular references, no overlapping sibling paths.

#### Scenario: Invalid project identifier
- **WHEN** the config specifies `project: unknown-lang`
- **THEN** the system reports an error listing the valid project identifiers

#### Scenario: Child with invalid config
- **WHEN** root declares `modules: [cli]` and `cli/.release.yaml` has an invalid project identifier
- **THEN** the system reports an error: `package "cli": invalid project identifier "unknown-lang"`

#### Scenario: Overlapping sibling paths
- **WHEN** root declares `modules: [a, a/b]`
- **THEN** the system reports a validation error about overlapping package paths
