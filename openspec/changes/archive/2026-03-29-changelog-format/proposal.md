## Why

The current changelog renders only the commit subject as a flat bullet (`- <subject>`), losing useful context like which commit produced the entry and which PR it came from. Teams reviewing a changelog need to quickly trace an entry back to its commit and PR for context.

## What Changes

- Render abbreviated commit hash at the start of each changelog line.
- Query the GitHub API to resolve PR numbers and linked issues for each commit.
- Fall back to regex extraction from commit messages when GitHub token is unavailable.
- Render PR and issue references inline.
- Update the `Entry` data model to carry structured item data (hash, references) instead of plain strings.
- Update the default template and custom-template data contract to expose the new fields.

## Capabilities

### New Capabilities
- `commit-metadata-extraction`: Resolve PR numbers and linked issue references for each commit via the GitHub API (with regex fallback) so they are available to the changelog renderer.

### Modified Capabilities
- `changelog-generation`: Entry items change from plain strings to structured objects carrying hash and references. Default rendering format becomes `- <hash> - <title> (<PR, related-issues>)`.

## Impact

- `internal/changelog/generator.go` -- Entry struct and Render/RenderBody methods.
- `internal/changelog/template.go` -- custom template data contract gains new fields.
- `internal/pipeline/pipeline.go` -- passes commit metadata through to changelog generation.
- `internal/github/` -- new package for GitHub API client (PR lookup per commit).
- `internal/git/commit.go` -- may need to expose short hash helper.
- `internal/commits/parser.go` -- ParsedCommit may carry extracted references.
- Existing custom `.tmpl` files are a **BREAKING** change: the template data shape changes (Groups values become structs instead of strings).
- `openspec/specs/changelog-generation/spec.md` -- spec update needed.
