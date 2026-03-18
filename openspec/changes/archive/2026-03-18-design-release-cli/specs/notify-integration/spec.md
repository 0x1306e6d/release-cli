## ADDED Requirements

### Requirement: Slack notification support
The system SHALL support sending release notifications to Slack via webhook URL.

#### Scenario: Send Slack notification
- **WHEN** `notify.slack.webhook` is configured with a valid webhook URL
- **AND** a release completes successfully
- **THEN** a message is sent to the configured Slack webhook with the version number and changelog summary

#### Scenario: Slack notification with channel override
- **WHEN** `notify.slack.channel` is configured
- **THEN** the notification is sent to the specified channel

### Requirement: Generic webhook notification support
The system SHALL support sending release notifications to arbitrary HTTP endpoints via webhook.

#### Scenario: Send webhook notification
- **WHEN** `notify.webhook.url` is configured
- **AND** a release completes successfully
- **THEN** a JSON payload is POSTed to the URL containing version, project name, changelog, and release metadata

### Requirement: Notification failure does not fail the release
Notification failures SHALL be reported as warnings but SHALL NOT cause the release pipeline to fail, since the release (version bump, tag, publish) has already completed.

#### Scenario: Slack webhook unreachable
- **WHEN** the Slack webhook URL is unreachable
- **THEN** the system reports a warning but the release is considered successful

### Requirement: Notification is the final pipeline step
Notifications SHALL execute after all other pipeline steps (including publish) have completed successfully.

#### Scenario: Notification after publish
- **WHEN** a release is executed with both publish and notify configured
- **THEN** notifications are sent only after publishing completes successfully
