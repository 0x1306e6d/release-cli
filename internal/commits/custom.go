package commits

import (
	"github.com/0x1306e6d/release-cli/internal/version"
)

// CustomCommits implements a user-defined commit convention.
type CustomCommits struct {
	MajorTypes map[string]bool
	MinorTypes map[string]bool
	PatchTypes map[string]bool
}

// NewCustomCommits creates a CustomCommits convention from type lists.
func NewCustomCommits(major, minor, patch []string) *CustomCommits {
	c := &CustomCommits{
		MajorTypes: make(map[string]bool),
		MinorTypes: make(map[string]bool),
		PatchTypes: make(map[string]bool),
	}
	for _, t := range major {
		c.MajorTypes[t] = true
	}
	for _, t := range minor {
		c.MinorTypes[t] = true
	}
	for _, t := range patch {
		c.PatchTypes[t] = true
	}
	return c
}

func (c *CustomCommits) Parse(subject, body string) *ParsedCommit {
	m := conventionalRe.FindStringSubmatch(subject)
	if m == nil {
		return nil
	}

	commitType := m[1]
	scope := m[2]
	bangBreaking := m[3] == "!"
	description := m[4]

	breaking := bangBreaking || hasBreakingChangeFooter(body)

	bump := c.bumpType(commitType, breaking)
	if bump == nil {
		return nil // type not recognized in custom mapping
	}

	return &ParsedCommit{
		Type:     commitType,
		Scope:    scope,
		Subject:  description,
		Body:     body,
		Breaking: breaking,
		Bump:     *bump,
	}
}

func (c *CustomCommits) bumpType(commitType string, breaking bool) *version.BumpType {
	if breaking {
		bump := version.BumpMajor
		return &bump
	}
	if c.MajorTypes[commitType] {
		bump := version.BumpMajor
		return &bump
	}
	if c.MinorTypes[commitType] {
		bump := version.BumpMinor
		return &bump
	}
	if c.PatchTypes[commitType] {
		bump := version.BumpPatch
		return &bump
	}
	return nil
}
