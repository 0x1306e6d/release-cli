package detector

import (
	"strings"
	"testing"
)

// stubDetector is a minimal Detector for testing.
type stubDetector struct {
	name       string
	aliases    []string
	detectFunc func(string) bool
}

func (s *stubDetector) Name() string                          { return s.name }
func (s *stubDetector) Aliases() []string                     { return s.aliases }
func (s *stubDetector) Detect(dir string) bool                { return s.detectFunc(dir) }
func (s *stubDetector) ReadVersion(dir string) (Version, error) { return Version{}, nil }
func (s *stubDetector) WriteVersion(dir string, v Version) error { return nil }
func (s *stubDetector) DefaultPublishTargets() []string       { return nil }
func (s *stubDetector) SnapshotSuffix() string                { return "" }

func TestRegistry_ExactNameLookup(t *testing.T) {
	d := &stubDetector{name: "java-gradle", aliases: []string{"java"}}
	r := NewRegistry(d)

	got, err := r.Resolve("java-gradle", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "java-gradle" {
		t.Errorf("got %q, want %q", got.Name(), "java-gradle")
	}
}

func TestRegistry_GeneralIdentifierResolvesToSpecific(t *testing.T) {
	gradle := &stubDetector{
		name:       "java-gradle",
		aliases:    []string{"java"},
		detectFunc: func(dir string) bool { return dir == "gradle-project" },
	}
	maven := &stubDetector{
		name:       "java-maven",
		aliases:    []string{"java"},
		detectFunc: func(dir string) bool { return false },
	}
	r := NewRegistry(gradle, maven)

	got, err := r.Resolve("java", "gradle-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "java-gradle" {
		t.Errorf("got %q, want %q", got.Name(), "java-gradle")
	}
}

func TestRegistry_AmbiguousDetection(t *testing.T) {
	gradle := &stubDetector{
		name:       "java-gradle",
		aliases:    []string{"java"},
		detectFunc: func(dir string) bool { return true },
	}
	maven := &stubDetector{
		name:       "java-maven",
		aliases:    []string{"java"},
		detectFunc: func(dir string) bool { return true },
	}
	r := NewRegistry(gradle, maven)

	_, err := r.Resolve("java", "/some/dir")
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error should mention ambiguity: %v", err)
	}
}

func TestRegistry_NoBuildToolDetected(t *testing.T) {
	gradle := &stubDetector{
		name:       "java-gradle",
		aliases:    []string{"java"},
		detectFunc: func(dir string) bool { return false },
	}
	r := NewRegistry(gradle)

	_, err := r.Resolve("java", "/some/dir")
	if err == nil {
		t.Fatal("expected error for no matching build tool")
	}
	if !strings.Contains(err.Error(), "no matching build tool") {
		t.Errorf("error should mention no match: %v", err)
	}
}

func TestRegistry_UnknownIdentifier(t *testing.T) {
	r := NewRegistry()

	_, err := r.Resolve("unknown", "/some/dir")
	if err == nil {
		t.Fatal("expected error for unknown identifier")
	}
	if !strings.Contains(err.Error(), "unknown project identifier") {
		t.Errorf("error should mention unknown identifier: %v", err)
	}
}
