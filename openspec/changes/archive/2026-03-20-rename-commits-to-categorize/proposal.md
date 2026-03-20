## Why

The `commits` config section name describes the *input* (commit messages) rather than what the section actually controls: how commits are *categorized* into version bump levels. Renaming to `categorize` better communicates intent — the section defines the categorization strategy, not the commits themselves. This is a naming clarification while the config surface is still small and adoption is early.

## What Changes

- **BREAKING**: Rename the `commits` top-level config key to `categorize` in `.release.yaml`
- Rename the Go config structs and fields from `CommitsConfig` to `CategorizeConfig`
- Update the known top-level keys map, validation, defaults, and init template
- Update all internal references from `cfg.Commits` to `cfg.Categorize`
- Update the project's own `.release.yaml` to use the new key name

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `release-config`: Rename `commits` top-level section to `categorize` in the config schema
- `commit-analysis`: Update config references from `commits.convention` / `commits.types` to `categorize.convention` / `categorize.types`

## Impact

- `internal/config/config.go`: Rename `CommitsConfig` struct and `Commits` field to `CategorizeConfig` / `Categorize`; update YAML tag
- `internal/config/validate.go`: Update validation references from `commits.*` to `categorize.*`
- `internal/config/load.go`: Update `knownTopLevelKeys` entry and default application
- `internal/pipeline/pipeline.go`: Update `resolveConvention` to read from `cfg.Categorize`
- `internal/cli/status.go`: Update convention resolution references
- `internal/cli/init.go`: Update generated config template key from `commits:` to `categorize:`
- `.release.yaml`: Update config key
- Tests referencing `commits` config key
