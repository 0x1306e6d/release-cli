package config

// Config represents the full .release.yaml configuration.
type Config struct {
	Project   string            `yaml:"project"`
	Version   VersionConfig     `yaml:"version"`
	Categorize CategorizeConfig  `yaml:"categorize"`
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

// CategorizeConfig configures commit categorization.
type CategorizeConfig struct {
	Convention string              `yaml:"convention"`
	Types      CategorizeTypesConfig `yaml:"types"`
}

// CategorizeTypesConfig defines custom type-to-bump mappings.
type CategorizeTypesConfig struct {
	Major []string `yaml:"major"`
	Minor []string `yaml:"minor"`
	Patch []string `yaml:"patch"`
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
	if c.Categorize.Convention == "" {
		c.Categorize.Convention = "conventional"
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
