package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0x1306e6d/release-cli/internal/commits"
)

// Item represents a single changelog line item with structured metadata.
type Item struct {
	Hash       string   // abbreviated commit hash (first 7 chars)
	Title      string   // commit subject (after convention parsing)
	References []string // PR numbers and issue refs, e.g., ["#42", "#15"]
}

// Entry holds the data for a single changelog entry.
type Entry struct {
	Version string
	Date    string
	Groups  map[string][]Item // type label → list of items
	Grouped bool              // when true, render with ### group headings
}

// Generate creates a changelog entry from parsed commits.
// refs maps full commit hashes to their resolved references (PR numbers, issue refs).
// When refs is nil, no references are attached to items.
func Generate(version string, parsed []commits.ParsedCommit, refs map[string][]string) Entry {
	groups := make(map[string][]Item)
	for _, c := range parsed {
		label := typeLabel(c.Type, c.Breaking)
		title := c.Subject
		if c.Scope != "" {
			title = fmt.Sprintf("**%s:** %s", c.Scope, title)
		}
		hash := c.Hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		var itemRefs []string
		if refs != nil {
			itemRefs = refs[c.Hash]
		}
		groups[label] = append(groups[label], Item{
			Hash:       hash,
			Title:      title,
			References: itemRefs,
		})
	}
	return Entry{
		Version: version,
		Date:    time.Now().Format("2006-01-02"),
		Groups:  groups,
	}
}

// String formats the item as "<hash> - <title> (<refs>)".
func (item Item) String() string {
	s := item.Hash + " - " + item.Title
	if len(item.References) > 0 {
		s += " (" + strings.Join(item.References, ", ") + ")"
	}
	return s
}

// Render formats the entry using the default template (with version heading).
func (e Entry) Render() string {
	return fmt.Sprintf("## %s (%s)\n", e.Version, e.Date) + e.RenderBody()
}

// RenderBody formats the entry content without the version heading line.
func (e Entry) RenderBody() string {
	var b strings.Builder
	wroteGroup := false

	writeGroup := func(label string, items []Item) {
		if e.Grouped {
			fmt.Fprintf(&b, "\n### %s\n\n", label)
		} else if wroteGroup {
			b.WriteString("\n")
		}
		for _, item := range items {
			fmt.Fprintf(&b, "- %s\n", item.String())
		}
		wroteGroup = true
	}

	// Render groups in a fixed order.
	for _, label := range groupOrder {
		if items, ok := e.Groups[label]; ok {
			writeGroup(label, items)
		}
	}

	// Render any remaining groups not in the fixed order.
	for label, items := range e.Groups {
		if !isInOrder(label) {
			writeGroup(label, items)
		}
	}

	return b.String()
}

// WriteFile prepends the entry to the changelog file, creating it if needed.
func WriteFile(dir, filename, content string) error {
	path := filepath.Join(dir, filename)

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading changelog: %w", err)
	}

	var result string
	if len(existing) > 0 {
		result = content + "\n" + string(existing)
	} else {
		result = "# Changelog\n\n" + content
	}

	return os.WriteFile(path, []byte(result), 0644)
}

var groupOrder = []string{
	"Breaking Changes",
	"Features",
	"Bug Fixes",
	"Other",
}

func isInOrder(label string) bool {
	for _, l := range groupOrder {
		if l == label {
			return true
		}
	}
	return false
}

func typeLabel(commitType string, breaking bool) string {
	if breaking {
		return "Breaking Changes"
	}
	switch commitType {
	case "feat":
		return "Features"
	case "fix":
		return "Bug Fixes"
	default:
		return "Other"
	}
}
