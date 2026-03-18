package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// HelmDetector detects Helm chart projects via Chart.yaml.
type HelmDetector struct{}

func (d *HelmDetector) Name() string      { return "helm" }
func (d *HelmDetector) Aliases() []string  { return nil }

func (d *HelmDetector) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "Chart.yaml"))
	return err == nil
}

var chartVersionRe = regexp.MustCompile(`(?m)^version:\s*(.+)$`)

func (d *HelmDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "Chart.yaml"))
	if err != nil {
		return Version{}, fmt.Errorf("reading Chart.yaml: %w", err)
	}
	m := chartVersionRe.FindSubmatch(data)
	if m == nil {
		return Version{}, fmt.Errorf("no version field in Chart.yaml")
	}
	return Version{Raw: string(m[1])}, nil
}

func (d *HelmDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "Chart.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading Chart.yaml: %w", err)
	}
	updated := chartVersionRe.ReplaceAll(data, []byte(fmt.Sprintf("version: %s", v.Raw)))
	return os.WriteFile(path, updated, 0644)
}

func (d *HelmDetector) DefaultPublishTargets() []string {
	return []string{"github"}
}

func (d *HelmDetector) SnapshotSuffix() string { return "" }
