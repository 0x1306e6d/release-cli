package changelog

import (
	"regexp"
	"sort"
)

var refRe = regexp.MustCompile(`#(\d+)`)

// ExtractReferences extracts GitHub-style #N references from the commit
// subject and body. Returns a deduplicated, sorted slice (e.g., ["#3", "#42"]).
func ExtractReferences(subject, body string) []string {
	seen := make(map[string]bool)
	var refs []string

	for _, text := range []string{subject, body} {
		for _, m := range refRe.FindAllString(text, -1) {
			if !seen[m] {
				seen[m] = true
				refs = append(refs, m)
			}
		}
	}

	sort.Strings(refs)
	return refs
}
