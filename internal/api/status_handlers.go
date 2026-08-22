package api

import (
	"net/http"
	"time"

	"github.com/pinggopher/ping-gopher/internal/db"
)

type PublicStatusResponse struct {
	SystemStatus    string                 `json:"system_status"` // "All Systems Operational" or "Partial System Outage"
	TotalMonitors   int                    `json:"total_monitors"`
	UpMonitors      int                    `json:"up_monitors"`
	DownMonitors    int                    `json:"down_monitors"`
	Monitors        []PublicMonitorCard    `json:"monitors"`
	ActiveIncidents []PublicIncidentReport `json:"active_incidents"`
}

type PublicMonitorCard struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	CheckIntervalSeconds int              `json:"check_interval_seconds"`
	Status               db.MonitorStatus `json:"status"`
	SSLExpirationDate    *time.Time       `json:"ssl_expiration_date,omitempty"`
}

type PublicIncidentReport struct {
	ID          string            `json:"id"`
	MonitorName string            `json:"monitor_name"`
	StartedAt   time.Time         `json:"started_at"`
	Cause       string            `json:"cause"`
	Status      db.IncidentStatus `json:"status"`
}

// PublicStatusHandler returns an unauthenticated system operational summary for public status pages.
func (h *APIHandler) PublicStatusHandler(w http.ResponseWriter, r *http.Request) {
	var monitors []db.Monitor
	if err := h.DB.Where("status != ?", db.StatusPaused).Find(&monitors).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to query system status")
		return
	}

	upCount := 0
	downCount := 0
	publicMonitors := make([]PublicMonitorCard, 0, len(monitors))

	for _, m := range monitors {
		if m.Status == db.StatusUp {
			upCount++
		} else if m.Status == db.StatusDown {
			downCount++
		}

		publicMonitors = append(publicMonitors, PublicMonitorCard{
			ID:                   m.ID.String(),
			Name:                 m.Name,
			CheckIntervalSeconds: m.CheckIntervalSeconds,
			Status:               m.Status,
			SSLExpirationDate:    m.SSLExpirationDate,
		})
	}

	systemStatus := "All Systems Operational"
	if downCount > 0 {
		systemStatus = "Partial System Outage"
	}

	// Fetch active incidents (OPEN or INVESTIGATING)
	var incidents []db.Incident
	h.DB.Preload("Monitor").Where("status != ?", db.IncidentResolved).Order("started_at desc").Find(&incidents)

	publicIncidents := make([]PublicIncidentReport, 0, len(incidents))
	for _, inc := range incidents {
		monitorName := "Target Monitor"
		if inc.Monitor.Name != "" {
			monitorName = inc.Monitor.Name
		}
		publicIncidents = append(publicIncidents, PublicIncidentReport{
			ID:          inc.ID.String(),
			MonitorName: monitorName,
			StartedAt:   inc.StartedAt,
			Cause:       inc.Cause,
			Status:      inc.Status,
		})
	}

	JSONResponse(w, http.StatusOK, PublicStatusResponse{
		SystemStatus:    systemStatus,
		TotalMonitors:   len(monitors),
		UpMonitors:      upCount,
		DownMonitors:    downCount,
		Monitors:        publicMonitors,
		ActiveIncidents: publicIncidents,
	})
}
