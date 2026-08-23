package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SirZeck/ping-gopher/internal/validator"
)

// DiscordWebhookPayload represents a rich Discord Embed notification payload.
type DiscordWebhookPayload struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color"` // Integer color code (0xE74C3C red, 0x2ECC71 green)
	Fields      []DiscordField `json:"fields,omitempty"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordFooter struct {
	Text string `json:"text"`
}

// SendDiscordAlert dispatches a color-coded Discord Embed alert card with SSRF validation.
func SendDiscordAlert(discordWebhookURL string, payload WebhookPayload, timeout time.Duration) error {
	if discordWebhookURL == "" {
		return nil
	}

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// Validate target webhook URL against SSRF
	if err := validator.ValidateSafeURL(discordWebhookURL); err != nil {
		return fmt.Errorf("SSRF validation rejected Discord webhook URL: %w", err)
	}

	embedColor := 0xE74C3C // Red for OUTAGE
	statusTitle := "🚨 OUTAGE DETECTED"
	if payload.Status == "UP" {
		embedColor = 0x2ECC71 // Green for RECOVERY
		statusTitle = "✅ SERVICE RECOVERED"
	}

	discordPayload := DiscordWebhookPayload{
		Username: "PingGopher Monitor",
		Embeds: []DiscordEmbed{
			{
				Title: fmt.Sprintf("%s: %s", statusTitle, payload.MonitorName),
				Color: embedColor,
				Fields: []DiscordField{
					{
						Name:   "Target Name",
						Value:  payload.MonitorName,
						Inline: true,
					},
					{
						Name:   "Status",
						Value:  fmt.Sprintf("`%s`", payload.Status),
						Inline: true,
					},
					{
						Name:   "Target URL",
						Value:  payload.TargetURL,
						Inline: false,
					},
					{
						Name:   "Cause / Details",
						Value:  payload.Cause,
						Inline: false,
					},
				},
				Footer: &DiscordFooter{
					Text: "PingGopher Uptime Monitoring Engine",
				},
				Timestamp: payload.Timestamp,
			},
		},
	}

	bodyBytes, err := json.Marshal(discordPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	req, err := http.NewRequest("POST", discordWebhookURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create Discord HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PingGopher-Notifier/1.1.0")

	resp, err := sharedWebhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch Discord webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook server returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}
