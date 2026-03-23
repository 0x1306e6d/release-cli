## 1. Remove freeform implementation

- [x] 1.1 Delete `internal/commits/freeform.go`
- [x] 1.2 Delete `internal/commits/freeform_test.go`
- [x] 1.3 Remove `"freeform"` case from `ResolveConvention` in `internal/commits/parser.go`

## 2. Update config

- [x] 2.1 Remove `"freeform"` from `validConventions` in `internal/config/validate.go`
- [x] 2.2 Change `CommitConventionParams()` to return `""` instead of `"freeform"` when `Commits` is nil
- [x] 2.3 Update `IsGroupedChangelog()` to check `c.Commits != nil` only (remove `"freeform"` string check)
- [x] 2.4 Update init template in `internal/cli/init.go` to list `conventional, angular, custom` (drop `freeform`)

## 3. Update project config

- [x] 3.1 Remove the `changes` section from `.release.yaml` (keep `project` and `publish` only)

## 4. Update tests

- [x] 4.1 Update `pipeline_test.go` freeform test to use nil `Commits` instead of `convention: freeform`
- [x] 4.2 Run all tests and verify they pass
