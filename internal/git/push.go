package git

import "fmt"

// Push pushes the current branch and any specified tags to the remote
// in a single command to avoid partial-failure states.
func Push(dir string, tags ...string) error {
	return PushWithOptions(dir, "origin", nil, tags...)
}

// PushWithOptions pushes the current branch and tags according to a release policy.
func PushWithOptions(dir, remote string, options []string, tags ...string) error {
	args := append([]string{"push"}, options...)
	args = append(args, remote, "HEAD")
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
