## Why

The `categorize` config is currently a required top-level section that always runs a convention to parse commits. With freeform, all commits land under "### Other" in the changelog (a meaningless heading) and always bump patch (no override). This forces users who don't need categorization to configure it anyway, and provides no way to control the bump level.

Additionally, naming the section `categorize` ties the config to a specific action rather than describing the input. The config should describe *what* the release is built from (changes), not *how* one step works.

## What Changes

- **BREAKING**: Replace top-level `categorize` with `changes.commits` — a new `changes` namespace that describes the source of release changes, with `commits` as the first supported source
- **BREAKING**: Make change source configuration opt-in — when `changes` is absent, all commits are accepted with a flat changelog and patch bump by default
- Flat changelog rendering when no change source is configured (no `### {label}` headings)
- Add `--bump major|minor|patch` CLI flag to override convention-derived bump for any convention

## Capabilities

### New Capabilities

- `bump-override`: CLI flag `--bump` that overrides the convention-derived version bump level

### Modified Capabilities

- `release-config`: Replace top-level `categorize` with `changes.commits`, make it optional
- `changelog-generation`: Flat rendering (no group headings) when no change source is configured
- `commit-analysis`: When no change source is configured, accept all commits with patch bump (freeform behavior internally)
- `release-pipeline`: Wire `--bump` override into the pipeline, use it before the "no releasable changes" check

## Impact

- **Config**: `.release.yaml` structure changes — `categorize` is replaced by `changes.commits`. Existing configs with top-level `categorize` will need to be updated.
- **CLI**: New `--bump` flag on `release` subcommand
- **Internal packages**: `config`, `pipeline`, `changelog`, `cli`, `version`

## Future Extensibility

The `changes` namespace is designed to accommodate additional source types without breaking changes:

- `changes.commits.enrich: true` — enrich commit-based changelogs with linked PR/issue context from GitHub
- `changes.pulls` — PR/milestone-based changelogs as an alternative to commit parsing
