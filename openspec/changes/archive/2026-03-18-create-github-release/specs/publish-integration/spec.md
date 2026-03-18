## MODIFIED Requirements

### Requirement: GitHub Releases as default publish target
The system SHALL support publishing to GitHub Releases as the default publish target for projects hosted on GitHub. The pipeline SHALL construct a `GitHubPublisher` from configuration and environment, derive the repository owner and name from the git remote URL, and call `Publish()` during the publish step.

#### Scenario: Publish GitHub Release
- **WHEN** `publish.github.enabled: true` is configured (the default)
- **AND** the release tag `v1.4.0` is created
- **AND** `GITHUB_TOKEN` is set in the environment
- **THEN** a GitHub Release is created with the tag `v1.4.0` and the changelog entry as the release body

#### Scenario: Publish with artifacts
- **WHEN** `publish.github.artifacts` includes `dist/*.tar.gz`
- **AND** matching files exist in the `dist/` directory
- **THEN** both the GitHub Release is created and the matching files are uploaded as release assets

#### Scenario: Skip publish when token is missing
- **WHEN** `publish.github.enabled: true` is configured
- **AND** `GITHUB_TOKEN` is not set in the environment
- **THEN** the publish step is skipped with a warning message and the pipeline continues

#### Scenario: Skip publish when disabled
- **WHEN** `publish.github.enabled: false` is configured
- **THEN** no GitHub Release is created during the pipeline

#### Scenario: Dry-run publish reporting
- **WHEN** `--dry-run` flag is set
- **AND** `publish.github.enabled: true`
- **THEN** the system prints `[dry-run] Would publish GitHub Release` without making any API calls

## ADDED Requirements

### Requirement: Repository owner and name derived from git remote
The pipeline SHALL automatically derive the GitHub repository owner and name by parsing the `origin` remote URL. Both HTTPS and SSH URL formats SHALL be supported.

#### Scenario: HTTPS remote URL
- **WHEN** the git remote URL is `https://github.com/owner/repo.git`
- **THEN** the owner is parsed as `owner` and the repo is parsed as `repo`

#### Scenario: SSH remote URL
- **WHEN** the git remote URL is `git@github.com:owner/repo.git`
- **THEN** the owner is parsed as `owner` and the repo is parsed as `repo`

#### Scenario: Remote URL without .git suffix
- **WHEN** the git remote URL is `https://github.com/owner/repo`
- **THEN** the owner is parsed as `owner` and the repo is parsed as `repo`

#### Scenario: Unparseable remote URL
- **WHEN** the git remote URL cannot be parsed to extract owner and repo
- **THEN** the publish step is skipped with a warning message
