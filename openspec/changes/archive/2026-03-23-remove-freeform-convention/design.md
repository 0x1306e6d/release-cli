## Context

The `freeform` convention was added as an explicit option for users who don't follow any structured commit format. However, the system already has a default path for this exact case: when `changes.commits` is absent, it accepts all commits with patch bump and flat changelog. Having both paths is redundant and makes the config surface confusing.

## Goals / Non-Goals

**Goals:**
- Remove `freeform` as an explicit convention value
- Keep the existing default behavior when `changes.commits` is absent (accept all, patch, flat changelog)
- Clean up the code: remove `FreeformCommits` struct and its test

**Non-Goals:**
- Changing the default behavior when no convention is configured
- Adding a migration command or automatic config rewriting

## Decisions

### Remove `freeform` from valid conventions list

Remove `freeform` from `validConventions` in `validate.go`. Valid values become `conventional`, `angular`, `custom`.

**Rationale**: `freeform` adds no value over omitting `changes.commits`. Removing it simplifies the config surface.

### Delete `FreeformCommits` implementation

Remove `internal/commits/freeform.go` and `internal/commits/freeform_test.go`. In `ResolveConvention`, remove the `"freeform"` case. The `CommitConventionParams()` method already returns `"freeform"` when `Commits` is nil — but since `ResolveConvention` is only called when there's a convention configured, this internal sentinel value needs to change too.

**Approach**: When `Commits` is nil, `CommitConventionParams()` should return empty string `""` instead of `"freeform"`. The pipeline already checks `Commits == nil` to decide freeform behavior — it doesn't route through `ResolveConvention` at all. The returned convention name is only passed to `ResolveConvention` when `Commits != nil`.

### Update `IsGroupedChangelog` to check nil only

Currently checks `c.Commits.Convention != "freeform"`. After removal, just check `c.Commits != nil` — if commits config exists, the convention is always structured (conventional/angular/custom), so it always groups.

### Update `.release.yaml` and init template

- Remove the `changes` section from the project's own `.release.yaml` (just `project: go` + `publish`)
- Remove `freeform` from the init template comment listing valid conventions

## Risks / Trade-offs

- **Breaking change**: Users with `convention: freeform` will get a validation error. Mitigation: the fix is simple — delete the `changes.commits` section. The error message from validation will list valid conventions, making it clear `freeform` is no longer accepted.
