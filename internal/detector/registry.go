package detector

import (
	"fmt"
	"strings"
)

// Registry holds all registered detectors and resolves project identifiers.
type Registry struct {
	detectors []Detector
	byName    map[string]Detector
}

// NewRegistry creates a registry with the given detectors.
func NewRegistry(detectors ...Detector) *Registry {
	r := &Registry{
		detectors: detectors,
		byName:    make(map[string]Detector, len(detectors)),
	}
	for _, d := range detectors {
		r.byName[d.Name()] = d
	}
	return r
}

// Resolve returns the detector for the given project identifier.
// Exact names (e.g., "java-gradle") are looked up directly.
// General identifiers (e.g., "java") scan dir to find which specific detector matches.
func (r *Registry) Resolve(identifier string, dir string) (Detector, error) {
	// Try exact name lookup first.
	if d, ok := r.byName[identifier]; ok {
		return d, nil
	}

	// Treat identifier as a general alias — collect all detectors that claim it.
	var candidates []Detector
	for _, d := range r.detectors {
		for _, alias := range d.Aliases() {
			if alias == identifier {
				candidates = append(candidates, d)
				break
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("unknown project identifier %q (valid: %s)", identifier, r.validIdentifiers())
	}

	// Run detection on each candidate.
	var matched []Detector
	for _, d := range candidates {
		if d.Detect(dir) {
			matched = append(matched, d)
		}
	}

	switch len(matched) {
	case 0:
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = c.Name()
		}
		return nil, fmt.Errorf("project %q: no matching build tool detected (candidates: %s)", identifier, strings.Join(names, ", "))
	case 1:
		return matched[0], nil
	default:
		names := make([]string, len(matched))
		for i, m := range matched {
			names[i] = m.Name()
		}
		return nil, fmt.Errorf("project %q: ambiguous — multiple build tools detected: %s\nSpecify one explicitly (e.g., project: %s)", identifier, strings.Join(names, ", "), names[0])
	}
}

// Names returns all registered detector names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.detectors))
	for _, d := range r.detectors {
		names = append(names, d.Name())
	}
	return names
}

func (r *Registry) validIdentifiers() string {
	seen := make(map[string]bool)
	var ids []string
	for _, d := range r.detectors {
		if !seen[d.Name()] {
			ids = append(ids, d.Name())
			seen[d.Name()] = true
		}
		for _, a := range d.Aliases() {
			if !seen[a] {
				ids = append(ids, a)
				seen[a] = true
			}
		}
	}
	return strings.Join(ids, ", ")
}
