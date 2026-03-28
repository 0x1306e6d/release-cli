## Context

release-cli is a language-agnostic release automation tool that currently supports single-project repositories. The original v1 design doc explicitly listed "Architecture that allows monorepo support in the future" as a goal while deferring implementation. The detector interface already accepts a `dir` parameter, and the pipeline is well-isolated — both signals that the architecture is ready for this extension.

Monorepo usage is common in practice: a root Go module with CLI and workflow sub-packages, a service + Helm chart, or multiple microservices in one repo. Users today must either split into separate repos or maintain manual release scripts per package.

Key architectural constraints:
- Existing single-project configs (`project: go`) MUST continue working unchanged
- The detector interface MUST NOT change (it already accepts `dir string`)
- The pipeline is fixed-order with hooks — this model extends naturally to per-package execution
- Git tags are the release history — monorepo needs namespaced tags to avoid collisions

## Goals / Non-Goals

**Goals:**
- Support multiple independently versioned packages in a hierarchical structure
- Each package has its own `.release.yaml` — fully independent, no config inheritance
- Parent packages declare children; selecting a parent cascades to all descendants
- Path-based commit filtering per package
- Namespaced git tags: root uses `<name>/v<ver>`, children use `<relative-path>/v<ver>`
- Batched releases: cascading produces one commit with multiple tags
- CLI flags to target specific packages (`--package`, multi-value) or all (`--all`)
- Full backward compatibility with single-project configs

**Non-Goals:**
- Config inheritance between parent and child configs
- Cross-package dependency management (e.g., auto-bumping dependents)
- Shared/lockstep versioning across packages
- Automatic package discovery (packages must be explicitly declared)
- Workspace protocol support (npm workspaces, Go workspaces, Cargo workspaces)
- Parallel package releases (packages are released sequentially)

## Decisions

### 1. Multiple `.release.yaml` files — parent declares children

Each package directory has its own `.release.yaml`. A parent declares its children via a `modules` list of relative directory paths:

```yaml
# .release.yaml (root)
name: release-cli
project: go
modules:
  - cli
  - workflow
```

```yaml
# cli/.release.yaml
project: go
```

```yaml
# workflow/.release.yaml
project: go
modules:
  - sub    # workflow/sub/.release.yaml exists
```

Each `.release.yaml` is fully self-contained — it declares its own `project`, config sections, and optionally `modules` (children). A leaf package (no `modules` field) works exactly like today's single-project config.

**Why over alternatives:**
- *Alternative: single `.release.yaml` with all packages declared in one file* — Loses per-package isolation. One large config file becomes hard to manage. Requires config inheritance to avoid duplication, which adds complexity.
- *Alternative: multiple files but children declare their parent* — Inverted dependency. The parent can't enumerate its tree without scanning all directories.
- *Alternative: auto-discover `.release.yaml` files by scanning* — Surprising behavior. Slow on large repos. Explicit declaration is predictable.

### 2. `name` field for root tag prefix

The root `.release.yaml` requires a `name` field that determines its git tag prefix (e.g., `name: release-cli` → tags like `release-cli/v1.2.0`). Children derive their tag prefix from their relative path (e.g., `cli/v1.0.0`, `workflow/sub/v0.3.0`).

`name` is required only when the config has a `modules` list (i.e., it is a parent). Leaf packages and single-project configs do not need it.

**Why over alternatives:**
- *Alternative: derive from directory name* — Directory names can vary across clones. Explicit `name` is reliable.
- *Alternative: configurable tag prefix per package* — Unnecessary flexibility. Path-based prefixes are the widely used convention (Go modules, Lerna, etc.).

### 3. No config inheritance

Each `.release.yaml` is independent. A child does not inherit any config from its parent. If two packages need the same `changes.commits.convention`, both declare it.

**Why over alternatives:**
- *Alternative: parent sections as defaults, children override* — Adds merge semantics that are hard to reason about. Users can't read a child's config in isolation without knowing the parent.
- *Alternative: deep merge* — Even more confusing. Section-level override is simpler but still requires understanding two files to know the effective config.

### 4. Path-based commit filtering — parent sees all

Each package's commit analysis uses `git log -- <path>` to scope commits. A parent at `path: .` sees **all** commits under its path, including those in children's directories. This means the root changelog is a superset.

