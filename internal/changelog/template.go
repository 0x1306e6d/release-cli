package changelog

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// RenderCustom renders a changelog entry using a custom Go template file.
func RenderCustom(e Entry, templatePath string) (string, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.New("changelog").Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, e); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return buf.String(), nil
}
