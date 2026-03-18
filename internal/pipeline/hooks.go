package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// RunHook executes a hook command with release context environment variables.
func RunHook(dir, command, newVersion, prevVersion, project string) error {
	if command == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"RELEASE_VERSION="+newVersion,
		"RELEASE_PREV_VERSION="+prevVersion,
		"RELEASE_PROJECT="+project,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q failed: %v\nstdout: %s\nstderr: %s",
			command, err, stdout.String(), stderr.String())
	}
	return nil
}
