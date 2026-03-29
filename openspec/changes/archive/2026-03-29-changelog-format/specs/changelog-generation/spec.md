## MODIFIED Requirements

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

### Requirement: Generate changelog from parsed commits
The `Generate` function SHALL accept `[]ParsedCommit` (which now includes `Hash`) and produce an `Entry` with `Groups` of type `map[string][]Item`. Each `Item` SHALL contain `Hash` (abbreviated to 7 chars), `Title`, and `References` (resolved via GitHub API or regex fallback).

#### Scenario: Changelog with mixed commit types (categorized)
- **WHEN** `changes.commits` is configured with a convention
- **AND** commits include `feat: add export`, `fix: null pointer`, and `feat!: new API`
- **THEN** the changelog groups them under "Breaking Changes", "Features", and "Bug Fixes" sections, each item in the `<hash> - <title> (<refs>)` format

#### Scenario: Flat changelog without changes configured
- **WHEN** no `changes` section is configured
- **AND** commits include "Add export feature", "Fix null pointer"
- **THEN** the changelog renders as a flat list in the new item format without any `###` group headings

### Requirement: Headless body rendering
The `Entry` struct SHALL provide a `RenderBody()` method that renders the changelog content without the version heading line. Items SHALL use the new `<hash> - <title> (<refs>)` format.

#### Scenario: RenderBody with grouped entry
- **WHEN** an Entry has `Grouped: true` and groups "Features" and "Bug Fixes"
- **THEN** `RenderBody()` outputs `### Features` and `### Bug Fixes` headings with items in the new format, without a leading `## {Version} ({Date})` line

#### Scenario: RenderBody with ungrouped entry
- **WHEN** an Entry has `Grouped: false` and items with hashes and titles
- **THEN** `RenderBody()` outputs a flat list of items in the new format without a leading `## {Version} ({Date})` line

### Requirement: Customizable changelog template
The system SHALL support a custom Go template for changelog generation. The template SHALL receive the `Entry` struct with `Groups` of type `map[string][]Item`. Each `Item` exposes `.Hash`, `.Title`, and `.References` ([]string) fields.

#### Scenario: Custom template accessing item fields
- **WHEN** a custom template iterates over `{{ range $items }}` in a group
- **THEN** each item provides `{{ .Hash }}`, `{{ .Title }}`, and `{{ range .References }}`
