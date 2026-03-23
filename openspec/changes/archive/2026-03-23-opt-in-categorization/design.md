## Context

The release-cli currently requires a `categorize` section in config (defaulting to `conventional` if absent). The `categorize` section controls two concerns: how commits are parsed for bump determination, and how they're grouped in the changelog. For freeform users, this means all commits land in "### Other" and always bump patch — with no override.

The config structure is:
```yaml
categorize:        # top-level
  convention: freeform
  types: ...
changelog:
  enabled: true
  file: CHANGELOG.md
```

Key files: `internal/config/config.go` (struct), `internal/config/validate.go` (validation), `internal/config/load.go` (known keys), `internal/pipeline/pipeline.go` (resolveConvention, Run), `internal/cli/release.go` (CLI), `internal/cli/status.go` (resolveStatusConvention), `internal/cli/init.go` (template), `internal/changelog/generator.go` (Entry, Render).

## Goals / Non-Goals

**Goals:**
- Replace top-level `categorize` with `changes.commits` — a namespace that describes the source of release changes
- Make change source configuration opt-in — absent `changes` = all commits accepted, flat changelog, patch bump
- Add `--bump` CLI flag to override convention-derived bump for any convention
- Flat changelog rendering when no change source is configured
- Structure the config to accommodate future source types (`changes.pulls`, `changes.commits.enrich`)

**Non-Goals:**
- Implementing PR-based changelogs or enrichment (future work)
- Removing the freeform convention implementation (keep it internally for backward compatibility)
- Adding new conventions or changing existing convention behavior
- Modifying the custom template rendering path (`RenderCustom`)

## Decisions

### D1: Introduce `changes` namespace with `commits` sub-section
Replace top-level `CategorizeConfig` with a new `ChangesConfig` struct containing an optional `CommitsConfig`. This reflects that commit parsing is one possible source of change information, not the only one.

```yaml
# Before
categorize:
  convention: conventional
  types: ...

# After
changes:
  commits:
    convention: conventional
    types: ...
```

**Alternative considered**: Move `categorize` under `changelog`. Rejected because commit parsing drives bump determination even when `changelog.enabled: false` — nesting it under changelog misrepresents the relationship.

**Alternative considered**: Keep `categorize` top-level but optional. Rejected because `categorize` names an action, not an input, and doesn't accommodate future source types like PR-based changelogs.

### D2: Empty changes means "accept all commits"
When `Changes.Commits` is nil (no changes section), the pipeline uses the freeform convention internally. This avoids a new code path — freeform already accepts all commits and returns `BumpPatch`.

**Alternative considered**: Create a new "none" convention type. Rejected as unnecessary duplication of freeform behavior.

### D3: Flat rendering via `Grouped` bool on `Entry`
Add a `Grouped bool` field to `changelog.Entry`. The pipeline sets it based on convention: `true` for conventional/angular/custom, `false` for freeform/empty. `Render()` checks this field.

**Alternative considered**: Pass convention string to `Render()`. Rejected because `Grouped` is a cleaner abstraction — the renderer shouldn't need to know about convention names.

### D4: `--bump` flag on release subcommand only
Register `--bump` on `releaseCmd.Flags()` (not persistent). The `status` command shows the convention-derived prediction, unaffected by `--bump`.

**Alternative considered**: Global flag. Rejected because bump override only makes sense during release, not status or init.

### D5: Bump override applied after commit analysis
The pipeline still runs `commits.Analyze()` even when `--bump` is provided — commits are needed for changelog content. The override replaces the derived bump type afterward, before the "no releasable changes" check.

### D6: Config struct design for future extensibility
The `ChangesConfig` struct uses pointer fields for each source type, making presence/absence the opt-in signal:

```go
type ChangesConfig struct {
    Commits *CommitsConfig `yaml:"commits"`
    // Future: Pulls *PullsConfig `yaml:"pulls"`
}

type CommitsConfig struct {
    Convention string              `yaml:"convention"`
    Types      CategorizeTypesConfig `yaml:"types"`
    // Future: Enrich bool `yaml:"enrich"`
}
```

When `Changes.Commits` is nil, no commit convention is configured. This naturally extends to `Changes.Pulls` later without structural changes.

## Risks / Trade-offs

- **[Breaking config change]** → Users with top-level `categorize` must update their config. Acceptable at v0.0.x. The init command generates the new format.
- **[Default convention changes]** → Previously, omitting `categorize` defaulted to `conventional`. Now it defaults to no categorization (freeform behavior). This changes behavior for users with minimal configs. → Mitigated by being pre-1.0 and documenting the change.
- **[Freeform keyword kept valid]** → Users with `convention: freeform` still work. The freeform convention remains in the valid conventions list for backward compatibility.
