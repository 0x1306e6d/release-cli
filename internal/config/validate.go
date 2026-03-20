package config

import (
	"fmt"
	"strings"
)

var (
	validSchemes     = []string{"semver"}
	validConventions = []string{"conventional", "angular", "freeform", "custom"}
)

// validate checks the config for errors.
func (c *Config) validate() error {
	var errs []string

	if c.Project == "" {
		errs = append(errs, "project is required")
	}

	if !isValidEnum(c.Version.Scheme, validSchemes) {
		errs = append(errs, fmt.Sprintf("invalid version scheme %q (valid: %s)", c.Version.Scheme, strings.Join(validSchemes, ", ")))
	}

	if !isValidEnum(c.Categorize.Convention, validConventions) {
		errs = append(errs, fmt.Sprintf("invalid commit convention %q (valid: %s)", c.Categorize.Convention, strings.Join(validConventions, ", ")))
	}

	if c.Categorize.Convention == "custom" {
		if len(c.Categorize.Types.Major) == 0 && len(c.Categorize.Types.Minor) == 0 && len(c.Categorize.Types.Patch) == 0 {
			errs = append(errs, "custom commit convention requires at least one type mapping in categorize.types")
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
