package version

import (
	"fmt"
	"strconv"
	"strings"
)

// BumpType represents the kind of version increment.
type BumpType int

const (
	BumpPatch BumpType = iota
	BumpMinor
	BumpMajor
)

func (b BumpType) String() string {
	switch b {
	case BumpMajor:
		return "major"
	case BumpMinor:
		return "minor"
	case BumpPatch:
		return "patch"
	default:
		return "unknown"
	}
}

// ParseBumpType converts a string to a BumpType.
func ParseBumpType(s string) (BumpType, error) {
	switch strings.ToLower(s) {
	case "major":
		return BumpMajor, nil
	case "minor":
		return BumpMinor, nil
	case "patch":
		return BumpPatch, nil
	default:
		return 0, fmt.Errorf("invalid bump type %q: must be major, minor, or patch", s)
	}
}

// Semver represents a parsed semantic version.
type Semver struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string // e.g., "SNAPSHOT", "dev0", "rc.1"
}

// Parse parses a semver string, stripping an optional leading "v".
func Parse(s string) (Semver, error) {
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return Semver{}, fmt.Errorf("empty version string")
	}

	// Split off pre-release at first hyphen or first non-numeric segment after patch.
	var core, pre string

	// Handle Python-style pre-release (e.g., "1.4.0.dev0") by checking for 4+ dot segments.
	parts := strings.SplitN(s, "-", 2)
	core = parts[0]
	if len(parts) > 1 {
		pre = parts[1]
	}

	segments := strings.Split(core, ".")
	if len(segments) < 3 {
		return Semver{}, fmt.Errorf("invalid semver %q: expected major.minor.patch", s)
	}

	// If there are more than 3 dot-separated segments, treat extras as pre-release.
	// This handles Python-style "1.4.0.dev0".
	if len(segments) > 3 {
		pre = strings.Join(segments[3:], ".")
		if len(parts) > 1 {
			pre += "-" + parts[1]
		}
		segments = segments[:3]
	}

	major, err := strconv.Atoi(segments[0])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid major version %q", segments[0])
	}
	minor, err := strconv.Atoi(segments[1])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid minor version %q", segments[1])
	}
	patch, err := strconv.Atoi(segments[2])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid patch version %q", segments[2])
	}

	return Semver{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: pre,
	}, nil
}

// String returns the version as "major.minor.patch[-prerelease]".
func (v Semver) String() string {
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		return base + "-" + v.PreRelease
	}
	return base
}

// TagString returns the version as "vmajor.minor.patch".
func (v Semver) TagString() string {
	return "v" + v.CoreString()
}

// CoreString returns the version without pre-release (e.g., "1.4.0").
func (v Semver) CoreString() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsPreRelease returns true if the version has a pre-release suffix.
func (v Semver) IsPreRelease() bool {
	return v.PreRelease != ""
}

// Bump returns a new version incremented by the given bump type.
// Pre-release suffix is always stripped.
func (v Semver) Bump(bump BumpType) Semver {
	switch bump {
	case BumpMajor:
		return Semver{Major: v.Major + 1, Minor: 0, Patch: 0}
	case BumpMinor:
		return Semver{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
	case BumpPatch:
		return Semver{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	default:
		return v
	}
}

// WithPreRelease returns a copy with the given pre-release suffix.
func (v Semver) WithPreRelease(pre string) Semver {
	return Semver{Major: v.Major, Minor: v.Minor, Patch: v.Patch, PreRelease: pre}
}

// StripPreRelease returns the version without any pre-release suffix.
func (v Semver) StripPreRelease() Semver {
	return Semver{Major: v.Major, Minor: v.Minor, Patch: v.Patch}
}
