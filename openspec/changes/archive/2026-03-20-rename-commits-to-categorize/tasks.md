## 1. Config Layer

- [x] 1.1 Rename `CommitsConfig` to `CategorizeConfig` and `CommitTypesConfig` to `CategorizeTypesConfig` in `internal/config/config.go`; update the struct field from `Commits CommitsConfig` to `Categorize CategorizeConfig` with YAML tag `yaml:"categorize"`
- [x] 1.2 Update `knownTopLevelKeys` in `internal/config/load.go`: replace `"commits"` with `"categorize"`
- [x] 1.3 Update `applyDefaults` in `internal/config/config.go`: change `c.Commits.Convention` to `c.Categorize.Convention`
- [x] 1.4 Update `validate` in `internal/config/validate.go`: change all `c.Commits.*` references to `c.Categorize.*` and update error message strings from `commits.*` to `categorize.*`

## 2. Pipeline and CLI References

- [x] 2.1 Update `resolveConvention` in `internal/pipeline/pipeline.go` to read from `cfg.Categorize` instead of `cfg.Commits`
- [x] 2.2 Update `resolveStatusConvention` in `internal/cli/status.go` to read from `cfg.Categorize`
- [x] 2.3 Update `generateConfig` in `internal/cli/init.go` to output `categorize:` instead of `commits:` in the generated template

## 3. Project Config

- [x] 3.1 Update `.release.yaml` to use `categorize:` instead of `commits:`

## 4. Tests

- [x] 4.1 Update config test fixtures and assertions that reference the `commits` key to use `categorize`
- [x] 4.2 Run full test suite to verify no regressions
