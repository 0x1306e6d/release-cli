## 0.4.0 (2026-04-20)
- 7dbd7d7 - Allow monorepo root to omit project as a container (#1) (#1)

## 0.3.0 (2026-03-30)
- 46d84cb - Revise README structure and fix inaccurate documentation
- 2ea426a - Add doc-coauthoring, frontend-design, and skill-creator skills
- f522a56 - Add structured changelog format with commit hash and PR references
- dcd8538 - Add Claude rules for commit and PR conventions
- 83145be - Add monorepo support with hierarchical multi-package releases

## 0.2.0 (2026-03-23)
- Write README with installation, usage, configuration, and CI/CD docs
- Remove freeform as an explicit convention, default to accept-all when omitted
- Remove duplicate release heading by passing body-only content to GitHub publish

## 0.1.0 (2026-03-23)

- Add bump override input to release workflow
- Replace categorize with changes.commits, add --bump override and flat changelog
- Rename commits config section to categorize

## 0.0.2 (2026-03-18)

### Other

- Wire GitHub Release publishing into release pipeline

# Changelog

## 0.0.1 (2026-03-18)

### Other

- Archive add-freeform-convention and sync freeform spec
- Combine push into single git command for atomicity
- Push release commit and tag to remote after creation
- Fix release workflow by configuring git identity for CI
- Fix config validation to accept freeform commit convention
- Add freeform commit convention for plain English messages
- Add dogfooding release config, build pipeline, and CI workflow
- Implement core release CLI packages and wire up commands
- Scaffold Go project with cobra CLI and root command
- Add design-release-cli OpenSpec change artifacts
- Initialize OpenSpec
- Add VS Code and macOS entries to .gitignore
- Initial commit
