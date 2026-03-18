package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// JavaMavenDetector detects Maven-based Java projects.
type JavaMavenDetector struct{}

func (d *JavaMavenDetector) Name() string      { return "java-maven" }
func (d *JavaMavenDetector) Aliases() []string  { return []string{"java"} }

func (d *JavaMavenDetector) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "pom.xml"))
	return err == nil
}

// Matches the first <version>...</version> in pom.xml (project version).
var pomVersionRe = regexp.MustCompile(`(?s)<version>([^<]+)</version>`)

func (d *JavaMavenDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "pom.xml"))
	if err != nil {
		return Version{}, fmt.Errorf("reading pom.xml: %w", err)
	}
	m := pomVersionRe.FindSubmatch(data)
	if m == nil {
		return Version{}, fmt.Errorf("no <version> element in pom.xml")
	}
	return Version{Raw: string(m[1])}, nil
}

func (d *JavaMavenDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "pom.xml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading pom.xml: %w", err)
	}
	// Replace only the first occurrence (project version, not dependency versions).
	loc := pomVersionRe.FindIndex(data)
	if loc == nil {
		return fmt.Errorf("no <version> element in pom.xml")
	}
	replacement := fmt.Sprintf("<version>%s</version>", v.Raw)
	updated := make([]byte, 0, len(data))
	updated = append(updated, data[:loc[0]]...)
	updated = append(updated, replacement...)
	updated = append(updated, data[loc[1]:]...)
	return os.WriteFile(path, updated, 0644)
}

func (d *JavaMavenDetector) DefaultPublishTargets() []string {
	return []string{"github"}
}

func (d *JavaMavenDetector) SnapshotSuffix() string { return "SNAPSHOT" }
