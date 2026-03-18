package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// DartDetector detects Dart/Flutter projects via pubspec.yaml.
type DartDetector struct{}

func (d *DartDetector) Name() string      { return "dart" }
func (d *DartDetector) Aliases() []string  { return nil }

func (d *DartDetector) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "pubspec.yaml"))
	return err == nil
}

var pubspecVersionRe = regexp.MustCompile(`(?m)^version:\s*(.+)$`)

func (d *DartDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "pubspec.yaml"))
	if err != nil {
		return Version{}, fmt.Errorf("reading pubspec.yaml: %w", err)
	}
	m := pubspecVersionRe.FindSubmatch(data)
	if m == nil {
		return Version{}, fmt.Errorf("no version field in pubspec.yaml")
	}
	return Version{Raw: string(m[1])}, nil
}

func (d *DartDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "pubspec.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading pubspec.yaml: %w", err)
	}
	updated := pubspecVersionRe.ReplaceAll(data, []byte(fmt.Sprintf("version: %s", v.Raw)))
	return os.WriteFile(path, updated, 0644)
}

func (d *DartDetector) DefaultPublishTargets() []string {
	return []string{"github"}
}

func (d *DartDetector) SnapshotSuffix() string { return "" }
