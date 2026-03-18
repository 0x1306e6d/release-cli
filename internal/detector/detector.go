package detector

// Version holds a parsed version string.
type Version struct {
	Raw string // The original version string (e.g., "1.4.0-SNAPSHOT")
}

// Detector defines how a project type manages its versioning.
type Detector interface {
	// Name returns the canonical identifier (e.g., "java-gradle").
	Name() string

	// Aliases returns general identifiers this detector responds to (e.g., ["java"]).
	Aliases() []string

	// Detect returns true if this detector's build tool is present in dir.
	Detect(dir string) bool

	// ReadVersion reads the current version from the project manifest or git tags.
	ReadVersion(dir string) (Version, error)

	// WriteVersion writes a new version to the project manifest.
	// Detectors for tag-based ecosystems (e.g., Go) may be a no-op.
	WriteVersion(dir string, v Version) error

	// DefaultPublishTargets returns the default publish target names for this ecosystem.
	DefaultPublishTargets() []string

	// SnapshotSuffix returns the ecosystem-specific pre-release suffix
	// (e.g., "-SNAPSHOT" for Java, ".dev0" for Python). Empty if not applicable.
	SnapshotSuffix() string
}
