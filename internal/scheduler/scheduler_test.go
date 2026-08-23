package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/SirZeck/ping-gopher/internal/db"
	"github.com/SirZeck/ping-gopher/internal/worker"
)

func TestSchedulerRunCheckCycle(t *testing.T) {
	testDBPath := filepath.Join(t.TempDir(), "test_scheduler.db")

	database, err := db.InitDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	if sqlDB, err := database.DB(); err == nil {
		defer sqlDB.Close()
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	user := db.User{Email: "sched_test@pinggopher.com", PasswordHash: "pass"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	monitor := db.Monitor{
		UserID:               user.ID,
		Name:                 "Active Target",
		URL:                  server.URL,
		CheckIntervalSeconds: 10,
		Status:               db.StatusUp,
	}
	if err := database.Create(&monitor).Error; err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	workerEngine := worker.NewWorkerEngine(database, nil)
	sched := NewScheduler(database, workerEngine)

	err = sched.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("Scheduler RunOnce failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	var logs []db.PingLog
	database.Find(&logs, "monitor_id = ?", monitor.ID)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 PingLog dispatched by scheduler, got %d", len(logs))
	}
}
