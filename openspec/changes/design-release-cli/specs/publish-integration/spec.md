## ADDED Requirements

### Requirement: GitHub Releases as default publish target
The system SHALL support publishing to GitHub Releases as the default publish target for projects hosted on GitHub.

#### Scenario: Publish GitHub Release
- **WHEN** `publish.github.enabled: true` is configured
- **AND** the release tag `v1.4.0` is created
- **THEN** a GitHub Release is created with the tag `v1.4.0` and the changelog entry as the release body

### Requirement: GitHub Release artifact upload
The system SHALL support uploading artifacts to the GitHub Release using glob patterns specified in `publish.github.artifacts`.

#### Scenario: Upload artifacts
- **WHEN** `publish.github.artifacts` includes `dist/*.tar.gz`
- **AND** `dist/` contains `release-cli-linux-amd64.tar.gz` and `release-cli-darwin-arm64.tar.gz`
- **THEN** both files are uploaded as assets to the GitHub Release

#### Scenario: No artifacts configured
- **WHEN** `publish.github.artifacts` is not set
- **THEN** the GitHub Release is created without file attachments

### Requirement: GitHub Release draft mode
The system SHALL support creating GitHub Releases as drafts via `publish.github.draft: true`.

#### Scenario: Draft release
- **WHEN** `publish.github.draft: true` is configured
- **THEN** the GitHub Release is created as a draft (not publicly visible until manually published)

### Requirement: Publish target auto-detection
The detector MAY declare default publish targets for its ecosystem (e.g., `npm` for Node, `pypi` for Python). These defaults are included in the generated config during `release-cli init`.

#### Scenario: Node project default publish target
- **WHEN** `release-cli init` detects a Node project
- **THEN** the generated config includes `publish.npm.enabled: true`

### Requirement: Publish can be disabled
The user SHALL be able to disable any publish target by setting `enabled: false`.

#### Scenario: Disable GitHub publish
- **WHEN** the config specifies `publish.github.enabled: false`
- **THEN** no GitHub Release is created during the pipeline
