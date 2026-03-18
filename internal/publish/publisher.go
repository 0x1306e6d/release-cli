package publish

// ReleaseInfo holds the metadata needed for publishing.
type ReleaseInfo struct {
	TagName      string
	Version      string
	ChangelogBody string
	Project      string
}

// Publisher publishes a release to an external service.
type Publisher interface {
	Name() string
	Publish(info ReleaseInfo) error
}
