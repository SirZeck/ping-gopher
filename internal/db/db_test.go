package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDatabaseInitializationAndModels(t *testing.T) {
	testDBPath := filepath.Join(t.TempDir(), "test_pinggopher.db")

	database, err := InitDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	user := User{
		Email:        "admin@pinggopher.io",
		PasswordHash: "secret_hash",
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Fatalf("Expected auto-generated UUID for User, got Nil")
	}

	sslExpiry := time.Now().AddDate(0, 3, 0)
	monitor := Monitor{
		UserID:               user.ID,
		Name:                 "Production API",
		URL:                  "https://api.example.com/health",
		CheckIntervalSeconds: 30,
		Status:               StatusUp,
		SSLExpirationDate:    &sslExpiry,
	}
	if err := database.Create(&monitor).Error; err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	if monitor.ID == uuid.Nil {
		t.Fatalf("Expected auto-generated UUID for Monitor, got Nil")
	}

	pingLog := PingLog{
		MonitorID:      monitor.ID,
		StatusCode:     200,
		ResponseTimeMS: 45,
		ErrorMessage:   "",
		CreatedAt:      time.Now(),
	}
	if err := database.Create(&pingLog).Error; err != nil {
		t.Fatalf("Failed to create ping log: %v", err)
	}

	incident := Incident{
		MonitorID: monitor.ID,
		StartedAt: time.Now(),
		Cause:     "HTTP 503 Service Unavailable",
		Status:    IncidentOpen,
	}
	if err := database.Create(&incident).Error; err != nil {
		t.Fatalf("Failed to create incident: %v", err)
	}

	var fetchedUser User
	if err := database.Preload("Monitors.PingLogs").Preload("Monitors.Incidents").First(&fetchedUser, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if len(fetchedUser.Monitors) != 1 {
		t.Fatalf("Expected 1 monitor for user, got %d", len(fetchedUser.Monitors))
	}
	if len(fetchedUser.Monitors[0].PingLogs) != 1 {
		t.Fatalf("Expected 1 ping log for monitor, got %d", len(fetchedUser.Monitors[0].PingLogs))
	}
	if len(fetchedUser.Monitors[0].Incidents) != 1 {
		t.Fatalf("Expected 1 incident for monitor, got %d", len(fetchedUser.Monitors[0].Incidents))
	}
}
