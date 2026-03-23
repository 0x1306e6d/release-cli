package config

// Config represents the full .release.yaml configuration.
type Config struct {
	Project   string            `yaml:"project"`
	Version   VersionConfig     `yaml:"version"`
	Changes   ChangesConfig     `yaml:"changes"`
	Changelog ChangelogConfig   `yaml:"changelog"`
	Propagate []PropagateTarget `yaml:"propagate"`
	Hooks     HooksConfig       `yaml:"hooks"`
	Publish   PublishConfig     `yaml:"publish"`
	Notify    NotifyConfig      `yaml:"notify"`
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
	Convention string           `yaml:"convention"`
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
// "freeform" with empty type lists.
func (c *ChangesConfig) CommitConventionParams() (convention string, major, minor, patch []string) {
	if c.Commits == nil {
		return "freeform", nil, nil, nil
	}
	return c.Commits.Convention, c.Commits.Types.Major, c.Commits.Types.Minor, c.Commits.Types.Patch
}

// IsGroupedChangelog returns true when the configured convention supports
// grouped changelog rendering (i.e., not freeform and not absent).
func (c *ChangesConfig) IsGroupedChangelog() bool {
	if c.Commits == nil {
		return false
	}
	return c.Commits.Convention != "freeform"
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
	if c.Publish.GitHub.Enabled == nil {
		c.Publish.GitHub.Enabled = boolPtr(true)
	}
}
