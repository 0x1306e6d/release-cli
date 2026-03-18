package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GitHubPublisher creates GitHub Releases.
type GitHubPublisher struct {
	Token     string
	Owner     string
	Repo      string
	Draft     bool
	Artifacts []string // glob patterns
	Dir       string   // project directory for resolving artifact paths
}

func (g *GitHubPublisher) Name() string { return "github" }

func (g *GitHubPublisher) Publish(info ReleaseInfo) error {
	releaseID, err := g.createRelease(info)
	if err != nil {
		return err
	}

	// Upload artifacts if configured.
	for _, pattern := range g.Artifacts {
		matches, err := filepath.Glob(filepath.Join(g.Dir, pattern))
		if err != nil {
			return fmt.Errorf("globbing artifact pattern %q: %w", pattern, err)
		}
		for _, match := range matches {
			if err := g.uploadAsset(releaseID, match); err != nil {
				return fmt.Errorf("uploading %s: %w", filepath.Base(match), err)
			}
		}
	}

	return nil
}

type ghCreateReleaseReq struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Draft   bool   `json:"draft"`
}

type ghCreateReleaseResp struct {
	ID       int    `json:"id"`
	UploadURL string `json:"upload_url"`
}

func (g *GitHubPublisher) createRelease(info ReleaseInfo) (int, error) {
	payload := ghCreateReleaseReq{
		TagName: info.TagName,
		Name:    info.Version,
		Body:    info.ChangelogBody,
		Draft:   g.Draft,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", g.Owner, g.Repo)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("creating GitHub release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GitHub release creation failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result ghCreateReleaseResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (g *GitHubPublisher) uploadAsset(releaseID int, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	name := filepath.Base(filePath)
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets?name=%s",
		g.Owner, g.Repo, releaseID, name)

	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Content-Type", detectContentType(name))
	req.ContentLength = stat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("asset upload failed (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func detectContentType(name string) string {
	switch {
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return "application/gzip"
	case strings.HasSuffix(name, ".zip"):
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
