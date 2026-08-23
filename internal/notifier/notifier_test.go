package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/SirZeck/ping-gopher/internal/db"
)

func TestSendWebhookAlertSuccess(t *testing.T) {
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
	receivedChan := make(chan WebhookPayload, 1)

	// Mock webhook receiver server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode webhook payload: %v", err)
		}

		receivedChan <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	payload := WebhookPayload{
		Event:       "incident.open",
		MonitorID:   uuid.New().String(),
		MonitorName: "Test Target",
		TargetURL:   "https://api.example.com/health",
		Status:      "DOWN",
		Cause:       "HTTP 500 Internal Server Error",
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	err := SendWebhookAlert(server.URL, payload, 5*time.Second)
	if err != nil {
		t.Fatalf("SendWebhookAlert failed: %v", err)
	}

	select {
	case rec := <-receivedChan:
		if rec.Event != "incident.open" {
			t.Fatalf("Expected event 'incident.open', got '%s'", rec.Event)
		}
		if rec.MonitorName != "Test Target" {
			t.Fatalf("Expected monitor name 'Test Target', got '%s'", rec.MonitorName)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for webhook payload dispatch")
	}
}

func TestNotificationEngineOutageAndRecovery(t *testing.T) {
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
	receivedChan := make(chan WebhookPayload, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		json.NewDecoder(r.Body).Decode(&payload)
		receivedChan <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewNotificationEngine()

	monitor := db.Monitor{
		ID:   uuid.New(),
		Name: "Production Database",
		URL:  "https://db.example.com",
	}

	incident := db.Incident{
		ID:        uuid.New(),
		MonitorID: monitor.ID,
		StartedAt: time.Now(),
		Cause:     "Connection refused",
		Status:    db.IncidentOpen,
	}

	// 1. Notify Incident Created
	engine.NotifyIncidentCreated(monitor, incident, server.URL)

	select {
	case rec := <-receivedChan:
		if rec.Event != "incident.open" {
			t.Fatalf("Expected 'incident.open', got '%s'", rec.Event)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for incident created alert")
	}

	// 2. Notify Incident Resolved
	now := time.Now()
	incident.ResolvedAt = &now
	incident.Status = db.IncidentResolved

	engine.NotifyIncidentResolved(monitor, incident, server.URL)

	select {
	case rec := <-receivedChan:
		if rec.Event != "incident.resolved" {
			t.Fatalf("Expected 'incident.resolved', got '%s'", rec.Event)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for incident resolved alert")
	}
}
