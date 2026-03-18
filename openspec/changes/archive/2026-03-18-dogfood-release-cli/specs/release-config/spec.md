## MODIFIED Requirements

### Requirement: Configuration documentation includes Go project example
The release-config spec SHALL reference the project's own `.release.yaml` as the canonical Go project example.

#### Scenario: Go project example exists
- **WHEN** a user reads the project documentation for Go project configuration
- **THEN** they SHALL be able to reference `.release.yaml` in the repository root as a working example

#### Scenario: Example covers all pipeline stages
- **WHEN** the `.release.yaml` example is reviewed
- **THEN** it SHALL demonstrate project detection, commit convention, changelog, and publish configuration
