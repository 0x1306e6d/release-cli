package commits

import (
	"github.com/0x1306e6d/release-cli/internal/version"
)

// AngularCommits implements the Angular commit convention.
// Angular format is the same as Conventional Commits: type(scope): description.
// The key difference is the type-to-bump mapping.
type AngularCommits struct{}

func (a *AngularCommits) Parse(subject, body string) *ParsedCommit {
	m := conventionalRe.FindStringSubmatch(subject)
	if m == nil {
		return nil
	}

	commitType := m[1]
	scope := m[2]
	bangBreaking := m[3] == "!"
	description := m[4]

	breaking := bangBreaking || hasBreakingChangeFooter(body)

	bump := angularBumpType(commitType, breaking)

	return &ParsedCommit{
		Type:     commitType,
		Scope:    scope,
		Subject:  description,
		Body:     body,
		Breaking: breaking,
		Bump:     bump,
	}
}

func angularBumpType(commitType string, breaking bool) version.BumpType {
	if breaking {
		return version.BumpMajor
	}
	switch commitType {
	case "feat":
		return version.BumpMinor
	case "fix", "perf":
		return version.BumpPatch
	default:
		return version.BumpPatch
	}
}
