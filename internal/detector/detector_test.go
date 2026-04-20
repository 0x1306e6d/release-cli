package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoDetector_Detect(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example"), 0644)

	d := &GoDetector{}
	if !d.Detect(dir) {
		t.Error("expected Go project to be detected")
	}
	if d.Detect(t.TempDir()) {
		t.Error("expected no detection in empty dir")
	}
}

func TestNodeDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test", "version": "1.2.3"}`), 0644)

	d := &NodeDetector{}
	if !d.Detect(dir) {
		t.Fatal("expected Node project")
	}

	v, err := d.ReadVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.Raw != "1.2.3" {
		t.Errorf("version = %q, want %q", v.Raw, "1.2.3")
	}

	err = d.WriteVersion(dir, Version{Raw: "1.3.0"})
	if err != nil {
		t.Fatal(err)
	}

	v2, _ := d.ReadVersion(dir)
	if v2.Raw != "1.3.0" {
		t.Errorf("version after write = %q, want %q", v2.Raw, "1.3.0")
	}
}

func TestPythonDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(`[project]
name = "myproject"
version = "1.0.0"
`), 0644)

	d := &PythonDetector{}
	v, err := d.ReadVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.Raw != "1.0.0" {
		t.Errorf("version = %q, want %q", v.Raw, "1.0.0")
	}

	_ = d.WriteVersion(dir, Version{Raw: "1.1.0"})
	v2, _ := d.ReadVersion(dir)
	if v2.Raw != "1.1.0" {
		t.Errorf("version = %q, want %q", v2.Raw, "1.1.0")
	}
}

func TestRustDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(`[package]
name = "myapp"
version = "0.1.0"
edition = "2021"
`), 0644)

	d := &RustDetector{}
	v, _ := d.ReadVersion(dir)
	if v.Raw != "0.1.0" {
		t.Errorf("version = %q, want %q", v.Raw, "0.1.0")
	}

	_ = d.WriteVersion(dir, Version{Raw: "0.2.0"})
	v2, _ := d.ReadVersion(dir)
	if v2.Raw != "0.2.0" {
		t.Errorf("version = %q, want %q", v2.Raw, "0.2.0")
	}
}

func TestJavaGradleDetector_Detect(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "build.gradle"), []byte(""), 0644)

	d := &JavaGradleDetector{}
	if !d.Detect(dir) {
		t.Error("expected Java-Gradle detection")
	}
}

func TestJavaGradleDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "gradle.properties"), []byte("version=1.0.0-SNAPSHOT\ngroup=com.example\n"), 0644)

	d := &JavaGradleDetector{}
	v, _ := d.ReadVersion(dir)
	if v.Raw != "1.0.0-SNAPSHOT" {
		t.Errorf("version = %q, want %q", v.Raw, "1.0.0-SNAPSHOT")
	}

	_ = d.WriteVersion(dir, Version{Raw: "1.0.0"})
	v2, _ := d.ReadVersion(dir)
	if v2.Raw != "1.0.0" {
		t.Errorf("version = %q, want %q", v2.Raw, "1.0.0")
	}
}

func TestJavaMavenDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	pom := `<?xml version="1.0"?>
<project>
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>myapp</artifactId>
  <version>1.0.0-SNAPSHOT</version>
</project>`
	_ = os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(pom), 0644)

	d := &JavaMavenDetector{}
	v, _ := d.ReadVersion(dir)
	if v.Raw != "1.0.0-SNAPSHOT" {
		t.Errorf("version = %q, want %q", v.Raw, "1.0.0-SNAPSHOT")
	}

	_ = d.WriteVersion(dir, Version{Raw: "1.0.0"})
	v2, _ := d.ReadVersion(dir)
	if v2.Raw != "1.0.0" {
		t.Errorf("version = %q, want %q", v2.Raw, "1.0.0")
	}
}

func TestDartDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: myapp\nversion: 1.0.0\n"), 0644)

	d := &DartDetector{}
	v, _ := d.ReadVersion(dir)
	if v.Raw != "1.0.0" {
		t.Errorf("version = %q, want %q", v.Raw, "1.0.0")
	}

	_ = d.WriteVersion(dir, Version{Raw: "1.1.0"})
	data, _ := os.ReadFile(filepath.Join(dir, "pubspec.yaml"))
	if !strings.Contains(string(data), "version: 1.1.0") {
		t.Errorf("expected updated version in pubspec.yaml")
	}
}

func TestHelmDetector_ReadWriteVersion(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: mychart\nversion: 0.1.0\n"), 0644)

	d := &HelmDetector{}
	v, _ := d.ReadVersion(dir)
	if v.Raw != "0.1.0" {
		t.Errorf("version = %q, want %q", v.Raw, "0.1.0")
	}

	d.WriteVersion(dir, Version{Raw: "0.2.0"})
	data, _ := os.ReadFile(filepath.Join(dir, "Chart.yaml"))
	if !strings.Contains(string(data), "version: 0.2.0") {
		t.Errorf("expected updated version in Chart.yaml")
	}
}

func TestDefaultRegistry_AllDetectors(t *testing.T) {
	r := DefaultRegistry()
	expected := []string{"go", "node", "python", "rust", "java-gradle", "java-maven", "dart", "helm"}
	names := r.Names()
	if len(names) != len(expected) {
		t.Fatalf("got %d detectors, want %d: %v", len(names), len(expected), names)
	}
	for _, name := range expected {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing detector: %s", name)
		}
	}
}
