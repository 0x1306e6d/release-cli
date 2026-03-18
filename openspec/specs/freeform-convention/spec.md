## Requirements

### Requirement: Freeform convention treats all commits as releasable
The system SHALL support a `freeform` commit convention that accepts every commit as a releasable change, regardless of message format.

#### Scenario: Plain English commit is releasable
- **WHEN** the config specifies `commits.convention: freeform`
- **AND** a commit message is `Add user export feature`
- **THEN** the commit is parsed as a releasable change with bump type `patch`

#### Scenario: Single-word commit is releasable
- **WHEN** the config specifies `commits.convention: freeform`
- **AND** a commit message is `cleanup`
- **THEN** the commit is parsed as a releasable change with bump type `patch`

### Requirement: Freeform convention always produces patch bumps
The system SHALL assign `patch` as the bump type for every commit parsed under the `freeform` convention. No commit SHALL produce a `minor` or `major` bump.

#### Scenario: Multiple freeform commits result in patch bump
- **WHEN** the config specifies `commits.convention: freeform`
- **AND** commits since last release include `Add new feature`, `Fix login bug`, `Update docs`
- **THEN** the determined bump type is `patch`

### Requirement: Freeform convention uses full subject as description
The system SHALL use the entire commit subject line as the parsed commit's description and SHALL set the commit type to `other`.

#### Scenario: Full subject preserved in parsed commit
- **WHEN** the config specifies `commits.convention: freeform`
- **AND** a commit message subject is `Refactor authentication middleware for clarity`
- **THEN** the parsed commit has type `other` and subject `Refactor authentication middleware for clarity`

### Requirement: Freeform convention does not detect breaking changes
The system SHALL NOT detect breaking changes from freeform commits. The `Breaking` field SHALL always be `false`.

#### Scenario: Message containing BREAKING CHANGE is not treated as breaking
- **WHEN** the config specifies `commits.convention: freeform`
- **AND** a commit body contains `BREAKING CHANGE: removed old API`
- **THEN** the parsed commit has `Breaking: false` and bump type `patch`
