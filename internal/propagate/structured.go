package propagate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// propagateStructured updates a field in a structured file (YAML, JSON, TOML).
func propagateStructured(dir, file, field, newVersion string) error {
	path := filepath.Join(dir, file)
	ext := strings.ToLower(filepath.Ext(file))

	switch ext {
	case ".yaml", ".yml":
		return updateYAMLField(path, field, newVersion)
	case ".json":
		return updateJSONField(path, field, newVersion)
	case ".toml":
		return updateTOMLField(path, field, newVersion)
	default:
		return fmt.Errorf("unsupported structured file type %q (supported: .yaml, .yml, .json, .toml)", ext)
	}
}

func updateYAMLField(path, field, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	if doc.Content == nil || len(doc.Content) == 0 {
		return fmt.Errorf("empty YAML document")
	}

	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("expected YAML mapping at top level")
	}

	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == field {
			mapping.Content[i+1].Value = value
			out, err := yaml.Marshal(&doc)
			if err != nil {
				return err
			}
			return os.WriteFile(path, out, 0644)
		}
	}
	return fmt.Errorf("field %q not found in %s", field, path)
}

func updateJSONField(path, field, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	if _, ok := raw[field]; !ok {
		return fmt.Errorf("field %q not found in %s", field, path)
	}

	v, _ := json.Marshal(value)
	raw[field] = v

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0644)
}

func updateTOMLField(path, field, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Simple line-by-line replacement for top-level TOML fields.
	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, field) {
			rest := strings.TrimPrefix(trimmed, field)
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "=") {
				lines[i] = fmt.Sprintf(`%s = "%s"`, field, value)
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("field %q not found in %s", field, path)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
