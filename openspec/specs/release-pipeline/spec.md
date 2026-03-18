## ADDED Requirements

### Requirement: Fixed pipeline step order
The release pipeline SHALL execute steps in the following fixed order: detect → analyze commits → bump manifest → propagate → generate changelog → commit → tag → publish → notify. The user SHALL NOT be able to reorder steps.

#### Scenario: Full pipeline execution
- **WHEN** `release-cli release` is run
- **THEN** steps execute in order: detect, analyze, bump, propagate, changelog, commit, tag, publish, notify

### Requirement: Steps can be skipped based on config
Individual pipeline steps SHALL be skipped when their corresponding feature is disabled (e.g., changelog disabled, no publish targets, no notify targets).

#### Scenario: Skip changelog step
- **WHEN** `changelog.enabled: false` is set in config
- **THEN** the changelog step is skipped and the pipeline continues with commit

#### Scenario: Skip publish step
- **WHEN** no publish targets are enabled
- **THEN** the publish step is skipped and the pipeline continues with notify

### Requirement: Lifecycle hooks
The system SHALL support lifecycle hooks at the following points: `pre-bump`, `post-bump`, `pre-publish`, `post-publish`. Each hook SHALL execute one or more shell commands.

#### Scenario: Pre-bump hook runs before version bump
- **WHEN** the config specifies `hooks.pre-bump: ["make validate"]`
- **THEN** `make validate` is executed before the manifest version is updated

#### Scenario: Hook failure aborts the pipeline
- **WHEN** a lifecycle hook command exits with a non-zero status
- **THEN** the pipeline is aborted and an error is reported with the failed command and its output

### Requirement: Hook commands have access to version variables
Hook commands SHALL have access to release context via environment variables: `RELEASE_VERSION` (new version), `RELEASE_PREV_VERSION` (current version), `RELEASE_PROJECT` (project type).

#### Scenario: Hook uses version variable
- **WHEN** a hook command is `echo "Releasing $RELEASE_VERSION"`
- **AND** the new version is `1.4.0`
- **THEN** the command outputs `Releasing 1.4.0`

### Requirement: Dry-run mode
The system SHALL support a `--dry-run` flag that shows what each step would do without making any changes to files, git, or external services.

#### Scenario: Dry-run output
- **WHEN** `release-cli release --dry-run` is run
- **THEN** the system prints a preview of all steps (version bump, files changed, tag name, publish targets) without executing any of them

### Requirement: SNAPSHOT post-release step
When snapshot mode is enabled, the pipeline SHALL include an additional step after tagging that bumps the manifest to the next development version and creates a "prepare next development iteration" commit.

#### Scenario: Snapshot step after release
- **WHEN** snapshot mode is enabled
- **AND** the release version is `1.4.0`
- **THEN** after the release tag is created, the manifest is bumped to `1.5.0-SNAPSHOT` and a commit "Prepare next development iteration" is created

### Requirement: Pipeline reports progress
The system SHALL report progress for each pipeline step as it executes, showing what action was taken and the result.

#### Scenario: Progress output
- **WHEN** a release is executed
- **THEN** each step prints a status line (e.g., `✓ Bumped version: 1.3.0 → 1.4.0`, `✓ Updated CHANGELOG.md`, `✓ Tagged v1.4.0`)
