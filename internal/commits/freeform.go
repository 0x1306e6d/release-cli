package commits

import (
	"github.com/0x1306e6d/release-cli/internal/version"
)

// FreeformCommits treats every commit as a releasable patch bump.
// No structured format is required — plain English messages are accepted.
type FreeformCommits struct{}

func (f *FreeformCommits) Parse(subject, body string) *ParsedCommit {
	return &ParsedCommit{
		Type:     "other",
		Subject:  subject,
		Body:     body,
		Breaking: false,
		Bump:     version.BumpPatch,
	}
}
