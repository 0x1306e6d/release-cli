## Context

The changelog currently stores each entry as a plain string in `Entry.Groups` (`map[string][]string`). The `Generate` function receives `[]ParsedCommit` but discards the commit hash, body, and any PR/issue references. The desired format is:

```
- <hash> - <title> (<PR, related-issues>)
```

This requires structured item data flowing from git commits through parsing into changelog rendering.

## Goals / Non-Goals

**Goals:**
- Render abbreviated commit hash at the start of each changelog line
- Query GitHub API to resolve PR numbers and linked issues for each commit
- Fall back to regex extraction from commit messages when no token is available
- Preserve grouped and flat rendering modes
- Update the custom template data contract so users can access the new fields

**Non-Goals:**
- Changing how commits are parsed by conventions (type, scope, bump detection stays the same)
- Backward-compatible template data -- this is a breaking change to the `Entry` shape

## Decisions

### 1. Introduce an `Item` struct to replace plain strings

The `Entry.Groups` field changes from `map[string][]string` to `map[string][]Item`.

```go
type Item struct {
    Hash       string   // abbreviated commit hash (first 7 chars)
    Title      string   // commit subject (after convention parsing)
    References []string // PR numbers and issue refs, e.g., ["#42", "#15"]
}
```

**Why over alternatives:**
- Keeping plain strings and encoding hash/refs into the string would make custom templates harder to use and break parsing.
- A separate parallel slice (hashes []string alongside descriptions []string) would be fragile and harder to extend.

### 2. Pass `RawCommit` data alongside `ParsedCommit` into `Generate`

Currently `Generate` only receives `[]ParsedCommit`. The hash lives on `RawCommit`, not `ParsedCommit`. Two options:

- **Option A**: Add `Hash` field to `ParsedCommit`.
- **Option B**: Change `Generate` signature to accept a paired struct or both slices.

**Decision**: Option A -- add `Hash` to `ParsedCommit`. It is the simplest change. The parser already receives `RawCommit` data; the `Analyze` function just needs to copy the hash through.

### 3. Query GitHub API for PR metadata (primary), regex fallback (secondary)

**Primary path -- GitHub API:**
Use `GET /repos/{owner}/{repo}/commits/{sha}/pulls` to find PRs associated with each commit. This endpoint returns the PR number and any linked issues. A new `internal/github/` package will encapsulate this client logic, reusing the existing `GITHUB_TOKEN` environment variable and `git.RemoteOwnerRepo()` for owner/repo resolution.

To keep the release fast, batch lookups concurrently (bounded goroutine pool, e.g., 5 concurrent requests). Cache results in-memory per pipeline run.

**Fallback path -- regex extraction:**
When `GITHUB_TOKEN` is not set or the API call fails, fall back to regex extraction from the commit subject and body. PR numbers and issue references appear in two common patterns:
- Merge commit subjects: `Merge pull request #42 from ...` or GitHub squash: `Add feature (#42)`
- Commit body trailers: `Closes #15`, `Fixes #7`, `Refs #3`

A regex-based extraction function in `internal/changelog/refs.go` will scan both subject and body.

**Pattern**: `#(\d+)` captures all `#N` references. Deduplicate and sort.

### 4. Default rendering format

```
- a1b2c3d - Add login validation (#42, #15)
```

When no references exist: `- a1b2c3d - Add login validation`
When scope is present: `- a1b2c3d - **auth:** Add login validation (#42)`

### 5. Custom template contract update

The `Entry` struct passed to custom templates changes. The `Item` struct fields are exposed directly. This is a **breaking change** for existing `.tmpl` files. Document the migration in the release notes.

## Risks / Trade-offs

- **[Breaking change for custom templates]** -> Document in CHANGELOG and provide a migration example. The number of users with custom templates is expected to be very small at this stage.
- **[GitHub API rate limiting]** -> With many commits, concurrent API calls could hit rate limits. Mitigated by bounding concurrency and falling back to regex when API fails. Typical releases have tens of commits, well within limits.
- **[API unavailability]** -> If the token is missing or the API is unreachable, the regex fallback ensures changelog generation is never blocked.
- **[Regex fallback may over-match]** -> `#(\d+)` could match non-PR numbers in commit messages. Acceptable trade-off since GitHub linkifies the same pattern. Users can use custom templates to suppress references if needed.
