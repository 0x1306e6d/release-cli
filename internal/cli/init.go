package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x1306e6d/release-cli/internal/config"
	"github.com/0x1306e6d/release-cli/internal/detector"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .release.yaml for the current project",
	Long: `Detect the project type from manifest files and generate a .release.yaml
configuration file with sensible defaults.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check if config already exists.
	configPath := filepath.Join(dir, config.ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%s already exists. Edit it manually or delete it first.", config.ConfigFileName)
	}

	// Detect project type.
	registry := detector.DefaultRegistry()
	detectedName, snapshot := detectProject(dir, registry)

	if detectedName == "" {
		return fmt.Errorf("could not detect project type. Create %s manually with:\n  project: <type>\n\nSupported types: %s",
			config.ConfigFileName, strings.Join(registry.Names(), ", "))
	}

	// Generate config.
	yaml := generateConfig(detectedName, snapshot)
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("✓ Created %s (project: %s)\n", config.ConfigFileName, detectedName)
	if snapshot {
		fmt.Println("  Snapshot mode enabled (detected pre-release version in manifest)")
	}
	return nil
}

func detectProject(dir string, registry *detector.Registry) (string, bool) {
	for _, name := range registry.Names() {
		det, err := registry.Resolve(name, dir)
		if err != nil {
			continue
		}
		if det.Detect(dir) {
			// Check for snapshot in current version.
			snapshot := false
			v, err := det.ReadVersion(dir)
			if err == nil && v.Raw != "" {
				if strings.Contains(v.Raw, "SNAPSHOT") ||
					strings.Contains(v.Raw, ".dev") ||
					strings.Contains(v.Raw, "-rc") ||
					strings.Contains(v.Raw, "-alpha") ||
					strings.Contains(v.Raw, "-beta") {
					snapshot = true
				}
			}
			return det.Name(), snapshot
		}
	}
	return "", false
}

func generateConfig(project string, snapshot bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("project: %s\n", project))

	b.WriteString("\n# version:\n")
	b.WriteString("#   scheme: semver\n")
	if snapshot {
		b.WriteString("version:\n  snapshot: true\n")
	} else {
		b.WriteString("#   snapshot: false\n")
	}

	b.WriteString("\n# commits:\n")
	b.WriteString("#   convention: conventional  # conventional, angular, custom\n")

	b.WriteString("\n# changelog:\n")
	b.WriteString("#   enabled: true\n")
	b.WriteString("#   file: CHANGELOG.md\n")

	b.WriteString("\n# propagate:\n")
	b.WriteString("#   - file: Dockerfile\n")
	b.WriteString("#     type: docker-label\n")

	b.WriteString("\n# hooks:\n")
	b.WriteString("#   pre-bump: \"\"\n")
	b.WriteString("#   post-bump: \"\"\n")
	b.WriteString("#   pre-publish: \"\"\n")
	b.WriteString("#   post-publish: \"\"\n")

	b.WriteString("\n# publish:\n")
	b.WriteString("#   github:\n")
	b.WriteString("#     enabled: true\n")
	b.WriteString("#     draft: false\n")
	b.WriteString("#     artifacts: []\n")

	b.WriteString("\n# notify:\n")
	b.WriteString("#   slack:\n")
	b.WriteString("#     webhook: ${SLACK_WEBHOOK_URL}\n")
	b.WriteString("#   webhook:\n")
	b.WriteString("#     url: ${WEBHOOK_URL}\n")

	return b.String()
}
