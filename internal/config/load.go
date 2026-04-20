package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigFileName is the expected config file name.
const ConfigFileName = ".release.yaml"

var knownTopLevelKeys = map[string]bool{
	"project":   true,
	"name":      true,
	"modules":   true,
	"version":   true,
	"changes":   true,
	"changelog": true,
	"propagate": true,
	"hooks":     true,
	"publish":   true,
	"notify":    true,
}

var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads and parses the .release.yaml config from the given directory.
// Returns the parsed config, any warnings, and an error if loading fails.
func Load(dir string) (*Config, []string, error) {
	path := filepath.Join(dir, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("config file not found: %s (run 'release-cli init' to create one)", path)
		}
		return nil, nil, fmt.Errorf("reading config: %w", err)
	}

	resolved, err := resolveEnvVars(string(data))
	if err != nil {
		return nil, nil, err
	}

	warnings := checkUnknownKeys([]byte(resolved))

	var cfg Config
	if err := yaml.Unmarshal([]byte(resolved), &cfg); err != nil {
		return nil, nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, warnings, err
	}

	return &cfg, warnings, nil
}

// resolveEnvVars replaces ${ENV_VAR} references with their values.
func resolveEnvVars(input string) (string, error) {
	var missing []string
	result := envVarRegex.ReplaceAllStringFunc(input, func(match string) string {
		varName := envVarRegex.FindStringSubmatch(match)[1]
		value, ok := os.LookupEnv(varName)
		if !ok {
			missing = append(missing, varName)
			return match
		}
		return value
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("undefined environment variables: %s", strings.Join(missing, ", "))
	}
	return result, nil
}

// checkUnknownKeys returns warnings for unrecognized top-level config keys.
func checkUnknownKeys(data []byte) []string {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil
	}
	var warnings []string
	for key := range raw {
		if !knownTopLevelKeys[key] {
			warnings = append(warnings, fmt.Sprintf("warning: unknown config key %q", key))
		}
	}
	return warnings
}
