package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// PythonDetector detects Python projects via pyproject.toml.
type PythonDetector struct{}

func (d *PythonDetector) Name() string      { return "python" }
func (d *PythonDetector) Aliases() []string  { return nil }

func (d *PythonDetector) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "pyproject.toml"))
	return err == nil
}

var pyVersionRe = regexp.MustCompile(`(?m)^version\s*=\s*"([^"]+)"`)

func (d *PythonDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err != nil {
		return Version{}, fmt.Errorf("reading pyproject.toml: %w", err)
	}
	m := pyVersionRe.FindSubmatch(data)
	if m == nil {
		return Version{}, fmt.Errorf("no version field in pyproject.toml")
	}
	return Version{Raw: string(m[1])}, nil
}

func (d *PythonDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading pyproject.toml: %w", err)
	}
	updated := pyVersionRe.ReplaceAll(data, []byte(fmt.Sprintf(`version = "%s"`, v.Raw)))
	return os.WriteFile(path, updated, 0644)
}

func (d *PythonDetector) DefaultPublishTargets() []string {
	return []string{"github", "pypi"}
}

func (d *PythonDetector) SnapshotSuffix() string { return "dev0" }
