package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// WebhookNotifier sends release notifications to a generic HTTP endpoint.
type WebhookNotifier struct {
	URL string
}

func (w *WebhookNotifier) Name() string { return "webhook" }

func (w *WebhookNotifier) Notify(info ReleaseInfo) error {
	payload := map[string]string{
		"version":      info.Version,
		"prev_version": info.PrevVersion,
		"project":      info.Project,
		"changelog":    info.ChangelogBody,
		"tag":          info.TagName,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending webhook notification: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}
