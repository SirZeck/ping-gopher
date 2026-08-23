package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/SirZeck/ping-gopher/internal/db"
	"github.com/SirZeck/ping-gopher/internal/worker"
)

type CreateMonitorRequest struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
}

type UpdateMonitorRequest struct {
	Name                 string           `json:"name,omitempty"`
	URL                  string           `json:"url,omitempty"`
	CheckIntervalSeconds int              `json:"check_interval_seconds,omitempty"`
	Status               db.MonitorStatus `json:"status,omitempty"`
}

// CreateMonitorHandler creates a new monitor target for the authenticated tenant.
func (h *APIHandler) CreateMonitorHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.URL == "" {
		JSONError(w, http.StatusBadRequest, "Monitor name and URL are required")
		return
	}

	if err := worker.ValidateSafeURL(req.URL); err != nil {
		JSONError(w, http.StatusBadRequest, fmt.Sprintf("Prohibited monitor URL: %v", err))
		return
	}

	if req.CheckIntervalSeconds <= 0 {
		req.CheckIntervalSeconds = 60
	}

	monitor := db.Monitor{
		UserID:               userID,
		Name:                 req.Name,
		URL:                  req.URL,
		CheckIntervalSeconds: req.CheckIntervalSeconds,
		Status:               db.StatusUp, // Enabled by default on creation
	}

	if err := h.DB.Create(&monitor).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create monitor")
		return
	}

	JSONResponse(w, http.StatusCreated, monitor)
}

// ListMonitorsHandler lists all monitors owned by the authenticated tenant.
func (h *APIHandler) ListMonitorsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var monitors []db.Monitor
	if err := h.DB.Where("user_id = ?", userID).Find(&monitors).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to query monitors")
		return
	}

	JSONResponse(w, http.StatusOK, monitors)
}

// GetMonitorHandler fetches details for a specific monitor.
func (h *APIHandler) GetMonitorHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	monitorID, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	var monitor db.Monitor
	if err := h.DB.Where("id = ? AND user_id = ?", monitorID, userID).First(&monitor).Error; err != nil {
		JSONError(w, http.StatusNotFound, "Monitor not found")
		return
	}

	JSONResponse(w, http.StatusOK, monitor)
}

// UpdateMonitorHandler updates monitor parameters or status.
func (h *APIHandler) UpdateMonitorHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	monitorID, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	var monitor db.Monitor
	if err := h.DB.Where("id = ? AND user_id = ?", monitorID, userID).First(&monitor).Error; err != nil {
		JSONError(w, http.StatusNotFound, "Monitor not found")
		return
	}

	var req UpdateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name != "" {
		monitor.Name = req.Name
	}
	if req.URL != "" {
		monitor.URL = req.URL
	}
	if req.CheckIntervalSeconds > 0 {
		monitor.CheckIntervalSeconds = req.CheckIntervalSeconds
	}
	if req.Status != "" {
		monitor.Status = req.Status
	}

	if err := h.DB.Save(&monitor).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to update monitor")
		return
	}

	JSONResponse(w, http.StatusOK, monitor)
}

// DeleteMonitorHandler removes a monitor target.
func (h *APIHandler) DeleteMonitorHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	monitorID, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	result := h.DB.Where("id = ? AND user_id = ?", monitorID, userID).Delete(&db.Monitor{})
	if result.Error != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete monitor")
		return
	}

	if result.RowsAffected == 0 {
		JSONError(w, http.StatusNotFound, "Monitor not found")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"message": "Monitor deleted successfully"})
}

// GetMonitorLogsHandler returns recent PingLog execution telemetry for a monitor.
func (h *APIHandler) GetMonitorLogsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	monitorID, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	// Verify ownership
	var monitor db.Monitor
	if err := h.DB.Where("id = ? AND user_id = ?", monitorID, userID).First(&monitor).Error; err != nil {
		JSONError(w, http.StatusNotFound, "Monitor not found")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var logs []db.PingLog
	h.DB.Where("monitor_id = ?", monitorID).Order("created_at desc").Limit(limit).Find(&logs)

	JSONResponse(w, http.StatusOK, logs)
}

// GetMonitorIncidentsHandler returns outage incident history for a monitor.
func (h *APIHandler) GetMonitorIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	monitorID, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid monitor ID")
		return
	}

	// Verify ownership
	var monitor db.Monitor
	if err := h.DB.Where("id = ? AND user_id = ?", monitorID, userID).First(&monitor).Error; err != nil {
		JSONError(w, http.StatusNotFound, "Monitor not found")
		return
	}

	var incidents []db.Incident
	h.DB.Where("monitor_id = ?", monitorID).Order("created_at desc").Find(&incidents)

	JSONResponse(w, http.StatusOK, incidents)
}
