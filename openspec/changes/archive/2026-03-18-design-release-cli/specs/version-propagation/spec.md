## ADDED Requirements

### Requirement: Propagate version to secondary files
The system SHALL update all files listed in the `propagate` config section with the new version after bumping the primary manifest.

#### Scenario: Propagate to multiple targets
- **WHEN** the config lists two propagation targets
- **AND** a release bumps the version to `1.4.0`
- **THEN** both target files are updated with the new version

### Requirement: Structured field propagation
The system SHALL support updating version values in structured files (YAML, JSON, TOML) by specifying a field path.

#### Scenario: Update YAML field
- **WHEN** a propagation target specifies `file: helm/Chart.yaml` and `field: appVersion`
- **AND** the new version is `1.4.0`
- **THEN** the `appVersion` field in `helm/Chart.yaml` is set to `1.4.0`

#### Scenario: Update JSON field
- **WHEN** a propagation target specifies `file: config.json` and `field: version`
- **AND** the new version is `1.4.0`
- **THEN** the `version` field in `config.json` is set to `1.4.0`

### Requirement: Pattern-based propagation
The system SHALL support updating version values in arbitrary files using a pattern template with a `{{.Version}}` placeholder.

#### Scenario: Update Go version constant
- **WHEN** a propagation target specifies `file: internal/version/version.go` and `pattern: 'const Version = "{{.Version}}"'`
- **AND** the current file contains `const Version = "1.3.0"`
- **AND** the new version is `1.4.0`
- **THEN** the file is updated to contain `const Version = "1.4.0"`

#### Scenario: Pattern not found in file
- **WHEN** a propagation target specifies a pattern
- **AND** the pattern (with any version) does not match any line in the target file
- **THEN** the system SHALL report an error indicating the pattern was not found

### Requirement: Built-in propagation types
The system SHALL provide named propagation types for common patterns (e.g., `docker-label`) that expand to predefined patterns.

#### Scenario: Docker label propagation type
- **WHEN** a propagation target specifies `file: Dockerfile` and `type: docker-label`
- **AND** the Dockerfile contains `LABEL version="1.3.0"`
- **AND** the new version is `1.4.0`
- **THEN** the line is updated to `LABEL version="1.4.0"`

### Requirement: Propagation runs after manifest bump
Propagation SHALL execute after the primary manifest has been bumped and before the release commit is created, so that all version changes are included in a single commit.

#### Scenario: All version changes in one commit
- **WHEN** a release bumps the version
- **THEN** the manifest bump and all propagation changes are included in the same git commit
