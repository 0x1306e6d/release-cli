## Why

When publishing a GitHub Release, the release title already displays the version (e.g. "0.1.0"). The changelog body passed as the release notes also starts with `## 0.1.0 (2026-03-23)`, creating a redundant heading. This looks unprofessional and wastes vertical space in the release page.

## What Changes

- Changelog rendering gains the ability to omit the version heading, producing a "headless" body suitable for contexts where the version is already displayed externally (e.g. GitHub Releases).
- The GitHub publisher passes the headless changelog body to the release API instead of the full entry.

## Capabilities

### New Capabilities

_None_

### Modified Capabilities

- `publish-integration`: The GitHub release body should use a headless changelog entry (no version heading) instead of the full rendered entry.
- `changelog-generation`: The renderer should support producing output without the version heading line.

## Impact

- `internal/changelog/generator.go` — `Render()` method needs a headless option.
- `internal/publish/github.go` — consumes the headless body.
- `internal/pipeline/pipeline.go` — passes the correct variant to the publisher.
- Existing CHANGELOG.md file output is **not affected**; it continues to include the heading.
