package propagate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x1306e6d/release-cli/internal/config"
)

func TestPropagatePattern(t *testing.T) {
	dir := t.TempDir()
	file := "version.go"
	os.WriteFile(filepath.Join(dir, file), []byte(`package version

const Version = "1.3.0"
`), 0644)

	targets := []config.PropagateTarget{
		{File: file, Pattern: `const Version = "{{.Version}}"`},
	}

	if err := Propagate(dir, targets, "1.4.0"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, file))
	if !strings.Contains(string(data), `const Version = "1.4.0"`) {
		t.Errorf("expected updated version, got:\n%s", data)
	}
}

func TestPropagateDockerLabel(t *testing.T) {
	dir := t.TempDir()
	file := "Dockerfile"
	os.WriteFile(filepath.Join(dir, file), []byte(`FROM alpine:3.18
LABEL version="1.3.0"
CMD ["./app"]
`), 0644)

	targets := []config.PropagateTarget{
		{File: file, Type: "docker-label"},
	}

	if err := Propagate(dir, targets, "1.4.0"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, file))
	if !strings.Contains(string(data), `LABEL version="1.4.0"`) {
		t.Errorf("expected updated label, got:\n%s", data)
	}
}

func TestPropagateStructuredYAML(t *testing.T) {
	dir := t.TempDir()
	file := "Chart.yaml"
	os.WriteFile(filepath.Join(dir, file), []byte("apiVersion: v2\nappVersion: 1.3.0\nname: mychart\n"), 0644)

	targets := []config.PropagateTarget{
		{File: file, Field: "appVersion"},
	}

	if err := Propagate(dir, targets, "1.4.0"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, file))
	if !strings.Contains(string(data), "appVersion: 1.4.0") {
		t.Errorf("expected updated appVersion, got:\n%s", data)
	}
}

func TestPropagateStructuredJSON(t *testing.T) {
	dir := t.TempDir()
	file := "config.json"
	os.WriteFile(filepath.Join(dir, file), []byte(`{"version": "1.3.0", "name": "app"}`+"\n"), 0644)

	targets := []config.PropagateTarget{
		{File: file, Field: "version"},
	}

	if err := Propagate(dir, targets, "1.4.0"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, file))
	if !strings.Contains(string(data), `"version": "1.4.0"`) {
		t.Errorf("expected updated version, got:\n%s", data)
	}
}

func TestPropagatePatternNotFound(t *testing.T) {
	dir := t.TempDir()
	file := "version.go"
	os.WriteFile(filepath.Join(dir, file), []byte("package main\n"), 0644)

	targets := []config.PropagateTarget{
		{File: file, Pattern: `const Version = "{{.Version}}"`},
	}

	err := Propagate(dir, targets, "1.4.0")
	if err == nil {
		t.Fatal("expected error for pattern not found")
	}
}
