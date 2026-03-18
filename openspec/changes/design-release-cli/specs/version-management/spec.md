## ADDED Requirements

### Requirement: Read current version from manifest
The system SHALL read the current version from the manifest file declared by the detector. If the manifest file does not exist or has no version, the system SHALL report an error.

#### Scenario: Read version from package.json
- **WHEN** the project type is `node` and `package.json` contains `"version": "1.2.3"`
- **THEN** the current version is parsed as `1.2.3`

#### Scenario: Manifest file missing
- **WHEN** the detector expects `package.json` but the file does not exist
- **THEN** the system reports an error indicating the manifest file was not found

### Requirement: Read current version from git tags as fallback
For project types without a manifest file (e.g., Go), the system SHALL read the current version from the latest semver-compatible git tag.

#### Scenario: Read version from git tag
- **WHEN** the project type is `go` and the latest semver tag is `v1.2.3`
- **THEN** the current version is parsed as `1.2.3`

#### Scenario: No git tags exist
- **WHEN** the project type uses git tags and no semver tags exist
- **THEN** the system SHALL treat the current version as `0.0.0`

### Requirement: Calculate next version from bump type
The system SHALL calculate the next version by applying the bump type (major, minor, patch) to the current version according to the configured versioning scheme.

#### Scenario: Minor bump on semver
- **WHEN** the current version is `1.2.3` and the bump type is `minor`
- **THEN** the next version is `1.3.0`

#### Scenario: Major bump on semver
- **WHEN** the current version is `1.2.3` and the bump type is `major`
- **THEN** the next version is `2.0.0`

### Requirement: Bump version in manifest file
The system SHALL write the new version back to the manifest file, preserving the file's formatting and structure as much as possible.

#### Scenario: Bump version in package.json
- **WHEN** the project type is `node` and the next version is `1.3.0`
- **THEN** `package.json` is updated with `"version": "1.3.0"` and no other fields are modified

### Requirement: Support semver versioning scheme
The system SHALL support Semantic Versioning (major.minor.patch) with optional pre-release and build metadata suffixes.

#### Scenario: Parse semver with pre-release
- **WHEN** the version string is `1.4.0-SNAPSHOT`
- **THEN** it is parsed as major=1, minor=4, patch=0, pre-release=SNAPSHOT

### Requirement: Strip pre-release suffix on release
When releasing a version with a pre-release suffix (e.g., SNAPSHOT), the system SHALL strip the suffix and apply commit analysis to determine the final release version from the last release tag, not from the pre-release version.

#### Scenario: Release from SNAPSHOT version
- **WHEN** the manifest version is `1.4.0-SNAPSHOT`
- **AND** the last release tag is `v1.3.0`
- **AND** commit analysis determines a minor bump
- **THEN** the release version is `1.4.0`

#### Scenario: Release from SNAPSHOT with breaking change
- **WHEN** the manifest version is `1.4.0-SNAPSHOT`
- **AND** the last release tag is `v1.3.0`
- **AND** commit analysis determines a major bump
- **THEN** the release version is `2.0.0` (overriding the SNAPSHOT hint)

### Requirement: Post-release SNAPSHOT bump
When snapshot mode is enabled, the system SHALL create a second commit after the release that bumps the manifest to the next development version with the ecosystem-appropriate pre-release suffix.

#### Scenario: Post-release snapshot for Java
- **WHEN** snapshot mode is enabled and the release version is `1.4.0`
- **AND** the project type is `java-gradle`
- **THEN** after the release commit and tag, the manifest is bumped to `1.5.0-SNAPSHOT` in a separate commit

#### Scenario: Post-release snapshot for Python
- **WHEN** snapshot mode is enabled and the release version is `1.4.0`
- **AND** the project type is `python`
- **THEN** after the release commit and tag, the manifest is bumped to `1.5.0.dev0` in a separate commit

### Requirement: Git tag always created
The system SHALL create a git tag for every release, regardless of whether the version source is a manifest or git tags. The tag format SHALL be `v<version>` (e.g., `v1.4.0`).

#### Scenario: Tag created after manifest bump
- **WHEN** a release is executed for version `1.4.0`
- **THEN** a git tag `v1.4.0` is created on the release commit

### Requirement: Manifest override in config
The user SHALL be able to override the default manifest file path and field/pattern in `.release.yaml` for cases where the project structure is non-standard.

#### Scenario: Override manifest path
- **WHEN** the config specifies `version.manifest: custom/version.txt`
- **THEN** the system reads and writes the version from that file instead of the detector's default
