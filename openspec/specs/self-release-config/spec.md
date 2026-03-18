## ADDED Requirements

### Requirement: Project ships with a release configuration
The repository SHALL include a `.release.yaml` at the project root that configures release-cli to release itself.

#### Scenario: Config declares Go project type
- **WHEN** `.release.yaml` is loaded
- **THEN** `project` SHALL be set to `go`

#### Scenario: Config uses conventional commits
- **WHEN** `.release.yaml` is loaded
- **THEN** `commits.convention` SHALL be set to `conventional`

#### Scenario: Config enables changelog generation
- **WHEN** `.release.yaml` is loaded
- **THEN** `changelog.enabled` SHALL be `true` and `changelog.file` SHALL be `CHANGELOG.md`

#### Scenario: Config enables GitHub Releases publishing
- **WHEN** `.release.yaml` is loaded
- **THEN** `publish.github.enabled` SHALL be `true`

#### Scenario: Config attaches cross-platform binaries
- **WHEN** `.release.yaml` is loaded
- **THEN** `publish.github.artifacts` SHALL include glob patterns matching all platform tarballs produced by the build (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64)

### Requirement: Version scheme is semver from git tags
The configuration SHALL use `version.scheme: semver` and rely on git tags as the version source (standard Go project behavior).

#### Scenario: No manifest version override
- **WHEN** `release-cli status` is run in the project root
- **THEN** the current version SHALL be read from the latest `v*` git tag
