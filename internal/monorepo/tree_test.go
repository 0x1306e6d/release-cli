package monorepo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTree_SimpleMonorepo(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: my-project
project: go
modules:
  - cli
  - lib
`,
		"cli/.release.yaml": "project: go\n",
		"lib/.release.yaml": "project: node\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Name != "my-project" {
		t.Errorf("root name = %q, want %q", root.Name, "my-project")
	}
	if len(root.Children) != 2 {
		t.Fatalf("children count = %d, want 2", len(root.Children))
	}
	if root.Children[0].Name != "cli" {
		t.Errorf("child[0].Name = %q, want %q", root.Children[0].Name, "cli")
	}
	if root.Children[1].Name != "lib" {
		t.Errorf("child[1].Name = %q, want %q", root.Children[1].Name, "lib")
	}
}

func TestLoadTree_NestedHierarchy(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - workflow
`,
		"workflow/.release.yaml": `name: workflow
project: go
modules:
  - sub
`,
		"workflow/sub/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("root children = %d, want 1", len(root.Children))
	}
	workflow := root.Children[0]
	if workflow.Name != "workflow" {
		t.Errorf("workflow.Name = %q, want %q", workflow.Name, "workflow")
	}
	if len(workflow.Children) != 1 {
		t.Fatalf("workflow children = %d, want 1", len(workflow.Children))
	}
	sub := workflow.Children[0]
	if sub.Name != "workflow/sub" {
		t.Errorf("sub.Name = %q, want %q", sub.Name, "workflow/sub")
	}
}

func TestLoadTree_MissingChildConfig(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - missing
`,
	})
	// Create the directory but no .release.yaml
	_ = os.MkdirAll(filepath.Join(dir, "missing"), 0755)

	_, err := LoadTree(dir)
	if err == nil {
		t.Fatal("expected error for missing child config")
	}
}

func TestLoadTree_MissingChildDirectory(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - nonexistent
`,
	})

	_, err := LoadTree(dir)
	if err == nil {
		t.Fatal("expected error for missing child directory")
	}
}

func TestLoadTree_CircularReference(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - child
`,
		"child/.release.yaml": `name: child
project: go
modules:
  - ..
`,
	})

	_, err := LoadTree(dir)
	if err == nil {
		t.Fatal("expected error for circular reference")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention circular, got: %v", err)
	}
}

func TestFlatten(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - a
  - b
`,
		"a/.release.yaml": "project: go\n",
		"b/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	flat := root.Flatten()
	if len(flat) != 3 {
		t.Errorf("flatten count = %d, want 3", len(flat))
	}
}

func TestResolve_Found(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - cli
  - lib
`,
		"cli/.release.yaml": "project: go\n",
		"lib/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodes, err := root.Resolve("cli")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("resolve count = %d, want 1", len(nodes))
	}
	if nodes[0].Name != "cli" {
		t.Errorf("resolved name = %q, want %q", nodes[0].Name, "cli")
	}
}

func TestResolve_ParentIncludesDescendants(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - parent
`,
		"parent/.release.yaml": `name: parent
project: go
modules:
  - child
`,
		"parent/child/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodes, err := root.Resolve("parent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("resolve count = %d, want 2 (parent + child)", len(nodes))
	}
}

func TestResolve_NotFound(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - cli
`,
		"cli/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = root.Resolve("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestTagPrefix(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: my-app
project: go
modules:
  - svc
`,
		"svc/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.TagPrefix() != "my-app" {
		t.Errorf("root tag prefix = %q, want %q", root.TagPrefix(), "my-app")
	}
	if root.Children[0].TagPrefix() != "svc" {
		t.Errorf("child tag prefix = %q, want %q", root.Children[0].TagPrefix(), "svc")
	}
}

func TestLoadTree_ContainerRoot(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `modules:
  - cli
  - lib
`,
		"cli/.release.yaml": "project: go\n",
		"lib/.release.yaml": "project: node\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !root.Config.IsContainer() {
		t.Error("root should be a container")
	}
	if len(root.Children) != 2 {
		t.Fatalf("children count = %d, want 2", len(root.Children))
	}
	flat := root.Flatten()
	if len(flat) != 3 {
		t.Errorf("flatten count = %d, want 3 (container + 2 leaves)", len(flat))
	}
}

func TestLoadTree_NestedContainer(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `modules:
  - app
  - services
`,
		"app/.release.yaml": "project: node\n",
		"services/.release.yaml": `modules:
  - api
  - worker
`,
		"services/api/.release.yaml":    "project: go\n",
		"services/worker/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !root.Config.IsContainer() {
		t.Error("root should be a container")
	}
	services := root.Find("services")
	if services == nil {
		t.Fatal("services node not found")
	}
	if !services.Config.IsContainer() {
		t.Error("services should be a container")
	}
	flat := root.Flatten()
	if len(flat) != 5 {
		t.Errorf("flatten count = %d, want 5 (root + app + services + api + worker)", len(flat))
	}
}

func TestValidateTree_OverlappingSiblings(t *testing.T) {
	dir := setupMonorepo(t, map[string]string{
		".release.yaml": `name: root
project: go
modules:
  - a
`,
		"a/.release.yaml": `name: a
project: go
modules:
  - b
`,
		"a/b/.release.yaml": "project: go\n",
	})

	root, err := LoadTree(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// This tree is valid: a contains a/b but they are parent-child, not siblings.
	if err := ValidateTree(root); err != nil {
		t.Errorf("unexpected error for valid nested tree: %v", err)
	}
}

func setupMonorepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
