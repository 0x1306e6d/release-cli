## 1. Git Remote Parsing

- [x] 1.1 Add `RemoteOwnerRepo(dir string) (owner, repo string, err error)` function in `internal/git/` that parses the `origin` remote URL to extract owner and repo name, supporting both HTTPS and SSH formats
- [x] 1.2 Add tests for `RemoteOwnerRepo` covering HTTPS URLs, SSH URLs, URLs without `.git` suffix, and unparseable URLs

## 2. Wire Publish Step into Pipeline

- [x] 2.1 Import `publish` and `os` packages in `internal/pipeline/pipeline.go`
- [x] 2.2 Replace the publish placeholder (step 13) with code that: checks if GitHub publish is enabled, reads `GITHUB_TOKEN` from env, calls `RemoteOwnerRepo`, constructs `GitHubPublisher`, and calls `Publish()` with `ReleaseInfo` populated from the pipeline context (tag, version, changelog content)
- [x] 2.3 Skip publish with a warning when `GITHUB_TOKEN` is not set or remote URL cannot be parsed
- [x] 2.4 Add dry-run reporting for the publish step: print `[dry-run] Would publish GitHub Release` when publish is enabled

## 3. Verification

- [x] 3.1 Run existing tests to confirm no regressions (`go test ./...`)
- [x] 3.2 Verify `go build ./...` succeeds
