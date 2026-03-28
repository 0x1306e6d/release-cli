## 1. Config Schema Extension

- [x] 1.1 Add `Name string` and `Modules []string` fields to the `Config` struct in `internal/config/config.go`
- [x] 1.2 Add `name` and `modules` to the `knownTopLevelKeys` map in `internal/config/load.go`
- [x] 1.3 Implement validation: `name` is required when `modules` is present
- [x] 1.4 Add `IsMonorepo() bool` helper method on `Config` to check if `modules` is non-empty
- [x] 1.5 Write unit tests for config parsing and validation of `name` and `modules` fields

## 2. Package Tree Resolution

- [x] 2.1 Create `internal/monorepo/tree.go` with a `PackageNode` struct (name, path, config, children) and `PackageTree` type
- [x] 2.2 Implement `LoadTree(rootDir string)` that recursively reads `.release.yaml` from root and all declared children, building the full tree
- [x] 2.3 Implement tree validation: child directories exist, child configs are valid, no circular references, no overlapping sibling paths
- [x] 2.4 Implement `Resolve(packageName string) []*PackageNode` that finds a package by name and returns it plus all descendants (flattened)
- [x] 2.5 Implement tag prefix resolution: root uses `name` field, children use relative path from repo root
- [x] 2.6 Write unit tests for tree loading, validation, and resolution

## 3. Git Tag Namespacing

- [x] 3.1 Extend `git.LatestSemverTag()` in `internal/git/` to accept an optional prefix parameter for filtering tags (e.g., `cli/v*`)
- [x] 3.2 Extend `git.CreateTag()` to support namespaced tag format (`<prefix>/v<version>`)
- [x] 3.3 Write unit tests for prefixed tag lookup and namespaced tag creation

## 4. Path-Based Commit Filtering

- [x] 4.1 Extend `git.CommitsSince()` (or equivalent) to accept an optional path filter, using `git log -- <path>` semantics
- [x] 4.2 Update commit parser in `internal/commits/` to pass path filter through when in monorepo mode
- [x] 4.3 Write unit tests for path-filtered commit retrieval (including parent seeing all commits)

## 5. Package-Scoped Pipeline

- [x] 5.1 Define a `PackageContext` struct holding: package name, path, config, tag prefix, is-forced flag
- [x] 5.2 Refactor `pipeline.Run()` to accept an optional `PackageContext` that scopes operations (detect dir, tag prefix, commit filter path, changelog path)
- [x] 5.3 Scope detector resolution to use package path as the working directory
- [x] 5.4 Scope version read/write to package path manifest or namespaced tags
- [x] 5.5 Scope changelog generation to write under package path by default
- [x] 5.6 When `PackageContext.IsForced` is true and no commits found, apply patch bump instead of exiting
- [x] 5.7 Add `RELEASE_PACKAGE` and `RELEASE_PACKAGE_PATH` environment variables to hook execution
- [x] 5.8 Write unit tests for package-scoped pipeline execution

## 6. Batched Commit and Tags

- [x] 6.1 Create `internal/pipeline/batch.go` with a `BatchRelease` function that coordinates multiple package pipelines
- [x] 6.2 Implement per-package pipeline steps (detect through changelog) running independently for each package
- [x] 6.3 Implement batched git commit: stage all package changes, create one commit with message listing all packages and versions
- [x] 6.4 Implement multiple tag creation: one tag per package on the batched commit
- [x] 6.5 Implement batched push: push the commit and all tags in one operation
- [x] 6.6 Run per-package publish and notify after tagging
- [x] 6.7 Implement batched snapshot step: one commit bumping all packages to next dev version
- [x] 6.8 Write unit tests for batched commit, multi-tag creation, and publish ordering

## 7. CLI Changes

- [x] 7.1 Add `--package` string slice flag to `release` command in `internal/cli/release.go`
- [x] 7.2 Add `--all` bool flag to `release` command
- [x] 7.3 Implement monorepo release logic: load tree, resolve targets (with cascading), construct PackageContexts, run batch pipeline
- [x] 7.4 Implement `--all` logic as equivalent to selecting the root package
- [x] 7.5 Add error when running `release` in monorepo mode without `--package` or `--all`
- [x] 7.6 Add error when using `--package` or `--all` in single-project mode
- [x] 7.7 Add `--package` flag to `status` command in `internal/cli/status.go`
- [x] 7.8 Implement monorepo status summary table (package name, version, pending commits, next bump) for all packages in tree
- [x] 7.9 Implement single-package detailed status when `--package` is specified
- [x] 7.10 Write unit tests for CLI flag validation and monorepo command routing

## 8. Integration Tests

- [x] 8.1 Create test fixture: monorepo with root + two child packages in a temp git repo, each with their own `.release.yaml`
- [x] 8.2 Test: releasing a single leaf package creates namespaced tag and scoped changelog
- [x] 8.3 Test: releasing a parent cascades to all descendants with batched commit and multiple tags
- [x] 8.4 Test: `--all` releases entire tree
- [x] 8.5 Test: multi-value `--package` releases selected packages
- [x] 8.6 Test: unchanged descendant gets force-released with patch bump during cascade
- [x] 8.7 Test: parent commit analysis includes children's commits
- [x] 8.8 Test: dry-run with cascading shows correct preview
- [x] 8.9 Test: single-project config still works unchanged (backward compatibility)
