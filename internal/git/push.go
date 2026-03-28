package git

import "fmt"

// Push pushes the current branch and any specified tags to the remote
// in a single command to avoid partial-failure states.
func Push(dir string, tags ...string) error {
	args := []string{"push", "origin", "HEAD"}
	for _, tag := range tags {
		if tag != "" {
			args = append(args, tag)
		}
	}
	if _, err := run(dir, args...); err != nil {
		return fmt.Errorf("pushing to remote: %w", err)
	}
	return nil
}
