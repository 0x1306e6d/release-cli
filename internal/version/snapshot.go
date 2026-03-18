package version

// NextSnapshot returns the next development version after a release.
// It bumps minor and appends the ecosystem-specific suffix.
// For example: release 1.4.0 with suffix "-SNAPSHOT" → 1.5.0-SNAPSHOT.
func NextSnapshot(released Semver, suffix string) Semver {
	next := released.Bump(BumpMinor)
	return next.WithPreRelease(suffix)
}

// NormalizeSnapshotSuffix converts ecosystem suffixes to the hyphen-prefixed
// format used internally. Python uses ".dev0" but we store pre-release
// with a leading hyphen; the detector is responsible for formatting on write.
func NormalizeSnapshotSuffix(suffix string) string {
	if suffix == "" {
		return ""
	}
	// Strip leading "-" or "." for uniform internal representation.
	if suffix[0] == '-' || suffix[0] == '.' {
		return suffix[1:]
	}
	return suffix
}
