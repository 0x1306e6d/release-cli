## ADDED Requirements

### Requirement: Detector registry maps project identifiers to detectors
The system SHALL maintain a registry of detectors, each identified by a unique name (e.g., `go`, `node`, `java-gradle`). The registry SHALL support lookup by exact name.

#### Scenario: Lookup by exact name
- **WHEN** the config specifies `project: java-gradle`
- **THEN** the registry returns the `java-gradle` detector directly without scanning the filesystem

### Requirement: General identifiers auto-resolve to specific detectors
When a general identifier (e.g., `java`) is provided, the system SHALL scan the project directory to determine which specific detector matches (e.g., `java-gradle` if `build.gradle` exists).

#### Scenario: General identifier resolves to specific detector
- **WHEN** the config specifies `project: java`
- **AND** the project directory contains `build.gradle`
- **THEN** the system resolves to the `java-gradle` detector

#### Scenario: General identifier with no matching build tool
- **WHEN** the config specifies `project: java`
- **AND** the project directory contains no recognized build tool files
- **THEN** the system SHALL report an error indicating the build tool could not be detected

#### Scenario: General identifier with multiple matching build tools
- **WHEN** the config specifies `project: java`
- **AND** the project directory contains both `build.gradle` and `pom.xml`
- **THEN** the system SHALL report an error listing the ambiguous matches and ask the user to specify a specific identifier

### Requirement: Each detector knows its ecosystem's manifest
Each detector SHALL declare the default manifest file path and how to read/write version information from it. Detectors for ecosystems without a manifest file (e.g., Go) SHALL declare that versions are read from git tags.

#### Scenario: Node detector reads version from package.json
- **WHEN** the `node` detector reads the version
- **THEN** it reads the `version` field from `package.json`

#### Scenario: Go detector reads version from git tags
- **WHEN** the `go` detector reads the version
- **AND** no manifest override is configured
- **THEN** it reads the latest semver git tag

#### Scenario: Java-Gradle detector reads version from gradle.properties
- **WHEN** the `java-gradle` detector reads the version
- **THEN** it reads the `version` property from `gradle.properties`

### Requirement: Detectors for v1 ecosystems
The system SHALL ship with built-in detectors for: Go, Node (npm/pnpm/bun), Python (poetry/setuptools), Rust, Java (Gradle/Maven), Dart, and Helm.

#### Scenario: All v1 detectors are registered
- **WHEN** the detector registry is initialized
- **THEN** detectors for `go`, `node`, `python`, `rust`, `java-gradle`, `java-maven`, `dart`, and `helm` are available

### Requirement: Detector interface supports future extension
The detector interface SHALL use option structs or context objects for methods that may need additional parameters in the future, to avoid breaking changes when adding capabilities.

#### Scenario: Adding a new method to detector
- **WHEN** a new capability is added to the detector interface
- **THEN** existing detectors that do not implement it SHALL continue to compile and function with default behavior
