package commits

import (
	"github.com/0x1306e6d/release-cli/internal/version"
)

// freeformCommits treats every commit as a releasable patch bump.
// Used internally when no commit convention is configured.
type freeformCommits struct{}

// Freeform returns a Convention that accepts all commits as patch bumps.
func Freeform() Convention { return &freeformCommits{} }

func (f *freeformCommits) Parse(subject, body string) *ParsedCommit {
	return &ParsedCommit{
		Type:     "other",
		Subject:  subject,
		Body:     body,
		Breaking: false,
		Bump:     version.BumpPatch,
	}
}
