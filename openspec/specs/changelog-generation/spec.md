## ADDED Requirements

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

### Requirement: Headless body rendering
The `Entry` struct SHALL provide a `RenderBody()` method that renders the changelog content without the version heading line. The output SHALL start directly with the commit groups (or flat list), omitting the `## {Version} ({Date})` line.

#### Scenario: RenderBody with grouped entry
- **WHEN** an Entry has `Version: "1.4.0"`, `Date: "2026-03-18"`, `Grouped: true`, and groups "Features" and "Bug Fixes"
- **THEN** `RenderBody()` outputs `### Features` and `### Bug Fixes` headings with their items, without a leading `## 1.4.0 (2026-03-18)` line

#### Scenario: RenderBody with ungrouped entry
- **WHEN** an Entry has `Version: "1.4.0"`, `Date: "2026-03-18"`, `Grouped: false`, and items "Add export", "Fix null pointer"
- **THEN** `RenderBody()` outputs a flat list of items without a leading `## 1.4.0 (2026-03-18)` line

#### Scenario: Render still includes heading
- **WHEN** `Render()` is called on an Entry with `Version: "1.4.0"` and `Date: "2026-03-18"`
- **THEN** the output starts with `## 1.4.0 (2026-03-18)` (existing behavior unchanged)

### Requirement: Entry tracks grouping mode
The changelog `Entry` struct SHALL include a `Grouped` boolean field. When `Grouped` is `true`, `Render()` produces group headings. When `Grouped` is `false`, `Render()` produces a flat list.

#### Scenario: Grouped entry rendering
- **WHEN** an Entry has `Grouped: true` and groups "Features" and "Bug Fixes"
- **THEN** `Render()` outputs `### Features` and `### Bug Fixes` headings

#### Scenario: Ungrouped entry rendering
- **WHEN** an Entry has `Grouped: false`
- **THEN** `Render()` outputs all items as a flat list without any `###` headings
