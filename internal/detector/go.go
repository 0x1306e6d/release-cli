package detector

import "os"

// GoDetector detects Go projects. Versions come from git tags (no manifest write).
type GoDetector struct{}

func (d *GoDetector) Name() string      { return "go" }
func (d *GoDetector) Aliases() []string  { return nil }

func (d *GoDetector) Detect(dir string) bool {
	_, err := os.Stat(dir + "/go.mod")
	return err == nil
}

func (d *GoDetector) ReadVersion(dir string) (Version, error) {
	// Go uses git tags — version reading is delegated to the git package.
	return Version{}, nil
}

func (d *GoDetector) WriteVersion(dir string, v Version) error {
	// No manifest to write for Go projects.
	return nil
}

func (d *GoDetector) DefaultPublishTargets() []string {
	return []string{"github"}
}

func (d *GoDetector) SnapshotSuffix() string { return "" }
