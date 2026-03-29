## ADDED Requirements

### Requirement: Carry commit hash through parsing
The `ParsedCommit` struct SHALL include a `Hash` field. The `Analyze` function SHALL copy the `RawCommit.Hash` into `ParsedCommit.Hash` for every parsed commit.

#### Scenario: Hash preserved after parsing
- **WHEN** a `RawCommit` with `Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"` is parsed
- **THEN** the resulting `ParsedCommit` SHALL have `Hash` equal to `"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"`

### Requirement: Resolve PR and issue references via GitHub API
The system SHALL query the GitHub API (`GET /repos/{owner}/{repo}/commits/{sha}/pulls`) to find associated PRs for each commit when `GITHUB_TOKEN` is available. The PR number SHALL be included in the references. Linked issue numbers from the PR body SHALL also be extracted.

#### Scenario: Commit associated with a PR
- **WHEN** `GITHUB_TOKEN` is set and commit `a1b2c3d` is associated with PR `#42`
- **THEN** the resolved references SHALL include `"#42"`

#### Scenario: PR with linked issues
- **WHEN** `GITHUB_TOKEN` is set and commit `a1b2c3d` is associated with PR `#42` which closes `#15`
- **THEN** the resolved references SHALL include `["#15", "#42"]`

#### Scenario: Commit with no associated PR
- **WHEN** `GITHUB_TOKEN` is set but commit `b2c3d4e` has no associated PR
- **THEN** the resolved references SHALL be an empty slice

#### Scenario: Concurrent lookups bounded
- **WHEN** there are 20 commits to look up
- **THEN** the system SHALL make API calls with bounded concurrency (no more than 5 simultaneous requests)

### Requirement: Fall back to regex extraction when GitHub API is unavailable
When `GITHUB_TOKEN` is not set or the API call fails, the system SHALL fall back to extracting GitHub-style references (`#<number>`) from both the commit subject and body. References SHALL be deduplicated and returned as a string slice.

#### Scenario: No token available
- **WHEN** `GITHUB_TOKEN` is not set and the commit subject is `"Add login validation (#42)"`
- **THEN** the extracted references SHALL be `["#42"]`

#### Scenario: API call fails
- **WHEN** `GITHUB_TOKEN` is set but the API returns an error for a commit
- **THEN** the system SHALL fall back to regex extraction for that commit without failing the release

#### Scenario: Multiple references in body trailers (fallback)
- **WHEN** falling back to regex and the commit body contains `"Closes #15\nRefs #3"`
- **THEN** the extracted references SHALL include `["#3", "#15"]`

#### Scenario: References in both subject and body (fallback)
- **WHEN** falling back to regex and the subject is `"Fix crash (#42)"` and the body contains `"Fixes #15"`
- **THEN** the extracted references SHALL be `["#15", "#42"]` (deduplicated and sorted)

#### Scenario: No references anywhere
- **WHEN** the commit subject is `"Update README"` and the body is empty
- **THEN** the extracted references SHALL be an empty slice
