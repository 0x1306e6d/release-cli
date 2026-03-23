## Context

`Entry.Render()` always emits the version heading (`## 0.1.0 (2026-03-23)`) as the first line. The full rendered string is passed to `GitHubPublisher.Publish()` as `ChangelogBody`, which sets it as the release body. GitHub Releases already show the release name (set to the version) as the page title, so the heading is redundant.

The CHANGELOG.md file must keep the heading — the duplication only affects GitHub Releases.

## Goals / Non-Goals

**Goals:**
- Eliminate the duplicated version heading in GitHub Release notes.
- Keep the change minimal and backwards-compatible.

**Non-Goals:**
- Changing the CHANGELOG.md format or heading style.
- Supporting per-publisher body transformations or a plugin system.

## Decisions

### Add a `RenderBody()` method to `Entry`

Add a new method `RenderBody()` on `Entry` that renders everything *except* the version heading line. `Render()` remains unchanged and continues to be used for CHANGELOG.md.

**Why not a parameter on Render()?** A separate method is clearer — callers that want the heading call `Render()`, callers that want just the body call `RenderBody()`. No risk of breaking existing call sites.

**Why not strip the heading in the publisher?** The publisher shouldn't know about markdown structure. Keeping rendering concerns in the changelog package is cleaner.

### Pipeline passes `RenderBody()` output to the publisher

`runGitHubPublish` already receives `changelogBody` as a plain string. We'll generate the headless body alongside the full content in step 9 of the pipeline and pass it to the publisher instead.

## Risks / Trade-offs

- **Custom templates** (`changelog.template`): `RenderBody()` only applies to the default renderer. Custom templates already control their own format, so this is not a concern — the pipeline will pass the custom-rendered content as-is (users who use custom templates can omit the heading themselves). → No mitigation needed.
