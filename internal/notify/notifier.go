package notify

import "fmt"

// ReleaseInfo holds metadata for notifications.
type ReleaseInfo struct {
	Version       string
	PrevVersion   string
	Project       string
	ChangelogBody string
	TagName       string
}

// Notifier sends a release notification.
type Notifier interface {
	Name() string
	Notify(info ReleaseInfo) error
}

// NotifyAll sends notifications to all notifiers, collecting warnings for failures.
// Notification failures are warnings, not errors (the release already completed).
func NotifyAll(notifiers []Notifier, info ReleaseInfo) []string {
	var warnings []string
	for _, n := range notifiers {
		if err := n.Notify(info); err != nil {
			warnings = append(warnings, fmt.Sprintf("notification %s failed: %v", n.Name(), err))
		}
	}
	return warnings
}
