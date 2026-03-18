package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// RustDetector detects Rust projects via Cargo.toml.
type RustDetector struct{}

func (d *RustDetector) Name() string      { return "rust" }
func (d *RustDetector) Aliases() []string  { return nil }

func (d *RustDetector) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "Cargo.toml"))
	return err == nil
}

var cargoVersionRe = regexp.MustCompile(`(?m)^version\s*=\s*"([^"]+)"`)

func (d *RustDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return Version{}, fmt.Errorf("reading Cargo.toml: %w", err)
	}
	m := cargoVersionRe.FindSubmatch(data)
	if m == nil {
		return Version{}, fmt.Errorf("no version field in Cargo.toml")
	}
	return Version{Raw: string(m[1])}, nil
}

func (d *RustDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading Cargo.toml: %w", err)
	}
	updated := cargoVersionRe.ReplaceAll(data, []byte(fmt.Sprintf(`version = "%s"`, v.Raw)))
	return os.WriteFile(path, updated, 0644)
}

func (d *RustDetector) DefaultPublishTargets() []string {
	return []string{"github", "crates.io"}
}

func (d *RustDetector) SnapshotSuffix() string { return "" }
