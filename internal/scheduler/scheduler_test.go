package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pinggopher/ping-gopher/internal/db"
	"github.com/pinggopher/ping-gopher/internal/worker"
)

func TestSchedulerRunCheckCycle(t *testing.T) {
	testDBPath := "test_scheduler.db"
	defer os.Remove(testDBPath)

	database, err := db.InitDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	user := db.User{Email: "sched_test@pinggopher.com", PasswordHash: "pass"}
	database.Create(&user)

	// Create active monitor
	monitor := db.Monitor{
		UserID:               user.ID,
		Name:                 "Active Target",
		URL:                  server.URL,
		CheckIntervalSeconds: 10,
		Status:               db.StatusUp,
	}
	database.Create(&monitor)

	workerEngine := worker.NewWorkerEngine(database)
	sched := NewScheduler(database, workerEngine)

	err = sched.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("Scheduler RunOnce failed: %v", err)
	}

	// Wait briefly for async worker goroutine
	time.Sleep(200 * time.Millisecond)

	// Verify PingLog entry was generated
	var logs []db.PingLog
	database.Find(&logs, "monitor_id = ?", monitor.ID)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 PingLog dispatched by scheduler, got %d", len(logs))
	}
}
