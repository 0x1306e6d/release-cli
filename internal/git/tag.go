package git

import (
	"fmt"
	"sort"
	"strings"

	"github.com/0x1306e6d/release-cli/internal/version"
)

// ListSemverTags returns all tags that parse as valid semver, sorted descending.
// An optional prefix filters tags (e.g., "cli" matches "cli/v1.0.0").
func ListSemverTags(dir string, prefix ...string) ([]version.Semver, error) {
	var tagPrefix string
	if len(prefix) > 0 && prefix[0] != "" {
		tagPrefix = prefix[0] + "/v"
	}

	args := []string{"tag", "--list"}
	if tagPrefix != "" {
		args = append(args, tagPrefix+"*")
	}
	out, err := run(dir, args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var tags []version.Semver
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		versionStr := line
		if tagPrefix != "" {
			// Only accept tags matching the prefix.
			if !strings.HasPrefix(line, tagPrefix) {
				continue
			}
			// Strip prefix to get the version part (e.g., "cli/v1.0.0" -> "1.0.0").
			versionStr = "v" + strings.TrimPrefix(line, tagPrefix)
		}

		v, err := version.Parse(versionStr)
		if err != nil {
			continue // skip non-semver tags
		}
		tags = append(tags, v)
	}

	// Sort descending by version.
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Major != tags[j].Major {
			return tags[i].Major > tags[j].Major
		}
		if tags[i].Minor != tags[j].Minor {
			return tags[i].Minor > tags[j].Minor
		}
		return tags[i].Patch > tags[j].Patch
	})

	return tags, nil
}

// LatestSemverTag returns the highest semver tag, or 0.0.0 if none exist.
// An optional prefix filters tags (e.g., "cli" matches "cli/v1.0.0").
func LatestSemverTag(dir string, prefix ...string) (version.Semver, error) {
	tags, err := ListSemverTags(dir, prefix...)
	if err != nil {
		return version.Semver{}, err
	}
	if len(tags) == 0 {
		return version.Semver{Major: 0, Minor: 0, Patch: 0}, nil
	}
	return tags[0], nil
}

// NamespacedTagString returns a tag string with the given prefix.
// e.g., NamespacedTagString("cli", v) returns "cli/v1.2.0".
// If prefix is empty, returns the standard "v1.2.0" format.
func NamespacedTagString(prefix string, v version.Semver) string {
	if prefix == "" {
		return v.TagString()
	}
	return prefix + "/v" + v.CoreString()
}

// CreateTag creates an annotated git tag.
func CreateTag(dir string, tag string, message string) error {
	_, err := run(dir, "tag", "-a", tag, "-m", message)
	if err != nil {
		return fmt.Errorf("creating tag %s: %w", tag, err)
	}
	return nil
}
