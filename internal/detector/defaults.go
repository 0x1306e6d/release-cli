package detector

// DefaultRegistry returns a registry with all built-in detectors.
func DefaultRegistry() *Registry {
	return NewRegistry(
		&GoDetector{},
		&NodeDetector{},
		&PythonDetector{},
		&RustDetector{},
		&JavaGradleDetector{},
		&JavaMavenDetector{},
		&DartDetector{},
		&HelmDetector{},
	)
}
