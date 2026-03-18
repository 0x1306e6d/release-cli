## ADDED Requirements

### Requirement: Makefile with cross-compilation targets
The repository SHALL include a `Makefile` that builds release-cli for multiple platforms.

#### Scenario: Default build target
- **WHEN** `make build` is run
- **THEN** a `release-cli` binary for the host OS/arch SHALL be produced in `./dist/`

#### Scenario: Cross-compilation for all platforms
- **WHEN** `make release-artifacts` is run
- **THEN** compressed tarballs SHALL be produced for linux-amd64, linux-arm64, darwin-amd64, and darwin-arm64 in `./dist/`

#### Scenario: Tarball naming convention
- **WHEN** artifacts are built
- **THEN** each tarball SHALL be named `release-cli-<os>-<arch>.tar.gz`

### Requirement: Version injection at build time
The build process SHALL inject the release version into the binary via Go ldflags.

#### Scenario: Version set from git tag
- **WHEN** a binary is built with `make build` or `make release-artifacts`
- **THEN** `internal/cli.Version` SHALL be set to the current git tag (or `dev` if no tag)

#### Scenario: Version displayed at runtime
- **WHEN** `release-cli --version` is run on a release binary
- **THEN** the output SHALL display the version matching the git tag used at build time

### Requirement: Clean target
The Makefile SHALL provide a `clean` target to remove build artifacts.

#### Scenario: Clean removes dist directory
- **WHEN** `make clean` is run
- **THEN** the `./dist/` directory SHALL be removed
