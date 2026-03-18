package propagate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// propagatePattern replaces a version in a file using a pattern template.
// The pattern uses {{.Version}} as a placeholder, e.g.:
//
//	'const Version = "{{.Version}}"'
//
// The function builds a regex from the pattern where {{.Version}} is replaced
// with a version-matching group, then substitutes the new version.
func propagatePattern(dir, file, pattern, newVersion string) error {
	path := filepath.Join(dir, file)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Build regex: escape the literal parts, replace {{.Version}} with a capture group.
	parts := strings.SplitN(pattern, "{{.Version}}", 2)
	if len(parts) != 2 {
		return fmt.Errorf("pattern must contain {{.Version}} placeholder")
	}

	escapedBefore := regexp.QuoteMeta(parts[0])
	escapedAfter := regexp.QuoteMeta(parts[1])
	re, err := regexp.Compile(escapedBefore + `[^\s"']+` + escapedAfter)
	if err != nil {
		return fmt.Errorf("building pattern regex: %w", err)
	}

	if !re.Match(data) {
		return fmt.Errorf("pattern %q not found in %s", pattern, file)
	}

	replacement := parts[0] + newVersion + parts[1]
	updated := re.ReplaceAll(data, []byte(replacement))
	return os.WriteFile(path, updated, 0644)
}
