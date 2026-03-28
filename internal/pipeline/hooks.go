package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// HookOptions provides optional context for hook execution.
type HookOptions struct {
	PackageName string
	PackagePath string
}

// RunHook executes a hook command with release context environment variables.
func RunHook(dir, command, newVersion, prevVersion, project string, hookOpts ...HookOptions) error {
	if command == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	env := append(os.Environ(),
		"RELEASE_VERSION="+newVersion,
		"RELEASE_PREV_VERSION="+prevVersion,
		"RELEASE_PROJECT="+project,
	)
	if len(hookOpts) > 0 && hookOpts[0].PackageName != "" {
		env = append(env,
			"RELEASE_PACKAGE="+hookOpts[0].PackageName,
			"RELEASE_PACKAGE_PATH="+hookOpts[0].PackagePath,
		)
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q failed: %v\nstdout: %s\nstderr: %s",
			command, err, stdout.String(), stderr.String())
	}
	return nil
}
