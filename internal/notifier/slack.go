package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SirZeck/ping-gopher/internal/validator"
)

// SlackBlockKitPayload represents a rich Slack Block-Kit notification payload.
type SlackBlockKitPayload struct {
	Text   string       `json:"text"`
	Blocks []SlackBlock `json:"blocks,omitempty"`
}

type SlackBlock struct {
	Type   string          `json:"type"`             // "header", "section", "divider", "context"
	Text   *SlackTextBlock `json:"text,omitempty"`   // Header or section text
	Fields []SlackTextBlock`json:"fields,omitempty"` // Key-value grid fields
}

type SlackTextBlock struct {
	Type string `json:"type"` // "plain_text" or "mrkdwn"
	Text string `json:"text"`
}

// SendSlackAlert dispatches a rich Slack Block-Kit formatted alert payload with SSRF validation.
func SendSlackAlert(slackWebhookURL string, payload WebhookPayload, timeout time.Duration) error {
	if slackWebhookURL == "" {
		return nil
	}

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// Validate target webhook URL against SSRF
	if err := validator.ValidateSafeURL(slackWebhookURL); err != nil {
		return fmt.Errorf("SSRF validation rejected Slack webhook URL: %w", err)
	}

	headerIcon := "🚨"
	statusTitle := "OUTAGE DETECTED"
	if payload.Status == "UP" {
		headerIcon = "✅"
		statusTitle = "SERVICE RECOVERED"
	}

	slackPayload := SlackBlockKitPayload{
		Text: fmt.Sprintf("%s PingGopher Alert: %s is %s", headerIcon, payload.MonitorName, payload.Status),
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackTextBlock{
					Type: "plain_text",
					Text: fmt.Sprintf("%s PingGopher Alert: %s", headerIcon, statusTitle),
				},
			},
			{
				Type: "section",
				Fields: []SlackTextBlock{
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Monitor Target:*\n%s", payload.MonitorName),
					},
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Status:*\n`%s`", payload.Status),
					},
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Target URL:*\n<%s>", payload.TargetURL),
					},
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Timestamp:*\n%s", payload.Timestamp),
					},
				},
			},
			{
				Type: "section",
				Text: &SlackTextBlock{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Cause / Details:*\n>%s", payload.Cause),
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(slackPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", slackWebhookURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create Slack HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PingGopher-Notifier/1.1.0")

	resp, err := sharedWebhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch Slack webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Slack webhook server returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}
