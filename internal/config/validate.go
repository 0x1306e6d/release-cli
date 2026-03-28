package config

import (
	"fmt"
	"strings"
)

var (
	validSchemes     = []string{"semver"}
	validConventions = []string{"conventional", "angular", "custom"}
)

// validate checks the config for errors.
func (c *Config) validate() error {
	var errs []string

	if c.Project == "" {
		errs = append(errs, "project is required")
	}

	if len(c.Modules) > 0 && c.Name == "" {
		errs = append(errs, `"name" is required when "modules" is declared`)
	}

	if !isValidEnum(c.Version.Scheme, validSchemes) {
		errs = append(errs, fmt.Sprintf("invalid version scheme %q (valid: %s)", c.Version.Scheme, strings.Join(validSchemes, ", ")))
	}

	if c.Changes.Commits != nil {
		if !isValidEnum(c.Changes.Commits.Convention, validConventions) {
			errs = append(errs, fmt.Sprintf("invalid commit convention %q (valid: %s)", c.Changes.Commits.Convention, strings.Join(validConventions, ", ")))
		}

		if c.Changes.Commits.Convention == "custom" {
			if len(c.Changes.Commits.Types.Major) == 0 && len(c.Changes.Commits.Types.Minor) == 0 && len(c.Changes.Commits.Types.Patch) == 0 {
				errs = append(errs, "custom commit convention requires at least one type mapping in changes.commits.types")
			}
		}
	}

	for i, p := range c.Propagate {
		if p.File == "" {
			errs = append(errs, fmt.Sprintf("propagate[%d]: file is required", i))
		}
		if p.Type == "" && p.Field == "" && p.Pattern == "" {
			errs = append(errs, fmt.Sprintf("propagate[%d]: one of type, field, or pattern is required", i))
		}
	}

	if c.Notify.Slack != nil && c.Notify.Slack.Webhook == "" {
		errs = append(errs, "notify.slack: webhook is required")
	}

	if c.Notify.Webhook != nil && c.Notify.Webhook.URL == "" {
		errs = append(errs, "notify.webhook: url is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func isValidEnum(value string, valid []string) bool {
	for _, v := range valid {
		if value == v {
			return true
		}
	}
	return false
}
