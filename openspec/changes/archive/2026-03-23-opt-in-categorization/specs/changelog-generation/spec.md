## MODIFIED Requirements

### Requirement: Generate changelog from parsed commits
The system SHALL generate a changelog entry for each release. When a change source with categorization is configured, commits SHALL be grouped by type (e.g., Features, Bug Fixes, Breaking Changes). When no change source is configured, the changelog SHALL render as a flat list without group headings.

#### Scenario: Changelog with mixed commit types (categorized)
- **WHEN** `changes.commits` is configured with a convention
- **AND** commits include `feat: add export`, `fix: null pointer`, and `feat!: new API`
- **THEN** the changelog groups them under "Breaking Changes", "Features", and "Bug Fixes" sections

#### Scenario: Flat changelog without changes configured
- **WHEN** no `changes` section is configured
- **AND** commits include "Add export feature", "Fix null pointer"
- **THEN** the changelog renders as a flat list without any `###` group headings

#### Scenario: Freeform convention renders flat
- **WHEN** `changes.commits.convention: freeform` is configured
- **AND** commits include "Add export feature", "Fix null pointer"
- **THEN** the changelog renders as a flat list without any `###` group headings

## ADDED Requirements

### Requirement: Entry tracks grouping mode
The changelog `Entry` struct SHALL include a `Grouped` boolean field. When `Grouped` is `true`, `Render()` produces group headings. When `Grouped` is `false`, `Render()` produces a flat list.

#### Scenario: Grouped entry rendering
- **WHEN** an Entry has `Grouped: true` and groups "Features" and "Bug Fixes"
- **THEN** `Render()` outputs `### Features` and `### Bug Fixes` headings

#### Scenario: Ungrouped entry rendering
- **WHEN** an Entry has `Grouped: false`
- **THEN** `Render()` outputs all items as a flat list without any `###` headings
