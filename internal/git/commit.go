package git

import (
	"fmt"
	"strings"
)

// CommitLog represents a single git commit.
type CommitLog struct {
	Hash    string
	Subject string
	Body    string
}

// LogBetween returns commits between fromRef and toRef (exclusive fromRef).
// If fromRef is empty, returns all commits up to toRef.
// An optional pathFilter restricts results to commits touching files under that path.
func LogBetween(dir string, fromRef, toRef string, pathFilter ...string) ([]CommitLog, error) {
	var rangeSpec string
	if fromRef == "" {
		rangeSpec = toRef
	} else {
		rangeSpec = fromRef + ".." + toRef
	}

	// Use a delimiter to split fields reliably.
	const sep = "---RELEASE_CLI_SEP---"
	format := fmt.Sprintf("%%H%s%%s%s%%b%s", sep, sep, sep)

	args := []string{"log", "--format=" + format, rangeSpec}
	if len(pathFilter) > 0 && pathFilter[0] != "" {
		args = append(args, "--", pathFilter[0])
	}

	out, err := run(dir, args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var commits []CommitLog
	// Each commit produces hash<sep>subject<sep>body<sep>.
	raw := strings.Split(out, sep)
	for i := 0; i+2 < len(raw); i += 3 {
		hash := strings.TrimSpace(raw[i])
		subject := strings.TrimSpace(raw[i+1])
		body := strings.TrimSpace(raw[i+2])

		if hash == "" {
			continue
		}

		commits = append(commits, CommitLog{
			Hash:    hash,
			Subject: subject,
			Body:    body,
		})
	}

	return commits, nil
}

// CreateCommit stages the given files and creates a commit.
func CreateCommit(dir string, message string, files ...string) error {
	return CreateCommitWithOptions(dir, message, nil, files...)
}

// CreateCommitWithOptions stages files and creates a release commit with Git options.
func CreateCommitWithOptions(dir string, message string, options []string, files ...string) error {
	if err := ensureOnlyChangedFiles(dir, files); err != nil {
		return err
	}
	if len(files) > 0 {
		args := append([]string{"add", "--"}, files...)
		if _, err := run(dir, args...); err != nil {
			return fmt.Errorf("staging files: %w", err)
		}
	}
	args := append([]string{"commit"}, options...)
	args = append(args, "-m", message)
	_, err := run(dir, args...)
	if err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}
	return nil
}

func ensureOnlyChangedFiles(dir string, allowed []string) error {
	allowedSet := make(map[string]bool, len(allowed))
	for _, file := range allowed {
		allowedSet[file] = true
	}

	for _, args := range [][]string{{"diff", "--name-only"}, {"diff", "--cached", "--name-only"}, {"ls-files", "--others", "--exclude-standard"}} {
		out, err := run(dir, args...)
		if err != nil {
			return fmt.Errorf("checking release files: %w", err)
		}
		for _, file := range strings.Split(out, "\n") {
			if file != "" && !allowedSet[file] {
				return fmt.Errorf("unexpected release file %q", file)
			}
		}
	}
	return nil
}
