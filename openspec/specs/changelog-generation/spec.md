## ADDED Requirements

### Requirement: Generate changelog from parsed commits
The `Generate` function SHALL accept `[]ParsedCommit` (which now includes `Hash`) and produce an `Entry` with `Groups` of type `map[string][]Item`. Each `Item` SHALL contain `Hash` (abbreviated to 7 chars), `Title`, and `References` (resolved via GitHub API or regex fallback). When a change source with categorization is configured, commits SHALL be grouped by type. When no change source is configured, the changelog SHALL render as a flat list without group headings.

#### Scenario: Changelog with mixed commit types (categorized)
- **WHEN** `changes.commits` is configured with a convention
- **AND** commits include `feat: add export`, `fix: null pointer`, and `feat!: new API`
- **THEN** the changelog groups them under "Breaking Changes", "Features", and "Bug Fixes" sections, each item in the `<hash> - <title> (<refs>)` format

#### Scenario: Flat changelog without changes configured
- **WHEN** no `changes` section is configured
- **AND** commits include "Add export feature", "Fix null pointer"
- **THEN** the changelog renders as a flat list in the new item format without any `###` group headings

#### Scenario: Freeform convention renders flat
- **WHEN** `changes.commits.convention: freeform` is configured
- **AND** commits include "Add export feature", "Fix null pointer"
- **THEN** the changelog renders as a flat list in the new item format without any `###` group headings

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
Each changelog entry SHALL include the version number, release date, and commit items. Each item SHALL render as `- <short-hash> - <title> (<references>)` where short-hash is the first 7 characters of the commit hash. When references are empty, the parenthetical SHALL be omitted.

#### Scenario: Entry with hash, title, and PR reference
- **WHEN** an item has hash `"a1b2c3d"`, title `"Add login validation"`, and references `["#42"]`
- **THEN** it renders as `- a1b2c3d - Add login validation (#42)`

#### Scenario: Entry with multiple references
- **WHEN** an item has hash `"b2c3d4e"`, title `"Fix crash"`, and references `["#15", "#42"]`
- **THEN** it renders as `- b2c3d4e - Fix crash (#15, #42)`

#### Scenario: Entry with no references
- **WHEN** an item has hash `"c3d4e5f"`, title `"Update README"`, and references `[]`
- **THEN** it renders as `- c3d4e5f - Update README`

#### Scenario: Entry with scope
- **WHEN** an item has hash `"e5f6a7b"`, title `"Add export"`, scope rendered as `**cli:** Add export`, and references `["#10"]`
- **THEN** it renders as `- e5f6a7b - **cli:** Add export (#10)`

#### Scenario: Changelog entry structure
- **WHEN** version `1.4.0` is released on `2026-03-18`
- **THEN** the changelog entry starts with `## 1.4.0 (2026-03-18)` followed by commit items in the new format

### Requirement: Changelog can be disabled
The system SHALL allow disabling changelog generation via `changelog.enabled: false` in the config.

#### Scenario: Changelog disabled
- **WHEN** the config specifies `changelog.enabled: false`
- **THEN** no changelog file is created or modified during release

### Requirement: Customizable changelog template
The system SHALL support a custom Go template for changelog generation. The template SHALL receive the `Entry` struct with `Groups` of type `map[string][]Item`. Each `Item` exposes `.Hash`, `.Title`, and `.References` ([]string) fields.

#### Scenario: Custom template
- **WHEN** the config specifies `changelog.template: .release-changelog.tmpl`
- **THEN** the system uses that template file to render the changelog entry instead of the default format

#### Scenario: Custom template accessing item fields
- **WHEN** a custom template iterates over `{{ range $items }}` in a group
- **THEN** each item provides `{{ .Hash }}`, `{{ .Title }}`, and `{{ range .References }}`

### Requirement: Headless body rendering
The `Entry` struct SHALL provide a `RenderBody()` method that renders the changelog content without the version heading line. Items SHALL use the new `<hash> - <title> (<refs>)` format.

#### Scenario: RenderBody with grouped entry
- **WHEN** an Entry has `Grouped: true` and groups "Features" and "Bug Fixes"
- **THEN** `RenderBody()` outputs `### Features` and `### Bug Fixes` headings with items in the new format, without a leading `## {Version} ({Date})` line

#### Scenario: RenderBody with ungrouped entry
- **WHEN** an Entry has `Grouped: false` and items with hashes and titles
- **THEN** `RenderBody()` outputs a flat list of items in the new format without a leading `## {Version} ({Date})` line

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
