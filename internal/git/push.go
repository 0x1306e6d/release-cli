package git

import "fmt"

// Push pushes the current branch and the specified tag to the remote
// in a single command to avoid partial-failure states.
func Push(dir string, tag string) error {
	args := []string{"push", "origin", "HEAD"}
	if tag != "" {
		args = append(args, tag)
	}
	if _, err := run(dir, args...); err != nil {
		return fmt.Errorf("pushing to remote: %w", err)
	}
	return nil
}
