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
func LogBetween(dir string, fromRef, toRef string) ([]CommitLog, error) {
	var rangeSpec string
	if fromRef == "" {
		rangeSpec = toRef
	} else {
		rangeSpec = fromRef + ".." + toRef
	}

	// Use a delimiter to split fields reliably.
	const sep = "---RELEASE_CLI_SEP---"
	format := fmt.Sprintf("%%H%s%%s%s%%b%s", sep, sep, sep)

	out, err := run(dir, "log", "--format="+format, rangeSpec)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var commits []CommitLog
	entries := strings.Split(out, sep+"\n")
	// Rebuild from the raw format: each commit produces hash<sep>subject<sep>body<sep>
	// We split on <sep> tokens.
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

	_ = entries // unused, using raw split instead
	return commits, nil
}

// CreateCommit stages the given files and creates a commit.
func CreateCommit(dir string, message string, files ...string) error {
	if len(files) > 0 {
		args := append([]string{"add"}, files...)
		if _, err := run(dir, args...); err != nil {
			return fmt.Errorf("staging files: %w", err)
		}
	}
	_, err := run(dir, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}
	return nil
}
