package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// NodeDetector detects Node.js projects (npm/pnpm/bun).
type NodeDetector struct{}

func (d *NodeDetector) Name() string      { return "node" }
func (d *NodeDetector) Aliases() []string  { return nil }

func (d *NodeDetector) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "package.json"))
	return err == nil
}

func (d *NodeDetector) ReadVersion(dir string) (Version, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return Version{}, fmt.Errorf("reading package.json: %w", err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return Version{}, fmt.Errorf("parsing package.json: %w", err)
	}
	if pkg.Version == "" {
		return Version{}, fmt.Errorf("no version field in package.json")
	}
	return Version{Raw: pkg.Version}, nil
}

func (d *NodeDetector) WriteVersion(dir string, v Version) error {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading package.json: %w", err)
	}

	// Preserve formatting by using json.RawMessage round-trip.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing package.json: %w", err)
	}

	versionJSON, err := json.Marshal(v.Raw)
	if err != nil {
		return err
	}
	raw["version"] = versionJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling package.json: %w", err)
	}
	out = append(out, '\n')

	return os.WriteFile(path, out, 0644)
}

func (d *NodeDetector) DefaultPublishTargets() []string {
	return []string{"github", "npm"}
}

func (d *NodeDetector) SnapshotSuffix() string { return "" }
