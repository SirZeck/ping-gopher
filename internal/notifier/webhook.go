package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SirZeck/ping-gopher/internal/validator"
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
	DialContext:         validator.SafeDialContext(10 * time.Second),
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     90 * time.Second,
}

var sharedWebhookClient = &http.Client{
	Timeout:   10 * time.Second,
	Transport: sharedWebhookTransport,
}

// SendWebhookAlert dispatches a JSON POST payload to a configured webhook URL with SSRF validation and retries.
func SendWebhookAlert(webhookURL string, payload WebhookPayload, timeout time.Duration) error {
	if webhookURL == "" {
		return nil
	}

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if err := validator.ValidateSafeURL(webhookURL); err != nil {
		return fmt.Errorf("SSRF protection blocked webhook URL '%s': %w", webhookURL, err)
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

	var lastErr error
	maxRetries := 3
	backoff := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "PingGopher-Notifier/1.0")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return nil
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("webhook responded with HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			resp.Body.Close()
		}

		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("webhook delivery to %s failed after %d attempts: %w", webhookURL, maxRetries, lastErr)
}

type WebhookError string

func (e WebhookError) Error() string {
	return string(e)
}
