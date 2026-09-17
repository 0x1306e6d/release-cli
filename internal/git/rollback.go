package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Rollback restores release-owned files and removes local release refs until
// the first successful push.
type Rollback struct {
	dir          string
	head         string
	paths        []string
	files        map[string]fileState
	stagedPatch  string
	workingPatch string
	status       map[string]bool
	commit       string
	tags         []string
	pushed       bool
}

type fileState struct {
	exists bool
	data   []byte
	mode   os.FileMode
}

// NewRollback snapshots the release-owned paths before they are changed.
func NewRollback(dir string, files ...string) (*Rollback, error) {
	head, err := run(dir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	stagedPatch, err := run(dir, append([]string{"diff", "--binary", "--cached", "--"}, files...)...)
	if err != nil {
		return nil, err
	}
	workingPatch, err := run(dir, append([]string{"diff", "--binary", "--"}, files...)...)
	if err != nil {
		return nil, err
	}
	status, err := statusRecords(dir)
	if err != nil {
		return nil, err
	}

	r := &Rollback{
		dir:          dir,
		head:         head,
		paths:        append([]string(nil), files...),
		files:        make(map[string]fileState, len(files)),
		stagedPatch:  stagedPatch,
		workingPatch: workingPatch,
		status:       status,
	}
	for _, file := range files {
		if file == "" {
			continue
		}
		if _, err := run(dir, "ls-files", "--error-unmatch", "--", file); err == nil {
			continue
		}
		path := filepath.Join(dir, file)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			r.files[file] = fileState{}
			continue
		}
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		r.files[file] = fileState{exists: true, data: data, mode: info.Mode()}
	}
	return r, nil
}

// RecordCommit marks the current commit as release-cli-owned.
func (r *Rollback) RecordCommit() error {
	commit, err := run(r.dir, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	r.commit = commit
	return nil
}

// RecordTag marks a newly-created local tag as release-cli-owned.
func (r *Rollback) RecordTag(tag string) { r.tags = append(r.tags, tag) }

// MarkPushed disables local rollback. Remote state is never deleted.
func (r *Rollback) MarkPushed() { r.pushed = true }

// Rollback removes release-cli-owned local refs and restores owned paths.
// It returns unrelated changes introduced during the failed run.
func (r *Rollback) Rollback() ([]string, error) {
	if r.pushed {
		return nil, nil
	}

	var errs []error
	if r.commit != "" {
		current, err := run(r.dir, "rev-parse", "HEAD")
		if err != nil {
			errs = append(errs, err)
		} else if current == r.commit {
			if _, err := run(r.dir, "reset", "--mixed", r.head); err != nil {
				errs = append(errs, err)
			}
		}
	}
	for _, tag := range r.tags {
		if _, err := run(r.dir, "tag", "-d", tag); err != nil {
			errs = append(errs, err)
		}
	}
	if err := r.restoreFiles(); err != nil {
		errs = append(errs, err)
	}

	after, err := statusRecords(r.dir)
	if err != nil {
		errs = append(errs, err)
	}
	var unexpected []string
	for record := range after {
		if !r.status[record] {
			unexpected = append(unexpected, record)
		}
	}
	sort.Strings(unexpected)
	return unexpected, errors.Join(errs...)
}

func (r *Rollback) restoreFiles() error {
	var tracked []string
	for _, file := range r.paths {
		if file == "" {
			continue
		}
		if _, err := run(r.dir, "cat-file", "-e", "HEAD:"+file); err == nil {
			tracked = append(tracked, file)
		}
	}
	if len(tracked) > 0 {
		if _, err := run(r.dir, append([]string{"restore", "--source=HEAD", "--staged", "--worktree", "--"}, tracked...)...); err != nil {
			return err
		}
	}
	if r.stagedPatch != "" {
		if err := applyPatch(r.dir, r.stagedPatch, "--index"); err != nil {
			return err
		}
	}
	if r.workingPatch != "" {
		if err := applyPatch(r.dir, r.workingPatch); err != nil {
			return err
		}
	}
	for file, state := range r.files {
		path := filepath.Join(r.dir, file)
		if !state.exists {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, state.data, state.mode.Perm()); err != nil {
			return err
		}
	}
	return nil
}

func applyPatch(dir, patch string, args ...string) error {
	cmd := exec.Command("git", append([]string{"apply"}, args...)...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(patch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git apply: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func statusRecords(dir string) (map[string]bool, error) {
	out, err := run(dir, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	records := make(map[string]bool)
	for _, record := range strings.Split(out, "\n") {
		if record != "" {
			records[record] = true
		}
	}
	return records, nil
}
