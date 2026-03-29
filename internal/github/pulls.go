package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"sync"
)

const maxConcurrency = 5

// pullResponse is the JSON shape returned by the GitHub API for associated PRs.
type pullResponse struct {
	Number int    `json:"number"`
	Body   string `json:"body"`
}

// Client queries the GitHub API for commit-related PR metadata.
type Client struct {
	Token   string
	Owner   string
	Repo    string
	BaseURL string // override for testing; defaults to "https://api.github.com"

	HTTPClient *http.Client
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return "https://api.github.com"
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// LookupCommitPRs returns PR and linked issue references for a single commit SHA.
func (c *Client) LookupCommitPRs(sha string) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/pulls", c.baseURL(), c.Owner, c.Repo, sha)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}

	var pulls []pullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pulls); err != nil {
		return nil, err
	}

	return extractRefs(pulls), nil
}

var issueRefRe = regexp.MustCompile(`(?i)(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s+#(\d+)`)

func extractRefs(pulls []pullResponse) []string {
	seen := make(map[string]bool)
	var refs []string

	for _, pr := range pulls {
		ref := fmt.Sprintf("#%d", pr.Number)
		if !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
		for _, m := range issueRefRe.FindAllStringSubmatch(pr.Body, -1) {
			issueRef := "#" + m[1]
			if !seen[issueRef] {
				seen[issueRef] = true
				refs = append(refs, issueRef)
			}
		}
	}

	sort.Strings(refs)
	return refs
}

// ResolveCommitPRs looks up PR metadata for multiple commit SHAs concurrently.
// Returns a map from full SHA to references. On per-commit errors, the commit
// is silently skipped. Only commits with non-empty references are included.
func (c *Client) ResolveCommitPRs(shas []string) map[string][]string {
	result := make(map[string][]string)
	var mu sync.Mutex
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, sha := range shas {
		wg.Add(1)
		go func(sha string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			refs, err := c.LookupCommitPRs(sha)
			if err != nil || len(refs) == 0 {
				return
			}

			mu.Lock()
			result[sha] = refs
			mu.Unlock()
		}(sha)
	}

	wg.Wait()
	return result
}
