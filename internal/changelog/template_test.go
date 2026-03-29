package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderCustom_ItemFields(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "changelog.tmpl")
	tmpl := `## {{ .Version }}
{{ range $label, $items := .Groups }}### {{ $label }}
{{ range $items }}- {{ .Hash }} {{ .Title }}{{ if .References }} ({{ range $i, $ref := .References }}{{ if $i }}, {{ end }}{{ $ref }}{{ end }}){{ end }}
{{ end }}{{ end }}`
	if err := os.WriteFile(tmplPath, []byte(tmpl), 0644); err != nil {
		t.Fatal(err)
	}

	entry := Entry{
		Version: "2.0.0",
		Date:    "2026-03-29",
		Grouped: true,
		Groups: map[string][]Item{
			"Features": {
				{Hash: "a1b2c3d", Title: "add export", References: []string{"#42", "#15"}},
				{Hash: "b2c3d4e", Title: "add import"},
			},
		},
	}

	result, err := RenderCustom(entry, tmplPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "## 2.0.0") {
		t.Errorf("missing version header in:\n%s", result)
	}
	if !strings.Contains(result, "- a1b2c3d add export (#42, #15)") {
		t.Errorf("missing item with refs in:\n%s", result)
	}
	if !strings.Contains(result, "- b2c3d4e add import") {
		t.Errorf("missing item without refs in:\n%s", result)
	}
}
