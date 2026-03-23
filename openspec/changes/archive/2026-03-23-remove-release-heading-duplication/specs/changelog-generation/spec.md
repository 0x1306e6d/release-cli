## ADDED Requirements

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
