package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload represents the JSON notification format dispatched to external webhooks.
type WebhookPayload struct {
	Event       string `json:"event"`        // "incident.open" or "incident.resolved"
	MonitorID   string `json:"monitor_id"`   // Target monitor UUID
	MonitorName string `json:"monitor_name"` // Target monitor name
	TargetURL   string `json:"target_url"`   // Target URL
	Status      string `json:"status"`       // "DOWN" or "UP"
	Cause       string `json:"cause"`        // Outage cause description
	Timestamp   string `json:"timestamp"`    // ISO-8601 timestamp
}

var sharedWebhookTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     90 * time.Second,
}

var sharedWebhookClient = &http.Client{
	Timeout:   10 * time.Second,
	Transport: sharedWebhookTransport,
}

// SendWebhookAlert dispatches a JSON POST payload to a configured webhook URL.
func SendWebhookAlert(webhookURL string, payload WebhookPayload, timeout time.Duration) error {
	if webhookURL == "" {
		return nil
	}

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	client := sharedWebhookClient
	if timeout != 10*time.Second {
		client = &http.Client{
			Timeout:   timeout,
			Transport: sharedWebhookTransport,
		}
	}
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PingGopher-Notifier/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch webhook to %s: %w", webhookURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned non-2xx status code %d", resp.StatusCode)
	}

	return nil
}

type WebhookError string

func (e WebhookError) Error() string {
	return string(e)
}
