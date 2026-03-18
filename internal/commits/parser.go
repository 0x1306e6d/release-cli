package commits

import (
	"github.com/0x1306e6d/release-cli/internal/version"
)

// ParsedCommit holds the result of parsing a single commit message.
type ParsedCommit struct {
	Type     string // e.g., "feat", "fix"
	Scope    string // e.g., "auth", "cli"
	Subject  string // the description after type(scope):
	Body     string
	Breaking bool
	Bump     version.BumpType
}

// Convention parses commit messages according to a specific convention.
type Convention interface {
	Parse(subject, body string) *ParsedCommit
}

// Analyze parses commits and returns the highest bump type.
// Returns nil bump type if no releasable changes are found.
func Analyze(commits []RawCommit, conv Convention) ([]ParsedCommit, *version.BumpType) {
	var parsed []ParsedCommit
	var highestBump *version.BumpType

	for _, c := range commits {
		p := conv.Parse(c.Subject, c.Body)
		if p == nil {
			continue
		}
		parsed = append(parsed, *p)

		if highestBump == nil || p.Bump > *highestBump {
			bump := p.Bump
			highestBump = &bump
		}
	}

	return parsed, highestBump
}

// RawCommit is a commit with subject and body ready for parsing.
type RawCommit struct {
	Hash    string
	Subject string
	Body    string
}
