## 1. Config Restructuring

- [x] 1.1 Replace `Categorize CategorizeConfig` with `Changes ChangesConfig` in `Config` struct in `internal/config/config.go`. Add `ChangesConfig` with `Commits *CommitsConfig`. Rename `CategorizeConfig` to `CommitsConfig` and `CategorizeTypesConfig` to `CommitTypesConfig`.
- [x] 1.2 Update `applyDefaults()`: remove default convention assignment (empty = no categorization). Only apply commit-specific defaults when `c.Changes.Commits` is non-nil.
- [x] 1.3 Update `validate()` in `internal/config/validate.go`: change `c.Categorize` to `c.Changes.Commits`, skip convention validation when `Commits` is nil
- [x] 1.4 Replace `"categorize"` with `"changes"` in `knownTopLevelKeys` in `internal/config/load.go`
- [x] 1.5 Update `internal/config/load_test.go`: change `cfg.Categorize` to `cfg.Changes.Commits`
- [x] 1.6 Update `resolveConvention()` in `internal/pipeline/pipeline.go`: read from `cfg.Changes.Commits`, default to freeform when nil
- [x] 1.7 Update `resolveStatusConvention()` in `internal/cli/status.go`: read from `cfg.Changes.Commits`
- [x] 1.8 Update all test configs in `internal/pipeline/pipeline_test.go`: replace `Categorize` with `Changes.Commits`
- [x] 1.9 Update `generateConfig()` in `internal/cli/init.go`: nest convention under `changes.commits` section
- [x] 1.10 Update `.release.yaml`: restructure to new format with convention under `changes.commits`

## 2. Flat Changelog Rendering

- [x] 2.1 Add `Grouped bool` field to `Entry` struct in `internal/changelog/generator.go`
- [x] 2.2 Update `Render()` to render flat list (no `###` headings) when `Grouped` is `false`
- [x] 2.3 Set `entry.Grouped` in pipeline based on convention: `true` for conventional/angular/custom, `false` for freeform/empty
- [x] 2.4 Add `TestRender_Ungrouped_FlatList` in `internal/changelog/generator_test.go`
- [x] 2.5 Update existing `TestRender` to set `Grouped: true`

## 3. Bump Override

- [x] 3.1 Add `ParseBumpType(string) (BumpType, error)` to `internal/version/semver.go`
- [x] 3.2 Add `TestParseBumpType` table-driven test in `internal/version/semver_test.go`
- [x] 3.3 Add `BumpOverride *version.BumpType` to `Options` struct in `internal/pipeline/pipeline.go`
- [x] 3.4 Apply bump override after `commits.Analyze` and before "no releasable changes" check in `Run()`
- [x] 3.5 Add `--bump` flag to `releaseCmd` in `internal/cli/release.go`, parse and pass to `pipeline.Options`

## 4. Integration Tests

- [x] 4.1 Add `TestPipeline_BumpOverride` in `internal/pipeline/pipeline_test.go`: conventional fix commit + override minor → minor bump
- [x] 4.2 Add `TestPipeline_NoCategorize_FlatChangelog` in `internal/pipeline/pipeline_test.go`: no convention, verify flat changelog
- [x] 4.3 Update `TestPipeline_FreeformConvention` to verify flat changelog (no `### Other` heading)
- [x] 4.4 Run full test suite: `go test ./internal/config/ ./internal/changelog/ ./internal/pipeline/ ./internal/version/`
