package git

import (
	"fmt"
	"strings"
)

// RemoteOwnerRepo parses the origin remote URL to extract the repository
// owner and name. It supports both HTTPS and SSH URL formats.
func RemoteOwnerRepo(dir string) (owner, repo string, err error) {
	url, err := run(dir, "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("reading remote URL: %w", err)
	}
	return parseOwnerRepo(url)
}

func parseOwnerRepo(url string) (owner, repo string, err error) {
	// Strip trailing .git suffix.
	url = strings.TrimSuffix(url, ".git")

	// SSH format: git@github.com:owner/repo
	if strings.Contains(url, ":") && !strings.Contains(url, "://") {
		parts := strings.SplitN(url, ":", 2)
		return splitOwnerRepo(parts[1])
	}

	// HTTPS format: https://github.com/owner/repo
	// Strip scheme and host.
	idx := strings.Index(url, "://")
	if idx >= 0 {
		url = url[idx+3:]
	}
	// Remove host portion (e.g., "github.com/").
	slashIdx := strings.Index(url, "/")
	if slashIdx < 0 {
		return "", "", fmt.Errorf("cannot parse remote URL: %s", url)
	}
	return splitOwnerRepo(url[slashIdx+1:])
}

func splitOwnerRepo(path string) (owner, repo string, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("cannot extract owner/repo from path: %s", path)
	}
	return parts[0], parts[1], nil
}
