## ADDED Requirements

### Requirement: Generate changelog from parsed commits
The system SHALL generate a changelog entry for each release, grouping commits by type (e.g., Features, Bug Fixes, Breaking Changes).

#### Scenario: Changelog with mixed commit types
- **WHEN** commits include `feat: add export`, `fix: null pointer`, and `feat!: new API`
- **THEN** the changelog groups them under "Breaking Changes", "Features", and "Bug Fixes" sections

### Requirement: Write changelog to file
The system SHALL prepend the new release entry to the configured changelog file (default: `CHANGELOG.md`). Existing entries SHALL be preserved.

#### Scenario: Prepend to existing changelog
- **WHEN** `CHANGELOG.md` exists with previous entries
- **AND** a new release `1.4.0` is created
- **THEN** the `1.4.0` entry is added at the top, above existing entries

#### Scenario: Create changelog if it does not exist
- **WHEN** `CHANGELOG.md` does not exist
- **AND** a release is created
- **THEN** `CHANGELOG.md` is created with the release entry

### Requirement: Changelog entry format
Each changelog entry SHALL include the version number, release date, and grouped commits. Each commit entry SHALL include the commit message summary and a short hash linking to the commit.

#### Scenario: Changelog entry structure
- **WHEN** version `1.4.0` is released on `2026-03-18`
- **THEN** the changelog entry starts with `## [1.4.0] - 2026-03-18` followed by grouped commit summaries

### Requirement: Changelog can be disabled
The system SHALL allow disabling changelog generation via `changelog.enabled: false` in the config.

#### Scenario: Changelog disabled
- **WHEN** the config specifies `changelog.enabled: false`
- **THEN** no changelog file is created or modified during release

### Requirement: Customizable changelog template
The system SHALL support a custom Go template for changelog generation, allowing users to control the output format.

#### Scenario: Custom template
- **WHEN** the config specifies `changelog.template: .release-changelog.tmpl`
- **THEN** the system uses that template file to render the changelog entry instead of the default format
