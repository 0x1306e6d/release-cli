## Phase 1. Add release configuration

- [x] 1.1 Create `.release.yaml` at project root with Go project type, conventional commits, semver scheme, changelog enabled, and GitHub Releases publishing with artifact globs for all platform tarballs
- [x] 1.2 Add `dist/` to `.gitignore`
- [x] 1.3 Verify `release-cli status` reads version from git tags and reports correct project type

## Phase 2. Add Makefile for cross-compilation

- [x] 2.1 Create `Makefile` with `build` target that compiles for host OS/arch into `./dist/`, injecting version via ldflags from git tag
- [x] 2.2 Add `release-artifacts` target that cross-compiles for linux-amd64, linux-arm64, darwin-amd64, darwin-arm64 and produces `release-cli-<os>-<arch>.tar.gz` tarballs in `./dist/`
- [x] 2.3 Add `clean` target that removes the `./dist/` directory
- [x] 2.4 Run `make build` and verify `./dist/release-cli --version` outputs the correct tag
- [x] 2.5 Run `make release-artifacts` and verify all 4 platform tarballs exist in `./dist/`

## Phase 3. Add CI release workflow

- [x] 3.1 Create `.github/workflows/release.yml` triggered on `v*` tag push and `workflow_dispatch`
- [x] 3.2 Configure workflow permissions with `contents: write`
- [x] 3.3 Add steps: checkout, setup-go, bootstrap build (`go build`), `make release-artifacts`, `release-cli release`
- [x] 3.4 Ensure the bootstrap-built binary is on PATH before running `release-cli release`
- [x] 3.5 Run `release-cli release --dry-run` and verify the full pipeline previews correctly
