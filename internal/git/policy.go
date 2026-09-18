package git

import (
	"fmt"
	"regexp"
)

// ValidateReleaseBranch rejects releases from branches outside the configured pattern.
func ValidateReleaseBranch(dir, pattern string) error {
	if pattern == "" {
		return nil
	}
	branch, err := run(dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return fmt.Errorf("reading current branch: %w", err)
	}
	matched, err := regexp.MatchString(pattern, branch)
	if err != nil {
		return fmt.Errorf("invalid release branch pattern %q: %w", pattern, err)
	}
	if !matched {
		return fmt.Errorf("current branch %q does not match git.branch %q", branch, pattern)
	}
	return nil
}
