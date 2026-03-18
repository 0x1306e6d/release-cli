## Context

release-cli currently supports three commit conventions: `conventional` (default), `angular`, and `custom`. All three require commit messages to follow the `type(scope): description` pattern. Projects using plain English commit messages (e.g., "Add feature X", "Fix bug Y") get zero matches from the parser, resulting in "No releasable changes found."

The `Convention` interface (`internal/commits/parser.go`) defines a single `Parse(subject, body string) *ParsedCommit` method. Existing conventions return `nil` for non-matching commits. The `Analyze` function skips nil results, so if all commits return nil, the pipeline aborts.

## Goals / Non-Goals

**Goals:**
- Allow projects with plain English commit messages to use release-cli
- Add a `freeform` convention that treats every commit as a releasable patch bump
- Keep the change minimal — one new file, one switch case, one spec update

**Non-Goals:**
- Keyword-based heuristic parsing (e.g., inferring `minor` from "Add" or "New") — too fragile and opinionated
- Changing the default convention — `conventional` remains the default
- Freeform-specific changelog grouping or categorization

## Decisions

### Decision 1: All freeform commits map to patch

**Choice:** Every commit produces a `patch` bump. No attempt to infer `minor` or `major` from keywords.

**Alternatives considered:**
- Keyword heuristics ("Add" → minor, "Fix" → patch, "BREAKING" → major): Unreliable across languages and styles. Would need constant tuning and would still get things wrong.
- User-configurable regex-to-bump mappings: Over-engineered for the use case. Users who need that level of control can use the `custom` convention.

**Rationale:** Freeform is the "just make it work" option. Users who want semantic bumps should use a structured convention. Patch-only is safe — it never makes an unexpected major/minor release.

### Decision 2: Use full subject as ParsedCommit.Subject, type as "other"

**Choice:** Set `Type: "other"` and `Subject: <full commit subject>`. No scope, no breaking flag.

**Rationale:** The `ParsedCommit` struct requires a Type and Subject. Using "other" as the type keeps changelog grouping simple (all commits fall under one section). The full subject line is the most useful content for changelog entries.

### Decision 3: Convention name is "freeform"

**Choice:** `commits.convention: freeform` in `.release.yaml`.

**Alternatives considered:**
- `plain`: Less descriptive of what the convention actually does
- `all`: Ambiguous — could be misread as "all conventions combined"

## Risks / Trade-offs

- **Every commit triggers a release** → This is by design. Users opting into freeform accept that any push with new commits will bump the patch version. Mitigated by documentation clarity.
- **No major/minor bumps possible** → Users must manually tag or override if they want a minor/major release. Acceptable for the target audience (projects that don't use structured commits).
- **Changelog is flat** → All commits appear under a single "Other" category. This is consistent with having no type information to group by.
