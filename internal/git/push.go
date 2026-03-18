package git

import "fmt"

// Push pushes the current branch and the specified tag to the remote.
func Push(dir string, tag string) error {
	if _, err := run(dir, "push", "origin", "HEAD"); err != nil {
		return fmt.Errorf("pushing commits: %w", err)
	}
	if tag != "" {
		if _, err := run(dir, "push", "origin", tag); err != nil {
			return fmt.Errorf("pushing tag %s: %w", tag, err)
		}
	}
	return nil
}
