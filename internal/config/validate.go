package config

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validSchemes     = []string{"semver"}
	validConventions = []string{"conventional", "angular", "custom"}
	validCommitModes = []string{"full", "version-only"}
)

// validate checks the config for errors.
func (c *Config) validate() error {
	var errs []string

	if len(c.Modules) == 0 {
		// Single-project config: project is required.
		if c.Project == "" {
			errs = append(errs, "project is required")
		}
	} else if c.Project != "" && c.Name == "" {
		// Releasable monorepo node (has project + modules): name is required as tag prefix.
		errs = append(errs, `"name" is required when "modules" is declared with "project"`)
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

	if !isValidEnum(c.Commit.Mode, validCommitModes) {
		errs = append(errs, fmt.Sprintf("invalid commit mode %q (valid: %s)", c.Commit.Mode, strings.Join(validCommitModes, ", ")))
	}

	if c.Git.Branch != "" {
		if _, err := regexp.Compile(c.Git.Branch); err != nil {
			errs = append(errs, fmt.Sprintf("git.branch must be a valid regular expression: %v", err))
		}
	}
	if strings.Count(c.Git.ReleaseTagFormat(), "{{ .Version }}") != 1 {
		errs = append(errs, `git.tag-format must contain exactly one "{{ .Version }}" placeholder`)
	} else if !validTagFormat(c.Git.ReleaseTagFormat()) {
		errs = append(errs, "git.tag-format renders an invalid Git tag name")
	}
	for _, option := range append(c.Git.CommitArgs, c.Git.PushArgs...) {
		if !strings.HasPrefix(option, "-") {
			errs = append(errs, fmt.Sprintf("git options must start with '-': %q", option))
		}
	}
	for i, file := range c.Commit.Include {
		if file == "" {
			errs = append(errs, fmt.Sprintf("commit.include[%d]: file is required", i))
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

func validTagFormat(format string) bool {
	tag := strings.Replace(format, "{{ .Version }}", "1.2.3", 1)
	return tag != "" && !strings.HasPrefix(tag, ".") && !strings.HasSuffix(tag, ".") &&
		!strings.HasPrefix(tag, "/") && !strings.HasSuffix(tag, "/") &&
		!strings.Contains(tag, "..") && !strings.Contains(tag, "//") && !strings.Contains(tag, "@{") &&
		!strings.ContainsAny(tag, " ~^:?*[\\")
}

func isValidEnum(value string, valid []string) bool {
	for _, v := range valid {
		if value == v {
			return true
		}
	}
	return false
}
