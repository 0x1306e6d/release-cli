## MODIFIED Requirements

### Requirement: Fixed pipeline step order
The release pipeline SHALL execute steps in the following fixed order: detect → analyze commits → apply bump override (if provided) → bump manifest → propagate → generate changelog → commit → tag → publish → notify. The user SHALL NOT be able to reorder steps.

#### Scenario: Full pipeline execution
- **WHEN** `release-cli release` is run
- **THEN** steps execute in order: detect, analyze, bump override, bump manifest, propagate, changelog, commit, tag, publish, notify

#### Scenario: Pipeline with bump override
- **WHEN** `release-cli release --bump minor` is run
- **THEN** commits are analyzed for changelog content
- **AND** the bump override replaces the convention-derived bump
- **AND** the pipeline continues with the overridden bump level

## ADDED Requirements

### Requirement: Pipeline Options include bump override
The pipeline `Options` struct SHALL include a `BumpOverride *version.BumpType` field. When non-nil, it replaces the convention-derived bump after commit analysis.

#### Scenario: Bump override applied
- **WHEN** `Options.BumpOverride` is set to `BumpMinor`
- **AND** commits.Analyze returns `BumpPatch`
- **THEN** the pipeline uses `BumpMinor` for the release

#### Scenario: No bump override
- **WHEN** `Options.BumpOverride` is nil
- **THEN** the pipeline uses the convention-derived bump type as before
