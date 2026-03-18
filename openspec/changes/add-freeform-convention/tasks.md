## 1. Core Implementation

- [x] 1.1 Create `internal/commits/freeform.go` implementing `Convention` interface — `Parse` returns a `ParsedCommit` with `Type: "other"`, `Subject: <full subject>`, `Breaking: false`, `Bump: BumpPatch` for every commit
- [x] 1.2 Add `"freeform"` case to `resolveConvention` switch in `internal/pipeline/pipeline.go`

## 2. Tests

- [x] 2.1 Create `internal/commits/freeform_test.go` — test plain English subjects, empty subjects, subjects with special characters, and verify breaking change footers are ignored
- [x] 2.2 Add pipeline integration test in `internal/pipeline/pipeline_test.go` for freeform convention — verify plain commits produce a patch bump result instead of nil

## 3. Specs

- [x] 3.1 Update `openspec/specs/commit-analysis/spec.md` to add `freeform` as a supported convention value
