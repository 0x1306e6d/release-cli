package commits

import (
	"regexp"
	"strings"

	"github.com/0x1306e6d/release-cli/internal/version"
)

// conventionalRe matches: type(scope)!: description
var conventionalRe = regexp.MustCompile(`^(\w+)(?:\(([^)]*)\))?(!)?\s*:\s*(.*)$`)

// ConventionalCommits implements the Conventional Commits convention.
type ConventionalCommits struct{}

func (c *ConventionalCommits) Parse(subject, body string) *ParsedCommit {
	m := conventionalRe.FindStringSubmatch(subject)
	if m == nil {
		return nil
	}

	commitType := m[1]
	scope := m[2]
	bangBreaking := m[3] == "!"
	description := m[4]

	breaking := bangBreaking || hasBreakingChangeFooter(body)

	bump := conventionalBumpType(commitType, breaking)

	return &ParsedCommit{
		Type:     commitType,
		Scope:    scope,
		Subject:  description,
		Body:     body,
		Breaking: breaking,
		Bump:     bump,
	}
}

func conventionalBumpType(commitType string, breaking bool) version.BumpType {
	if breaking {
		return version.BumpMajor
	}
	switch commitType {
	case "feat":
		return version.BumpMinor
	default:
		return version.BumpPatch
	}
}

func hasBreakingChangeFooter(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "BREAKING CHANGE:") || strings.HasPrefix(line, "BREAKING-CHANGE:") {
			return true
		}
	}
	return false
}
