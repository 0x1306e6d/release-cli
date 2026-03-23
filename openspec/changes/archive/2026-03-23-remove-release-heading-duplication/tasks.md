## 1. Changelog Rendering

- [x] 1.1 Add `RenderBody()` method to `Entry` in `internal/changelog/generator.go` that renders the changelog content without the `## {Version} ({Date})` heading line
- [x] 1.2 Add unit tests for `RenderBody()` covering grouped and ungrouped entries

## 2. Pipeline Integration

- [x] 2.1 Update `pipeline.go` step 9 to generate both the full changelog content (for CHANGELOG.md) and the headless body (for the publisher)
- [x] 2.2 Pass the headless body to `runGitHubPublish` instead of the full changelog content

## 3. Verification

- [x] 3.1 Run existing tests to confirm `Render()` behavior is unchanged
- [x] 3.2 Verify the pipeline passes the headless body through to the `ReleaseInfo.ChangelogBody` field
