package notifier

import (
	"fmt"
	"time"

	"github.com/pinggopher/ping-gopher/internal/db"
)

// NotificationEngine manages multi-channel alert dispatches for incidents.
type NotificationEngine struct{}

// NewNotificationEngine creates a new NotificationEngine.
func NewNotificationEngine() *NotificationEngine {
	return &NotificationEngine{}
}

// NotifyIncidentCreated triggers real-time alerts when a new outage incident opens.
func (n *NotificationEngine) NotifyIncidentCreated(monitor db.Monitor, incident db.Incident, webhookURL string) {
	payload := WebhookPayload{
		Event:       "incident.open",
		MonitorID:   monitor.ID.String(),
		MonitorName: monitor.Name,
		TargetURL:   monitor.URL,
		Status:      string(db.StatusDown),
		Cause:       incident.Cause,
		Timestamp:   incident.StartedAt.Format(time.RFC3339),
	}

	go func() {
		if err := SendWebhookAlert(webhookURL, payload, 10*time.Second); err != nil {
			fmt.Printf("[NOTIFIER ERROR] Failed to send outage webhook alert: %v\n", err)
		} else {
			fmt.Printf("[NOTIFIER SUCCESS] Outage webhook alert dispatched for monitor '%s'\n", monitor.Name)
		}
	}()
}

// NotifyIncidentResolved triggers real-time alerts when an outage incident resolves.
func (n *NotificationEngine) NotifyIncidentResolved(monitor db.Monitor, incident db.Incident, webhookURL string) {
	resolvedAt := time.Now()
	if incident.ResolvedAt != nil {
		resolvedAt = *incident.ResolvedAt
	}

	payload := WebhookPayload{
		Event:       "incident.resolved",
		MonitorID:   monitor.ID.String(),
		MonitorName: monitor.Name,
		TargetURL:   monitor.URL,
		Status:      string(db.StatusUp),
		Cause:       "Target recovered and is responding with HTTP 200 OK",
		Timestamp:   resolvedAt.Format(time.RFC3339),
	}

	go func() {
		if err := SendWebhookAlert(webhookURL, payload, 10*time.Second); err != nil {
			fmt.Printf("[NOTIFIER ERROR] Failed to send recovery webhook alert: %v\n", err)
		} else {
			fmt.Printf("[NOTIFIER SUCCESS] Recovery webhook alert dispatched for monitor '%s'\n", monitor.Name)
		}
	}()
}
