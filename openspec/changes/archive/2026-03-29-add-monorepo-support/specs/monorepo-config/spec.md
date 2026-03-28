## ADDED Requirements

### Requirement: Multiple config files with parent-child declaration
The system SHALL support multiple `.release.yaml` files, one per package directory. A parent package declares its children via a `modules` list of relative directory paths. Each child directory MUST contain its own `.release.yaml`.

#### Scenario: Root with two children
- **WHEN** the root `.release.yaml` contains:
  ```yaml
  name: release-cli
  project: go
  modules:
    - cli
    - workflow
  ```
- **AND** `cli/.release.yaml` and `workflow/.release.yaml` exist
- **THEN** the system recognizes a tree: root `release-cli` with children `cli` and `workflow`

#### Scenario: Nested hierarchy
- **WHEN** `workflow/.release.yaml` contains `modules: [sub]`
- **AND** `workflow/sub/.release.yaml` exists
- **THEN** the system recognizes `sub` as a child of `workflow`, three levels deep

#### Scenario: Child directory missing config
- **WHEN** the root declares `modules: [cli]` but `cli/.release.yaml` does not exist
- **THEN** the system SHALL report a validation error: `package "cli": missing .release.yaml in "cli/"`

#### Scenario: Child directory does not exist
- **WHEN** the root declares `modules: [nonexistent]` and the directory does not exist
- **THEN** the system SHALL report a validation error: `package "nonexistent": directory does not exist`

### Requirement: No config inheritance
Each `.release.yaml` SHALL be fully independent. A child package does NOT inherit any config sections from its parent. Each config file is parsed and validated in isolation.

#### Scenario: Parent and child with different conventions
- **WHEN** the root uses `changes.commits.convention: conventional`
- **AND** `cli/.release.yaml` uses `changes.commits.convention: angular`
- **THEN** each package uses its own convention independently

#### Scenario: Child with minimal config
- **WHEN** `cli/.release.yaml` contains only `project: go`
- **THEN** the `cli` package uses default values for all other sections, not the parent's values

### Requirement: name field for root tag prefix
A config with a `modules` list SHALL require a `name` field. The `name` field determines the git tag prefix for that package (e.g., `name: release-cli` produces tags like `release-cli/v1.2.0`). Configs without `modules` SHALL NOT require `name`.

#### Scenario: Root with name
- **WHEN** the root config has `name: release-cli` and `modules: [cli]`
- **THEN** validation passes and the root's tag prefix is `release-cli`

#### Scenario: Parent missing name
- **WHEN** a config has `modules: [cli]` but no `name` field
- **THEN** the system SHALL report a validation error: `"name" is required when "modules" is declared`

#### Scenario: Leaf package without name
- **WHEN** `cli/.release.yaml` has `project: go` and no `name` or `modules`
- **THEN** validation passes — `name` is not required for leaf packages

### Requirement: Config tree validation
The system SHALL validate the full package tree after loading all configs: no circular references, no overlapping paths between siblings, and all declared children exist with valid configs.

#### Scenario: Circular reference
- **WHEN** root declares `modules: [cli]` and `cli/.release.yaml` declares `modules: [..]` pointing back to root
- **THEN** the system SHALL report a validation error about circular package references

#### Scenario: Overlapping sibling paths
- **WHEN** root declares `modules: [a, a/b]` where `a/b` is nested inside `a`
- **THEN** the system SHALL report a validation error about overlapping package paths

### Requirement: Single-project backward compatibility
A `.release.yaml` without a `modules` field SHALL behave exactly as today — single-project mode. The `name` field is ignored if present without `modules`.

#### Scenario: Existing single-project config
- **WHEN** the config contains only `project: go`
- **THEN** the system operates in single-project mode with `v<version>` tags, identical to current behavior

#### Scenario: Single-project with name field
- **WHEN** the config contains `name: foo` and `project: go` but no `modules`
- **THEN** the system operates in single-project mode — `name` is ignored
