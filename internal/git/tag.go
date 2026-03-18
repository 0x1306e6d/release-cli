package git

import (
	"fmt"
	"sort"
	"strings"

	"github.com/0x1306e6d/release-cli/internal/version"
)

// ListSemverTags returns all tags that parse as valid semver, sorted descending.
func ListSemverTags(dir string) ([]version.Semver, error) {
	out, err := run(dir, "tag", "--list")
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
		v, err := version.Parse(line)
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
func LatestSemverTag(dir string) (version.Semver, error) {
	tags, err := ListSemverTags(dir)
	if err != nil {
		return version.Semver{}, err
	}
	if len(tags) == 0 {
		return version.Semver{Major: 0, Minor: 0, Patch: 0}, nil
	}
	return tags[0], nil
}

// CreateTag creates an annotated git tag.
func CreateTag(dir string, tag string, message string) error {
	_, err := run(dir, "tag", "-a", tag, "-m", message)
	if err != nil {
		return fmt.Errorf("creating tag %s: %w", tag, err)
	}
	return nil
}
