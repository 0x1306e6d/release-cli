## Why

Many projects use plain English commit messages (e.g., "Add user export", "Fix login bug") without following Conventional Commits or Angular conventions. Currently, release-cli treats all non-conforming commits as non-releasable, causing the pipeline to abort with "No releasable changes found." Projects that don't use structured commit messages cannot use release-cli at all.

## What Changes

- Add a new `freeform` commit convention that treats every commit as a releasable patch bump
- Users can opt in via `commits.convention: freeform` in `.release.yaml`
- The freeform parser uses the full commit subject as the description and assigns `patch` as the bump type
- No special parsing or pattern matching is applied — all commits are accepted

## Capabilities

### New Capabilities
- `freeform-convention`: A commit convention that treats all commits as releasable patch-level changes, requiring no structured commit message format

### Modified Capabilities
- `commit-analysis`: Add `freeform` as a supported value for `commits.convention`, update the requirement for configurable conventions

## Impact

- `internal/commits/`: New `freeform.go` implementing the `Convention` interface
- `internal/pipeline/pipeline.go`: Update `resolveConvention` to handle `"freeform"`
- `internal/config/config.go`: No structural changes needed (convention is already a string field)
- `openspec/specs/commit-analysis/spec.md`: Updated to document the new convention option
