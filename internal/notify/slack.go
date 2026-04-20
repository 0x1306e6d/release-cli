package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackNotifier sends release notifications to Slack via webhook.
type SlackNotifier struct {
	WebhookURL string
	Channel    string
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Notify(info ReleaseInfo) error {
	text := fmt.Sprintf("*%s %s released* :rocket:\n%s", info.Project, info.Version, info.ChangelogBody)

	payload := map[string]string{"text": text}
	if s.Channel != "" {
		payload["channel"] = s.Channel
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending Slack notification: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	return nil
}
