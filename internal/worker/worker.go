package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/SirZeck/ping-gopher/internal/db"
	"github.com/SirZeck/ping-gopher/internal/notifier"
	"gorm.io/gorm"
)

// CheckPayload represents task parameter data passed into probe workers.
type CheckPayload struct {
	MonitorID string `json:"monitor_id"`
	TargetURL string `json:"target_url"`
}

// WorkerEngine processes synthetic probes and updates database telemetry.
type WorkerEngine struct {
	DB       *gorm.DB
	Notifier *notifier.NotificationEngine
}

// NewWorkerEngine creates a new WorkerEngine instance with alerting capabilities.
func NewWorkerEngine(database *gorm.DB, notifierEngine *notifier.NotificationEngine) *WorkerEngine {
	return &WorkerEngine{
		DB:       database,
		Notifier: notifierEngine,
	}
}

// ProcessHTTPCheck executes an HTTP uptime check for a monitor and handles incident state transitions.
func (w *WorkerEngine) ProcessHTTPCheck(payloadRaw []byte) error {
	var payload CheckPayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return fmt.Errorf("failed to parse check payload: %w", err)
	}

	monitorID, err := uuid.Parse(payload.MonitorID)
	if err != nil {
		return fmt.Errorf("invalid monitor ID '%s': %w", payload.MonitorID, err)
	}

	// 1. Fetch monitor from database
	var monitor db.Monitor
	if err := w.DB.First(&monitor, "id = ?", monitorID).Error; err != nil {
		return fmt.Errorf("monitor not found in database: %w", err)
	}

	// 2. Execute HTTP Probe
	probeResult := ExecuteHTTPProbe(monitor.URL, 10*time.Second)

	// 3. Record PingLog entry
	pingLog := db.PingLog{
		MonitorID:      monitor.ID,
		StatusCode:     probeResult.StatusCode,
		ResponseTimeMS: probeResult.ResponseTimeMS,
		ErrorMessage:   probeResult.ErrorMessage,
		CreatedAt:      time.Now(),
	}

	if err := w.DB.Create(&pingLog).Error; err != nil {
		return fmt.Errorf("failed to insert ping log: %w", err)
	}

	// 4. Update Monitor Status & Incident Management
	newStatus := db.StatusUp
	if !probeResult.IsUp {
		newStatus = db.StatusDown
	}

	previousStatus := monitor.Status
	monitor.Status = newStatus
	now := time.Now()
	w.DB.Model(&monitor).Updates(map[string]interface{}{
		"status":     newStatus,
		"updated_at": now,
	})

	// State Transition: UP -> DOWN => Create Incident & Dispatch Alert
	if previousStatus != db.StatusDown && newStatus == db.StatusDown {
		incident := db.Incident{
			MonitorID: monitor.ID,
			StartedAt: time.Now(),
			Cause:     probeResult.ErrorMessage,
			Status:    db.IncidentOpen,
		}
		if err := w.DB.Create(&incident).Error; err == nil && w.Notifier != nil {
			w.Notifier.NotifyIncidentCreated(monitor, incident, monitor.WebhookURL)
		}
	}

	// State Transition: DOWN -> UP => Resolve Open Incidents & Dispatch Recovery Alert
	if previousStatus == db.StatusDown && newStatus == db.StatusUp {
		now := time.Now()
		var openInc db.Incident
		if err := w.DB.Where("monitor_id = ? AND status != ?", monitor.ID, db.IncidentResolved).First(&openInc).Error; err == nil {
			openInc.Status = db.IncidentResolved
			openInc.ResolvedAt = &now
			w.DB.Save(&openInc)
			if w.Notifier != nil {
				w.Notifier.NotifyIncidentResolved(monitor, openInc, monitor.WebhookURL)
			}
		}
	}

	return nil
}

// ProcessSSLCheck executes an SSL certificate inspection probe for a monitor.
func (w *WorkerEngine) ProcessSSLCheck(payloadRaw []byte) error {
	var payload CheckPayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return fmt.Errorf("failed to parse check payload: %w", err)
	}

	monitorID, err := uuid.Parse(payload.MonitorID)
	if err != nil {
		return fmt.Errorf("invalid monitor ID '%s': %w", payload.MonitorID, err)
	}

	var monitor db.Monitor
	if err := w.DB.First(&monitor, "id = ?", monitorID).Error; err != nil {
		return fmt.Errorf("monitor not found in database: %w", err)
	}

	sslResult := ExecuteSSLProbe(monitor.URL, 10*time.Second)

	if sslResult.ExpirationDate != nil {
		w.DB.Model(&monitor).Update("ssl_expiration_date", sslResult.ExpirationDate)
	}

	return nil
}
