package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x1306e6d/release-cli/internal/commits"
	"github.com/0x1306e6d/release-cli/internal/version"
)

func TestGenerate_GroupsByType(t *testing.T) {
	parsed := []commits.ParsedCommit{
		{Hash: "aaaaaaa0000000000000000000000000000000aa", Type: "feat", Subject: "add export", Bump: version.BumpMinor},
		{Hash: "bbbbbbb0000000000000000000000000000000bb", Type: "fix", Subject: "null pointer", Bump: version.BumpPatch},
		{Hash: "ccccccc0000000000000000000000000000000cc", Type: "feat", Subject: "add import", Scope: "data", Bump: version.BumpMinor},
		{Hash: "ddddddd0000000000000000000000000000000dd", Type: "feat", Subject: "redesign API", Breaking: true, Bump: version.BumpMajor},
	}

	entry := Generate("1.4.0", parsed, nil)

	if len(entry.Groups["Features"]) != 2 {
		t.Errorf("features = %d, want 2", len(entry.Groups["Features"]))
	}
	if len(entry.Groups["Bug Fixes"]) != 1 {
		t.Errorf("bug fixes = %d, want 1", len(entry.Groups["Bug Fixes"]))
	}
	if len(entry.Groups["Breaking Changes"]) != 1 {
		t.Errorf("breaking = %d, want 1", len(entry.Groups["Breaking Changes"]))
	}
}

func TestGenerate_AbbreviatesHash(t *testing.T) {
	parsed := []commits.ParsedCommit{
		{Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", Type: "feat", Subject: "add export", Bump: version.BumpMinor},
	}
	entry := Generate("1.0.0", parsed, nil)
	item := entry.Groups["Features"][0]
	if item.Hash != "a1b2c3d" {
		t.Errorf("hash = %q, want %q", item.Hash, "a1b2c3d")
	}
}

func TestGenerate_WithReferences(t *testing.T) {
	hash := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	parsed := []commits.ParsedCommit{
		{Hash: hash, Type: "feat", Subject: "add export", Bump: version.BumpMinor},
	}
	refs := map[string][]string{
		hash: {"#42", "#15"},
	}
	entry := Generate("1.0.0", parsed, refs)
	item := entry.Groups["Features"][0]
	if len(item.References) != 2 || item.References[0] != "#42" {
		t.Errorf("refs = %v, want [#42 #15]", item.References)
	}
}

func TestRender(t *testing.T) {
	entry := Entry{
		Version: "1.4.0",
		Date:    "2026-03-18",
		Grouped: true,
		Groups: map[string][]Item{
			"Features":  {{Hash: "a1b2c3d", Title: "add export", References: []string{"#42"}}},
			"Bug Fixes": {{Hash: "b2c3d4e", Title: "null pointer"}},
		},
	}

	result := entry.Render()

	if !strings.Contains(result, "## 1.4.0 (2026-03-18)") {
		t.Errorf("missing header in:\n%s", result)
	}
	if !strings.Contains(result, "### Features") {
		t.Errorf("missing Features section in:\n%s", result)
	}
	if !strings.Contains(result, "- a1b2c3d - add export (#42)") {
		t.Errorf("missing feature item in:\n%s", result)
	}
	if !strings.Contains(result, "- b2c3d4e - null pointer") {
		t.Errorf("missing bug fix item in:\n%s", result)
	}
}

func TestRender_Ungrouped_FlatList(t *testing.T) {
	entry := Entry{
		Version: "1.4.0",
		Date:    "2026-03-23",
		Grouped: false,
		Groups: map[string][]Item{
			"Other": {
				{Hash: "a1b2c3d", Title: "Add export feature"},
				{Hash: "b2c3d4e", Title: "Fix login bug", References: []string{"#10"}},
			},
		},
	}

	result := entry.Render()

	if strings.Contains(result, "###") {
		t.Errorf("flat render should not contain ### headings:\n%s", result)
	}
	if !strings.Contains(result, "## 1.4.0 (2026-03-23)") {
		t.Errorf("missing header in:\n%s", result)
	}
	if !strings.Contains(result, "- a1b2c3d - Add export feature") {
		t.Errorf("missing item in:\n%s", result)
	}
	if !strings.Contains(result, "- b2c3d4e - Fix login bug (#10)") {
		t.Errorf("missing item in:\n%s", result)
	}
}

func TestRenderBody_Grouped(t *testing.T) {
	entry := Entry{
		Version: "1.4.0",
		Date:    "2026-03-18",
		Grouped: true,
		Groups: map[string][]Item{
			"Features":  {{Hash: "a1b2c3d", Title: "add export"}},
			"Bug Fixes": {{Hash: "b2c3d4e", Title: "null pointer"}},
		},
	}

	result := entry.RenderBody()

	if strings.Contains(result, "## 1.4.0") {
		t.Errorf("RenderBody should not contain version heading:\n%s", result)
	}
	if !strings.Contains(result, "### Features") {
		t.Errorf("missing Features section in:\n%s", result)
	}
	if !strings.Contains(result, "- a1b2c3d - add export") {
		t.Errorf("missing feature item in:\n%s", result)
	}
}

func TestRenderBody_Ungrouped(t *testing.T) {
	entry := Entry{
		Version: "1.4.0",
		Date:    "2026-03-23",
		Grouped: false,
		Groups: map[string][]Item{
			"Other": {
				{Hash: "a1b2c3d", Title: "Add export feature"},
				{Hash: "b2c3d4e", Title: "Fix login bug"},
			},
		},
	}

	result := entry.RenderBody()

	if strings.Contains(result, "## 1.4.0") {
		t.Errorf("RenderBody should not contain version heading:\n%s", result)
	}
	if strings.Contains(result, "###") {
		t.Errorf("ungrouped RenderBody should not contain ### headings:\n%s", result)
	}
	if !strings.Contains(result, "- a1b2c3d - Add export feature") {
		t.Errorf("missing item in:\n%s", result)
	}
	if !strings.Contains(result, "- b2c3d4e - Fix login bug") {
		t.Errorf("missing item in:\n%s", result)
	}
}

func TestItemString_WithRefs(t *testing.T) {
	item := Item{Hash: "a1b2c3d", Title: "Add feature", References: []string{"#42", "#15"}}
	got := item.String()
	want := "a1b2c3d - Add feature (#42, #15)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestItemString_NoRefs(t *testing.T) {
	item := Item{Hash: "a1b2c3d", Title: "Update README"}
	got := item.String()
	want := "a1b2c3d - Update README"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteFile_NewFile(t *testing.T) {
	dir := t.TempDir()
	content := "## 1.0.0\n\n- initial release\n"

	err := WriteFile(dir, "CHANGELOG.md", content)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if !strings.Contains(string(data), "# Changelog") {
		t.Error("expected Changelog header")
	}
	if !strings.Contains(string(data), "## 1.0.0") {
		t.Error("expected version entry")
	}
}

func TestWriteFile_Prepend(t *testing.T) {
	dir := t.TempDir()
	existing := "# Changelog\n\n## 1.0.0\n\n- initial release\n"
	os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(existing), 0644)

	newEntry := "## 1.1.0\n\n- new feature\n"
	err := WriteFile(dir, "CHANGELOG.md", newEntry)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	content := string(data)
	// New entry should come before old entry.
	newIdx := strings.Index(content, "## 1.1.0")
	oldIdx := strings.Index(content, "## 1.0.0")
	if newIdx >= oldIdx {
		t.Error("new entry should appear before old entry")
	}
}
