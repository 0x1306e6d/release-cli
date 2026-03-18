package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// JavaGradleDetector detects Gradle-based Java projects.
type JavaGradleDetector struct{}

func (d *JavaGradleDetector) Name() string      { return "java-gradle" }
func (d *JavaGradleDetector) Aliases() []string  { return []string{"java"} }

func (d *JavaGradleDetector) Detect(dir string) bool {
	for _, f := range []string{"build.gradle", "build.gradle.kts"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}

var gradleVersionRe = regexp.MustCompile(`(?m)^version\s*=\s*(.+)$`)

func (d *JavaGradleDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "gradle.properties"))
	if err != nil {
		return Version{}, fmt.Errorf("reading gradle.properties: %w", err)
	}
	m := gradleVersionRe.FindSubmatch(data)
	if m == nil {
		return Version{}, fmt.Errorf("no version field in gradle.properties")
	}
	return Version{Raw: string(m[1])}, nil
}

func (d *JavaGradleDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "gradle.properties")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading gradle.properties: %w", err)
	}
	updated := gradleVersionRe.ReplaceAll(data, []byte(fmt.Sprintf("version=%s", v.Raw)))
	return os.WriteFile(path, updated, 0644)
}

func (d *JavaGradleDetector) DefaultPublishTargets() []string {
	return []string{"github"}
}

func (d *JavaGradleDetector) SnapshotSuffix() string { return "SNAPSHOT" }
