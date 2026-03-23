## Why

The `freeform` convention is semantically identical to not setting a convention at all. When `changes.commits` is absent, the system already defaults to freeform behavior (accept all commits, patch bump, flat changelog). Having an explicit `freeform` value is redundant and confusing — it suggests there's a "convention" when the whole point is that there isn't one.

## What Changes

- **BREAKING**: Remove `freeform` as a valid `changes.commits.convention` value
- When `changes.commits` is absent, retain existing behavior: accept all commits, patch bump, flat changelog
- Update validation to reject `freeform` as an unknown convention
- Update the project's own `.release.yaml` to remove `changes.commits.convention: freeform` (just omit the `changes` section)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `freeform-convention`: Remove this capability entirely — the behavior it describes is the default when `changes.commits` is absent
- `release-config`: Remove `freeform` from valid convention values; update defaults documentation
- `commit-analysis`: Remove freeform convention references; clarify that absent `changes.commits` means "accept all commits"

## Impact

- **Config**: Users with `convention: freeform` must remove the `changes.commits` section (or at minimum the `convention` line)
- **Code**: Remove `internal/commits/freeform.go` and `freeform_test.go`; update parser, config, validation
- **Docs**: Update init template comments to no longer list `freeform`
