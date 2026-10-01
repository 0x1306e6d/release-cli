package config

// Config represents the full .release.yaml configuration.
type Config struct {
	Project   string            `yaml:"project"`
	Name      string            `yaml:"name"`
	Modules   []string          `yaml:"modules"`
	Version   VersionConfig     `yaml:"version"`
	Changes   ChangesConfig     `yaml:"changes"`
	Changelog ChangelogConfig   `yaml:"changelog"`
	Commit    CommitConfig      `yaml:"commit"`
	Git       GitConfig         `yaml:"git"`
	Propagate []PropagateTarget `yaml:"propagate"`
	Hooks     HooksConfig       `yaml:"hooks"`
	Publish   PublishConfig     `yaml:"publish"`
	Notify    NotifyConfig      `yaml:"notify"`
}

// GitConfig controls the Git operations performed for a release.
type GitConfig struct {
	Remote     string   `yaml:"remote"`
	Branch     string   `yaml:"branch"`
	TagFormat  string   `yaml:"tag-format"`
	SignTag    bool     `yaml:"sign-tag"`
	Push       *bool    `yaml:"push"`
	CommitArgs []string `yaml:"commit-args"`
	PushArgs   []string `yaml:"push-args"`
}

// RemoteName returns the configured remote or the legacy origin default.
func (g GitConfig) RemoteName() string {
	if g.Remote == "" {
		return "origin"
	}
	return g.Remote
}

// ReleaseTagFormat returns the configured tag template or the legacy v prefix.
func (g GitConfig) ReleaseTagFormat() string {
	if g.TagFormat == "" {
		return "v{{ .Version }}"
	}
	return g.TagFormat
}

// PushEnabled reports whether release commits and tags should be pushed.
func (g GitConfig) PushEnabled() bool {
	return g.Push == nil || *g.Push
}

// CommitConfig controls release commit contents and messages.
type CommitConfig struct {
	Mode    string   `yaml:"mode"`
	Include []string `yaml:"include"`
	Release string   `yaml:"release"`
	Next    string   `yaml:"next"`
}

// IsMonorepo returns true when the config declares child modules.
func (c *Config) IsMonorepo() bool {
	return len(c.Modules) > 0
}

// IsContainer returns true when the config groups sub-projects without being
// a releasable unit itself (modules declared, project omitted).
func (c *Config) IsContainer() bool {
	return len(c.Modules) > 0 && c.Project == ""
}

// VersionConfig configures version management.
type VersionConfig struct {
	Scheme   string `yaml:"scheme"`
	Snapshot bool   `yaml:"snapshot"`
	Manifest string `yaml:"manifest"`
	Field    string `yaml:"field"`
	Pattern  string `yaml:"pattern"`
}

// ChangesConfig configures the source of release changes.
type ChangesConfig struct {
	Commits *CommitsConfig `yaml:"commits"`
}

// CommitsConfig configures commit-based change detection.
type CommitsConfig struct {
	Convention string            `yaml:"convention"`
	Types      CommitTypesConfig `yaml:"types"`
}

// CommitTypesConfig defines custom type-to-bump mappings.
type CommitTypesConfig struct {
	Major []string `yaml:"major"`
	Minor []string `yaml:"minor"`
	Patch []string `yaml:"patch"`
}

// CommitConventionParams returns the convention name and type mappings
// for commit parsing. When no commits config is present, returns
// empty convention with empty type lists.
func (c *ChangesConfig) CommitConventionParams() (convention string, major, minor, patch []string) {
	if c.Commits == nil {
		return "", nil, nil, nil
	}
	return c.Commits.Convention, c.Commits.Types.Major, c.Commits.Types.Minor, c.Commits.Types.Patch
}

// IsGroupedChangelog returns true when a commit convention is configured,
// meaning the changelog should use grouped rendering.
func (c *ChangesConfig) IsGroupedChangelog() bool {
	return c.Commits != nil
}

// ChangelogConfig configures changelog generation.
type ChangelogConfig struct {
	Enabled  *bool  `yaml:"enabled"`
	File     string `yaml:"file"`
	Template string `yaml:"template"`
}

// PropagateTarget defines a file where the version should be propagated.
type PropagateTarget struct {
	File    string `yaml:"file"`
	Type    string `yaml:"type"`
	Field   string `yaml:"field"`
	Pattern string `yaml:"pattern"`
}

// HooksConfig defines lifecycle hook commands.
type HooksConfig struct {
	PreBump     string `yaml:"pre-bump"`
	PostBump    string `yaml:"post-bump"`
	PrePublish  string `yaml:"pre-publish"`
	PostPublish string `yaml:"post-publish"`
}

// PublishConfig configures publish targets.
type PublishConfig struct {
	GitHub GitHubPublishConfig `yaml:"github"`
}

// GitHubPublishConfig configures GitHub Releases publishing.
type GitHubPublishConfig struct {
	Enabled   *bool    `yaml:"enabled"`
	Draft     bool     `yaml:"draft"`
	Artifacts []string `yaml:"artifacts"`
}

// NotifyConfig configures notification targets.
type NotifyConfig struct {
	Slack   *SlackNotifyConfig   `yaml:"slack"`
	Webhook *WebhookNotifyConfig `yaml:"webhook"`
}

// SlackNotifyConfig configures Slack notifications.
type SlackNotifyConfig struct {
	Webhook string `yaml:"webhook"`
	Channel string `yaml:"channel"`
}

// WebhookNotifyConfig configures generic webhook notifications.
type WebhookNotifyConfig struct {
	URL string `yaml:"url"`
}

func boolPtr(b bool) *bool {
	return &b
}

// applyDefaults sets default values for unset config fields.
func (c *Config) applyDefaults() {
	if c.Version.Scheme == "" {
		c.Version.Scheme = "semver"
	}
	if c.Changes.Commits != nil && c.Changes.Commits.Convention == "" {
		c.Changes.Commits.Convention = "conventional"
	}
	if c.Changelog.Enabled == nil {
		c.Changelog.Enabled = boolPtr(true)
	}
	if c.Changelog.File == "" {
		c.Changelog.File = "CHANGELOG.md"
	}
	if c.Commit.Mode == "" {
		c.Commit.Mode = "full"
	}
	if c.Git.Remote == "" {
		c.Git.Remote = "origin"
	}
	if c.Git.TagFormat == "" {
		c.Git.TagFormat = "v{{ .Version }}"
	}
	if c.Git.Push == nil {
		c.Git.Push = boolPtr(true)
	}
	if c.Publish.GitHub.Enabled == nil {
		c.Publish.GitHub.Enabled = boolPtr(true)
	}
}
