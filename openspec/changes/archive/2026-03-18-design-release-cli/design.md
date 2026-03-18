## Context

This is a greenfield Go CLI tool. There is no existing codebase — only a repository with a README, license, and gitignore. The tool must work across any programming language ecosystem, detecting project types and executing release workflows with strong conventions and minimal configuration.

The release automation space has existing tools (semantic-release, release-it, changesets, goreleaser), but they are either language-specific, overly complex, or too flexible (leading to inconsistency). release-cli occupies the niche of: language-agnostic, convention-driven, single binary.

## Goals / Non-Goals

**Goals:**
- Single Go binary that works on any project regardless of language/ecosystem
- Detect project type automatically, with explicit declaration in config
- Manifest file as version source of truth (git-tag fallback for ecosystems without version files)
- Strong conventions — most projects need only `project: <type>` in config
- Consistent release pipeline across all project types
- Easy to contribute new detectors (open source, no plugin system)
- Architecture that allows monorepo support in the future

**Non-Goals:**
- Monorepo support in v1
- User-side plugin system (detectors are contributed upstream)
- Fully customizable pipeline ordering (hooks yes, reordering steps no)
- CI/CD pipeline generation (the tool runs in CI, doesn't generate CI config)
- Package manager operations beyond version bumping (no dependency resolution)

## Decisions

### 1. Project type as the central abstraction

The `project` field in `.release.yaml` (e.g., `go`, `node`, `java-gradle`) drives all behavior. Each project type maps to a detector that knows: where the manifest is, how to read/write versions, what publish targets exist, and what ecosystem conventions apply.

General identifiers (`java`) auto-resolve to specific ones (`java-gradle`) by scanning for build tool files. Specific identifiers skip detection.

**Why over alternatives:**
- *Alternative: separate `language` + `build-tool` + `manifest` fields* — More flexible but more config. The compound identifier (`java-gradle`) encodes the same information in one field.
- *Alternative: auto-detect everything at runtime, no config* — Less explicit. Users can't see what the tool will do without running it. Breaks the "config is the declaration" principle.

### 2. Detector interface for extensibility

All project-type-specific logic lives behind a `Detector` interface:

```
┌─────────────────────────────────────────────┐
│              Detector Interface              │
├─────────────────────────────────────────────┤
│ Name() string                               │
│ Aliases() []string                          │
│ Detect(dir string) bool                     │
│ ReadVersion(dir string) (Version, error)    │
│ WriteVersion(dir string, v Version) error   │
│ DefaultPublishTargets() []string            │
│ SnapshotSuffix() string                     │
└─────────────────────────────────────────────┘
```

A registry holds all detectors. General identifiers (e.g., `java`) run detection across all Java-family detectors. Specific identifiers (e.g., `java-gradle`) look up directly.

**Why over alternatives:**
- *Alternative: plugin system with dynamic loading* — Adds complexity for users and maintainers. Since the tool is open source, contributors add detectors via PRs. No runtime plugin loading needed.
- *Alternative: config-driven manifests (user specifies file + field)* — Still needed as an override, but auto-detection from project type covers 90% of cases.

### 3. Manifest as version source of truth with git-tag fallback

Most ecosystems have a canonical version file (package.json, Cargo.toml, pyproject.toml). For these, the manifest is authoritative — you can download a zip of the repo and know the version.

For ecosystems without a version file (Go, Swift), git tags are the version source. The detector declares which strategy its ecosystem uses.

Git tags are **always** created on release, regardless of strategy. Tags are derived from the manifest (manifest strategy) or are the primary version record (tag strategy).

**Why over alternatives:**
- *Alternative: git tags always as source of truth* — Breaks the "version lives without git" requirement. Most ecosystems expect version in a file.
- *Alternative: dedicated VERSION file for all projects* — Fights ecosystem conventions. package.json already has a version field.

### 4. SNAPSHOT / pre-release lifecycle

After a release, the manifest can optionally bump to a pre-release version (e.g., `1.4.0-SNAPSHOT` for Java, `1.4.0.dev0` for Python). This is auto-detected from the current manifest version and configurable.

The SNAPSHOT version is a development marker, not an input to version calculation. The next release version is always derived from: last release tag + commit analysis. If commits warrant a higher bump than the SNAPSHOT implies, the commit analysis wins.

Release with snapshots produces two commits:
1. Release commit: bump to `1.4.0`, changelog, tag
2. Prepare commit: bump to `1.5.0-SNAPSHOT`

**Why over alternatives:**
- *Alternative: no snapshot support* — Ignores a widespread convention (especially Java/Maven). Projects using SNAPSHOT would need manual post-release steps.
- *Alternative: SNAPSHOT version determines release version* — Fragile. If a breaking change lands after the SNAPSHOT was set, the version would be wrong.

### 5. Fixed pipeline with lifecycle hooks

The release pipeline has a fixed order: detect → analyze → bump → propagate → changelog → commit → tag → publish → notify. Users cannot reorder steps but can inject commands at lifecycle points (pre-bump, post-bump, pre-publish, post-publish).

**Why over alternatives:**
- *Alternative: fully customizable pipeline (user defines step order)* — This is the Makefile approach. It's flexible but leads to the exact inconsistency problem we're solving.
- *Alternative: no hooks* — Too rigid. Projects need to run tests, build artifacts, etc. at specific points.

### 6. Propagation as explicit configuration

Version propagation targets (Dockerfile labels, version constants, config files) are explicitly listed in `.release.yaml`. The tool supports built-in propagation types (e.g., `docker-label`) and pattern-based propagation for arbitrary files.

Propagation cannot be auto-detected — a Dockerfile might or might not have a version label. This is the one area where explicit config is required.

**Why over alternatives:**
- *Alternative: auto-scan all files for version strings* — Too magical, too error-prone. Would match version strings that shouldn't be updated.
- *Alternative: no propagation, users do it in hooks* — Fragile. A shell script to update version strings is exactly the kind of inconsistency we're eliminating.

### 7. Project structure

```
release-cli/
├── cmd/
│   └── release-cli/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/                     # Cobra commands
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── release.go
│   │   └── status.go
│   ├── config/                  # .release.yaml parsing
│   │   ├── config.go
│   │   └── validate.go
│   ├── detector/                # Project type detectors
│   │   ├── detector.go          # Interface
│   │   ├── registry.go          # Detector registry
│   │   ├── go.go
│   │   ├── node.go
│   │   ├── python.go
│   │   ├── rust.go
│   │   ├── java_gradle.go
│   │   ├── java_maven.go
│   │   └── ...
│   ├── version/                 # Version parsing, bumping
│   │   ├── semver.go
│   │   ├── calver.go
│   │   └── snapshot.go
│   ├── commits/                 # Commit parsing
│   │   ├── parser.go
│   │   ├── conventional.go
│   │   └── angular.go
│   ├── changelog/               # Changelog generation
│   │   ├── generator.go
│   │   └── template.go
│   ├── propagate/               # Version propagation
│   │   ├── propagator.go
│   │   ├── structured.go        # YAML/JSON/TOML field update
│   │   └── pattern.go           # Regex/template match
│   ├── pipeline/                # Release pipeline orchestration
│   │   ├── pipeline.go
│   │   ├── step.go              # Step interface
│   │   └── hooks.go
│   ├── publish/                 # Publish integrations
│   │   ├── publisher.go         # Interface
│   │   ├── github.go
│   │   └── ...
│   ├── notify/                  # Notification integrations
│   │   ├── notifier.go          # Interface
│   │   ├── slack.go
│   │   └── webhook.go
│   └── git/                     # Git operations
│       ├── git.go
│       ├── tag.go
│       └── commit.go
├── .release.yaml                # Dog-fooding: release-cli releases itself
├── go.mod
└── go.sum
```

## Risks / Trade-offs

- **[Ecosystem coverage]** Cannot ship detectors for every language on day one. → *Mitigation: Start with Go, Node, Python, Rust, Java (Gradle + Maven). Clear contributor guide for adding detectors. The Detector interface is simple to implement.*

- **[Manifest format diversity]** Each ecosystem's manifest has different formats (JSON, TOML, YAML, XML, properties files, Go source). Version read/write logic varies significantly. → *Mitigation: Each detector owns its parsing. Use well-tested Go libraries for each format (encoding/json, BurntSushi/toml, gopkg.in/yaml.v3, etc.).*

- **[SNAPSHOT version divergence]** If a team manually edits the SNAPSHOT version, it could conflict with commit-derived versions. → *Mitigation: SNAPSHOT is treated as a marker, not a version input. Commit analysis always determines the release version from the last tag.*

- **[Multi-language projects]** A repo with Go + Docker + Helm has multiple "project types" but v1 supports only one. → *Mitigation: Use propagation for secondary concerns (Docker labels, Helm chart version). The single `project:` field represents the primary artifact. Monorepo/multi-project support is a future extension.*

- **[Breaking interface changes]** If the Detector interface changes, all detectors must be updated. → *Mitigation: Keep the interface minimal. Use option structs for future extensions rather than adding methods.*

## Open Questions

- **CalVer details**: How does calendar versioning interact with commit analysis? CalVer bumps are date-based, not semantic. Does commit analysis still determine major/minor/patch within a calendar segment?
- **Release candidate workflow**: Should there be explicit RC support (`1.4.0-rc.1` → `1.4.0-rc.2` → `1.4.0`)? How does this interact with SNAPSHOT?
- **Dry-run UX**: Should `release` always show a dry-run preview first and require confirmation? Or is `--dry-run` a separate flag?
- **Tag format**: `v1.2.3` vs `1.2.3` — configurable or convention per ecosystem?
