## ADDED Requirements

### Requirement: CLI flag to override version bump level
The `release` subcommand SHALL accept a `--bump` flag with values `major`, `minor`, or `patch` that overrides the convention-derived bump level.

#### Scenario: Override bump to minor
- **WHEN** `release-cli release --bump minor` is run
- **AND** the convention would derive a patch bump
- **THEN** the system SHALL use minor as the bump level

#### Scenario: Override bump to major
- **WHEN** `release-cli release --bump major` is run
- **THEN** the system SHALL use major as the bump level regardless of convention analysis

#### Scenario: Invalid bump value
- **WHEN** `release-cli release --bump invalid` is run
- **THEN** the system SHALL report an error: invalid bump type, must be major, minor, or patch

### Requirement: Bump override works with all conventions
The `--bump` flag SHALL override the bump level regardless of which convention is configured (conventional, angular, freeform, custom, or none).

#### Scenario: Override with no changes configured
- **WHEN** no `changes` section exists
- **AND** `--bump minor` is provided
- **THEN** the system SHALL bump the minor version

#### Scenario: Override with conventional commits
- **WHEN** `changes.commits.convention: conventional` is configured
- **AND** all commits are `fix:` type (normally patch)
- **AND** `--bump minor` is provided
- **THEN** the system SHALL bump the minor version

### Requirement: Commits still analyzed for changelog when bump is overridden
When `--bump` is provided, the system SHALL still analyze commits for changelog content. Only the bump level is overridden.

#### Scenario: Changelog generated with override bump
- **WHEN** `--bump major` is provided
- **AND** there are `feat:` and `fix:` commits
- **THEN** the changelog includes all commit entries grouped normally
- **AND** the version bump is major

### Requirement: ParseBumpType utility function
The `version` package SHALL provide a `ParseBumpType(string) (BumpType, error)` function that converts string values to `BumpType` constants. The function SHALL be case-insensitive.

#### Scenario: Parse valid bump types
- **WHEN** `ParseBumpType("minor")` is called
- **THEN** it returns `BumpMinor` with no error

#### Scenario: Parse case-insensitive
- **WHEN** `ParseBumpType("MAJOR")` is called
- **THEN** it returns `BumpMajor` with no error

#### Scenario: Parse invalid value
- **WHEN** `ParseBumpType("invalid")` is called
- **THEN** it returns an error
