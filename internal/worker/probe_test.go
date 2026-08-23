package worker

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/SirZeck/ping-gopher/internal/db"
)

func TestExecuteHTTPProbeSuccess(t *testing.T) {
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
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
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
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
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
	testDBPath := filepath.Join(t.TempDir(), "test_worker.db")

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

	user := db.User{Email: "worker_test@pinggopher.com", PasswordHash: "pass"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	monitor := db.Monitor{
		UserID: user.ID,
		Name:   "Test Server",
		URL:    server.URL,
		Status: db.StatusPaused,
	}
	if err := database.Create(&monitor).Error; err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	engine := NewWorkerEngine(database, nil)

	payload := []byte(`{"monitor_id":"` + monitor.ID.String() + `","target_url":"` + server.URL + `"}`)
	err = engine.ProcessHTTPCheck(payload)
	if err != nil {
		t.Fatalf("ProcessHTTPCheck failed: %v", err)
	}

	var logs []db.PingLog
	database.Find(&logs, "monitor_id = ?", monitor.ID)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 PingLog, got %d", len(logs))
	}
	if logs[0].StatusCode != 200 {
		t.Fatalf("Expected PingLog status code 200, got %d", logs[0].StatusCode)
	}

	var updatedMonitor db.Monitor
	database.First(&updatedMonitor, "id = ?", monitor.ID)
	if updatedMonitor.Status != db.StatusUp {
		t.Fatalf("Expected Monitor status UP, got %s", updatedMonitor.Status)
	}
}

func TestExecuteHTTPAssertionProbe(t *testing.T) {
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("System Status: Operational"))
	}))
	defer server.Close()

	// 1. Success case (matching status 200 and keyword "Operational")
	res := ExecuteHTTPAssertionProbe(server.URL, 200, "Operational", 5*time.Second)
	if !res.IsUp {
		t.Fatalf("Expected assertion probe to be UP, got error: %s", res.ErrorMessage)
	}

	// 2. Keyword mismatch failure case
	resFail := ExecuteHTTPAssertionProbe(server.URL, 200, "DATABASE_ERROR", 5*time.Second)
	if resFail.IsUp {
		t.Fatalf("Expected assertion probe to fail on keyword mismatch, got UP")
	}

	// 3. Expected status code mismatch failure case
	resStatusFail := ExecuteHTTPAssertionProbe(server.URL, 201, "Operational", 5*time.Second)
	if resStatusFail.IsUp {
		t.Fatalf("Expected assertion probe to fail on status mismatch, got UP")
	}
}

func TestExecuteTCPProbe(t *testing.T) {
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	// Extract socket address host:port
	addr := server.Listener.Addr().String()
	res := ExecuteTCPProbe(addr, 5*time.Second)
	if !res.IsUp {
		t.Fatalf("Expected TCP probe to %s to be UP, got error: %s", addr, res.ErrorMessage)
	}
}

func TestExecuteDNSProbe(t *testing.T) {
	res := ExecuteDNSProbe("google.com", "A", 5*time.Second)
	if !res.IsUp {
		t.Fatalf("Expected DNS lookup for google.com to be UP, got error: %s", res.ErrorMessage)
	}
}
