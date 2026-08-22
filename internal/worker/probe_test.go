package worker

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pinggopher/ping-gopher/internal/db"
)

func TestExecuteHTTPProbeSuccess(t *testing.T) {
	// Start mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	result := ExecuteHTTPProbe(server.URL, 5*time.Second)

	if !result.IsUp {
		t.Fatalf("Expected probe to be UP, got DOWN with error: %s", result.ErrorMessage)
	}
	if result.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", result.StatusCode)
	}
}

func TestExecuteHTTPProbeFailure(t *testing.T) {
	// Start mock HTTP server returning 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	result := ExecuteHTTPProbe(server.URL, 5*time.Second)

	if result.IsUp {
		t.Fatalf("Expected probe to be DOWN for 500 error, got UP")
	}
	if result.StatusCode != 500 {
		t.Fatalf("Expected status code 500, got %d", result.StatusCode)
	}
}

func TestWorkerEngineProcessHTTPCheck(t *testing.T) {
	testDBPath := "test_worker.db"
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

	user := db.User{Email: "worker_test@pinggopher.com", PasswordHash: "pass"}
	database.Create(&user)

	monitor := db.Monitor{
		UserID: user.ID,
		Name:   "Test Server",
		URL:    server.URL,
		Status: db.StatusPaused,
	}
	database.Create(&monitor)

	engine := NewWorkerEngine(database)

	payload := []byte(`{"monitor_id":"` + monitor.ID.String() + `","target_url":"` + server.URL + `"}`)
	err = engine.ProcessHTTPCheck(payload)
	if err != nil {
		t.Fatalf("ProcessHTTPCheck failed: %v", err)
	}

	// Verify PingLog creation
	var logs []db.PingLog
	database.Find(&logs, "monitor_id = ?", monitor.ID)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 PingLog, got %d", len(logs))
	}
	if logs[0].StatusCode != 200 {
		t.Fatalf("Expected PingLog status code 200, got %d", logs[0].StatusCode)
	}

	// Verify Monitor Status updated to UP
	var updatedMonitor db.Monitor
	database.First(&updatedMonitor, "id = ?", monitor.ID)
	if updatedMonitor.Status != db.StatusUp {
		t.Fatalf("Expected Monitor status UP, got %s", updatedMonitor.Status)
	}
}
