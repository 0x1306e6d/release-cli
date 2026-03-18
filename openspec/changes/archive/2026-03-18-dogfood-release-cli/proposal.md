## Why

release-cli is feature-complete but has never been used to release itself. Dogfooding validates the tool end-to-end on a real Go project, surfaces usability gaps, and establishes the project's own release workflow so contributors can ship versions with a single command.

## What Changes

- Add a `.release.yaml` configuration file tailored for the release-cli Go project itself (git-tag versioning, conventional commits, GitHub Releases publishing, changelog generation).
- Add a GitHub Actions CI workflow that runs `release-cli release` on tagged pushes or manual dispatch, producing binaries for linux/darwin (amd64/arm64) and attaching them to a GitHub Release.
- Add a `Makefile` with build, test, and release targets that cross-compile the CLI for distribution.
- Add a `goreleaser`-style build matrix baked into the Makefile (no external tool dependency) so release artifacts are produced deterministically.
- Wire version information into the binary at build time via `-ldflags` so `release-cli --version` reports the real release version.

## Capabilities

### New Capabilities
- `self-release-config`: The `.release.yaml` configuration for dogfooding release-cli on itself, including project type, commit convention, changelog, and publish settings.
- `ci-release-workflow`: GitHub Actions workflow that builds cross-platform binaries and runs `release-cli release` to publish them.
- `build-and-distribute`: Makefile targets for cross-compilation, version injection via ldflags, and artifact packaging.

### Modified Capabilities
- `release-config`: Add documentation/examples showing how a Go project configures `.release.yaml` (the dogfood config serves as the canonical example).

## Impact

- **New files**: `.release.yaml`, `.github/workflows/release.yml`, `Makefile`
- **Modified files**: `cmd/release-cli/main.go` (accept ldflags for version), `internal/cli/root.go` (use injected version)
- **Dependencies**: None new. GitHub Actions uses `actions/checkout`, `actions/setup-go`, and the built binary.
- **Systems**: GitHub Releases will be the distribution channel. Tagged pushes trigger automated releases.
