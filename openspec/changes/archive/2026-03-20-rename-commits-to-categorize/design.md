## Context

The `.release.yaml` config uses `commits` as the top-level key for configuring how commit messages are parsed and mapped to version bump levels. The key currently holds two sub-fields: `convention` (which parsing strategy to use) and `types` (custom type-to-bump mappings). The name `commits` describes the input data rather than the action the section controls — categorizing commits into bump levels.

The project is in early adoption (v0.0.2), making this the right time for a naming change before the config surface stabilizes.

## Goals / Non-Goals

**Goals:**
- Rename the `commits` top-level config key to `categorize`
- Rename all associated Go types and struct fields to match
- Update all internal references, validation messages, defaults, and templates
- Keep the sub-field names (`convention`, `types`) unchanged

**Non-Goals:**
- Adding backward compatibility or migration support for the old `commits` key — clean break at this stage
- Changing any behavior of convention parsing or bump determination
- Renaming the `internal/commits/` Go package — it deals with commit parsing logic, not config naming

## Decisions

### Decision 1: Clean rename with no backward compatibility

**Choice:** Replace `commits` with `categorize` everywhere. Old configs using `commits:` will get an "unknown key" warning and fall back to default convention.

**Alternatives considered:**
- Accept both `commits` and `categorize` with deprecation warning: Adds complexity for a v0.0.x project with minimal users. Not worth it.
- Add a migration command: Over-engineered for a single key rename at this stage.

**Rationale:** At v0.0.2 with early adoption, a clean break is simpler and avoids carrying dead code.

### Decision 2: Keep `internal/commits/` package name unchanged

**Choice:** Only rename the config-layer types (`CommitsConfig` → `CategorizeConfig`, `cfg.Commits` → `cfg.Categorize`). The `internal/commits/` package stays as-is.

**Rationale:** The package implements commit parsing logic — its name describes what it does (parsing commits), not the config section. Renaming it would be a larger refactor with no clarity benefit.

### Decision 3: Keep sub-field names unchanged

**Choice:** `convention` and `types` remain as sub-fields under the new `categorize` key.

**Rationale:** These names are already clear. `categorize.convention` and `categorize.types` read well and accurately describe their purpose.

## Risks / Trade-offs

- **Breaking existing configs** → Acceptable at v0.0.2. The unknown-key warning will guide users to update their config.
- **Docs/README references to `commits:`** → Must be updated in the same change to avoid confusion.
