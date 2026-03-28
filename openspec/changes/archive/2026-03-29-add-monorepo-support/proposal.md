## Why

release-cli currently supports only single-project repositories. Many real-world repositories contain multiple independently versioned packages in a hierarchy (e.g., a root Go module with CLI and workflow sub-packages). Users in monorepo setups cannot use release-cli today without maintaining separate repos or manual workarounds. The original design doc explicitly planned for monorepo support as a future extension, and the detector/pipeline architecture was designed to accommodate it.

## What Changes

- Support multiple `.release.yaml` files: each package directory has its own independent config
- Parent packages declare children via a `modules` list of relative directory paths
- Root config requires a `name` field used as the git tag prefix (e.g., `release-cli/v1.2.0`)
- Child tag prefixes use the full relative path from the repo root (e.g., `cli/v1.0.0`, `workflow/sub/v0.3.0`)
- Hierarchical release cascading: selecting a parent forces release of all its descendants
- Commit scoping per package uses path-based filtering; parent packages see all commits under their path (including children's)
- Batched releases: cascading produces one commit with multiple tags
- Extend `release` and `status` commands with `--package` flag (accepts multiple values) and `--all` flag
- No config inheritance between parent and child — each `.release.yaml` is fully independent
- Full backward compatibility: repos without `modules` behave exactly as today

## Capabilities

### New Capabilities
- `monorepo-config`: Multi-file configuration with hierarchical package declaration, `name` field for tag prefix, and `modules` list for declaring children
- `package-scoped-release`: Hierarchical cascading, path-filtered commit analysis, namespaced tags, batched commit+tags, and per-package pipeline execution

### Modified Capabilities
- `release-config`: Extended to support optional `name` and `modules` fields alongside existing single-project config
- `release-pipeline`: Extended to run per-package with package context, batched commit for cascading releases, and namespaced tags
- `commit-analysis`: Extended to filter commits by file paths within a package directory

## Impact

- **Config**: Each package gets its own `.release.yaml`; parent lists children via `modules`
- **CLI**: `release` and `status` commands gain `--package` (multi-value) and `--all` flags
- **Git tags**: Monorepo packages use `<path>/v1.2.0` format; root uses `<name>/v1.2.0`; single-project repos keep `v1.2.0`
- **Pipeline**: Cascading release produces one batched commit with multiple tags
- **Detectors**: No interface changes needed; detectors already accept a `dir` parameter
- **Existing users**: Zero breaking changes; the feature is entirely additive
