## 1. Data Model Changes

- [x] 1.1 Add `Hash` field to `ParsedCommit` struct in `internal/commits/parser.go`
- [x] 1.2 Copy `RawCommit.Hash` into `ParsedCommit.Hash` in the `Analyze` function
- [x] 1.3 Introduce `Item` struct in `internal/changelog/generator.go` with `Hash`, `Title`, `References` fields
- [x] 1.4 Change `Entry.Groups` from `map[string][]string` to `map[string][]Item`

## 2. GitHub API PR Lookup

- [x] 2.1 Create `internal/github/pulls.go` with a client that queries `GET /repos/{owner}/{repo}/commits/{sha}/pulls` to resolve PR numbers and linked issues for a commit
- [x] 2.2 Add bounded-concurrency batch lookup function (`ResolveCommitPRs`) that processes a slice of commit hashes with a goroutine pool (max 5 concurrent)
- [x] 2.3 Create `internal/github/pulls_test.go` with tests using httptest for: commit with PR, commit without PR, PR with linked issues, API error graceful fallback

## 3. Regex Fallback Extraction

- [x] 3.1 Create `internal/changelog/refs.go` with `ExtractReferences(subject, body string) []string` that extracts `#N` patterns, deduplicates, and sorts
- [x] 3.2 Create `internal/changelog/refs_test.go` with tests for subject-only, body-only, both, and no-reference scenarios

## 4. Changelog Generation

- [x] 4.1 Update `Generate` function signature to accept resolved PR metadata alongside `[]ParsedCommit`, and build `[]Item` instead of `[]string`
- [x] 4.2 Populate each `Item` with `Hash` (first 7 chars), `Title`, `References` (from API results or regex fallback)
- [x] 4.3 Update `RenderBody` to format each item as `- <hash> - <title> (<refs>)`
- [x] 4.4 Update `internal/changelog/generator_test.go` to verify the new output format

## 5. Pipeline Integration

- [x] 5.1 Wire GitHub API PR lookup into the pipeline: resolve owner/repo, call `ResolveCommitPRs` when token is available, pass results to `Generate`
- [x] 5.2 Fall back to regex extraction when `GITHUB_TOKEN` is not set or API fails

## 6. Custom Template Contract

- [x] 6.1 Verify `RenderCustom` in `internal/changelog/template.go` works with the new `Entry` shape (Groups with `Item` structs)
- [x] 6.2 Update `internal/changelog/template_test.go` if it exists, or add a test that a custom template can access `.Hash`, `.Title`, `.References`

## 7. Spec Sync

- [x] 7.1 Archive the change to update `openspec/specs/changelog-generation/spec.md` with the modified requirements