**Why over alternatives:**
- *Alternative: residual commits only (exclude children's paths)* — More complex git queries. Parent changelogs become confusing — "why is this major feature not in the root changelog?"
- *Alternative: parent has no commits (pure grouping)* — Defeats the purpose of the root being a real package with its own version.

### 5. Hierarchical cascading — force all descendants

When a parent is selected for release, **all** its descendants are released too, even if some have no new commits. This is a forced release — the user explicitly chose the parent knowing it cascades.

Packages without commits since their last tag receive a patch bump by default (unless `--bump` overrides).

**Why over alternatives:**
- *Alternative: skip unchanged descendants* — Defeats the "release together" intent. If you select the root, you want everything released.
- *Alternative: ask per package* — Interactive prompts don't work in CI.

### 6. Batched commit and multiple tags

A cascading release produces **one git commit** containing all version bumps and changelogs, with **one tag per package** pointing to that commit:

```
commit: "Release release-cli 0.3.0, cli 0.3.0, workflow 0.1.5"
  → tag: release-cli/v0.3.0
  → tag: cli/v0.3.0
  → tag: workflow/v0.1.5
```

When releasing a single leaf package (non-cascading), the behavior is the same as today — one commit, one tag.

**Why over alternatives:**
- *Alternative: one commit per package* — Noisy history for coordinated releases. N packages = N commits.
- *Alternative: one tag for all packages* — Defeats independent versioning.

### 7. Namespaced git tags

Tag format by package type:
- **Single-project** (no `modules`): `v1.2.3` (unchanged from today)
- **Root parent**: `<name>/v1.2.3` (e.g., `release-cli/v1.2.3`)
- **Child package**: `<relative-path>/v1.2.3` (e.g., `cli/v1.2.3`, `workflow/sub/v1.0.0`)

`git.LatestSemverTag()` is extended to accept an optional prefix filter.

**Why over alternatives:**
- *Alternative: `v1.2.3-api` suffix* — Ambiguous with pre-release suffixes.
- *Alternative: configurable tag pattern* — Unnecessary flexibility.

### 8. CLI changes

```
release-cli release --package cli                    # Release specific package
release-cli release --package cli --package workflow  # Release multiple packages
release-cli release --all                            # Release all (same as selecting root)
release-cli status --package cli                     # Status for specific package
release-cli status                                   # Status for all packages (summary)
```

`--package` accepts multiple values for ad-hoc grouping (e.g., tightly coupled server + client). In monorepo mode, `release` without `--package` or `--all` reports an error. `--all` is equivalent to selecting the root package.

**Why over alternatives:**
- *Alternative: positional argument (`release-cli release api`)* — Conflicts with future subcommands. Flags are explicit and composable.
- *Alternative: `--group` with config-declared groups* — Over-engineered for v1. Multi-value `--package` covers the same use case without config changes.

### 9. Config discovery and tree resolution

When `release-cli` runs, it reads `.release.yaml` in the current directory. If `modules` is present, it recursively reads each child's `.release.yaml` to build the full package tree. Validation ensures:
- Each declared child directory exists and contains a `.release.yaml`
- No circular references in the tree
- No overlapping paths between siblings

**Why over alternatives:**
- *Alternative: scan all directories for `.release.yaml`* — Auto-discovery is surprising and slow on large repos. Explicit declaration is predictable.

## Risks / Trade-offs

- **[Duplicate config]** Without inheritance, packages may repeat common config (e.g., `changes.commits.convention`). → *Mitigation: Most packages need minimal config (just `project`). The duplication cost is low compared to the complexity cost of inheritance.*

- **[Root changelog noise]** The root sees all commits including children's, making its changelog a superset. → *Mitigation: This is intentional — the root represents the whole project. Users who want a clean root changelog can customize the template or disable it.*

- **[Force-releasing unchanged packages]** Cascading forces releases even without new commits. → *Mitigation: This matches the user's intent — "release everything." The patch bump default is safe. Users who want selective releases use `--package` instead.*

- **[Git history complexity]** Path-filtered `git log` can be slow on large repos with deep history. → *Mitigation: The `-- <path>` filter is well-optimized in git. For very large repos, users can configure shallow clones in CI.*

- **[Batched commit complexity]** A single commit touching multiple packages' manifests and changelogs is harder to revert selectively. → *Mitigation: This is the same trade-off any batched release makes. The alternative (N commits) is worse for history cleanliness.*

- **[Deep nesting]** Deeply nested hierarchies (3+ levels) could produce long tag prefixes. → *Mitigation: This is a rare scenario. The tag format is still valid and unambiguous. Users can keep hierarchies shallow.*
