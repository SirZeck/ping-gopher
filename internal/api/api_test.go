package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/SirZeck/ping-gopher/internal/config"
	"github.com/SirZeck/ping-gopher/internal/db"
)

func setupTestAPI(t *testing.T) (*httptest.Server, *APIHandler, func()) {
	t.Setenv("PINGGOPHER_ALLOW_LOOPBACK", "true")
	testDBPath := filepath.Join(t.TempDir(), "test_api.db")

	database, err := db.InitDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}

	cfg := &config.Config{
		Role:         "all",
		Port:         "8080",
		DatabasePath: testDBPath,
		JWTSecret:    "test-jwt-secret-key-12345",
	}

	handler := NewAPIHandler(database, cfg)
	ts := httptest.NewServer(handler.SetupRouter())

	cleanup := func() {
		ts.Close()
		if sqlDB, err := database.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return ts, handler, cleanup
}

func TestStaticAssets(t *testing.T) {
	ts, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Test GET /
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for GET /, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test GET /style.css
	resp, err = http.Get(ts.URL + "/style.css")
	if err != nil {
		t.Fatalf("GET /style.css failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for GET /style.css, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	t.Logf("GET /style.css Content-Type: %s, Body length: %d", resp.Header.Get("Content-Type"), len(body))
	resp.Body.Close()

	// Test GET /app.js
	resp, err = http.Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatalf("GET /app.js failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for GET /app.js, got %d", resp.StatusCode)
	}
	t.Logf("GET /app.js Content-Type: %s", resp.Header.Get("Content-Type"))
	resp.Body.Close()
}

func TestSignupAndLoginAPI(t *testing.T) {
	ts, _, cleanup := setupTestAPI(t)
	defer cleanup()

	signupPayload := map[string]string{
		"email":    "tenant@pinggopher.io",
		"password": "SecretPassword123!",
	}
	body, _ := json.Marshal(signupPayload)

	resp, err := http.Post(ts.URL+"/v1/auth/signup", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Signup request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", resp.StatusCode)
	}

	var env ResponseEnvelope
	json.NewDecoder(resp.Body).Decode(&env)

	if !env.Success {
		t.Fatalf("Expected success true, got false")
	}

	loginPayload := map[string]string{
		"email":    "tenant@pinggopher.io",
		"password": "SecretPassword123!",
	}
	body, _ = json.Marshal(loginPayload)

	resp, err = http.Post(ts.URL+"/v1/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestMonitorCRUDAPI(t *testing.T) {
	ts, _, cleanup := setupTestAPI(t)
	defer cleanup()

	signupPayload := map[string]string{
		"email":    "crud_user@pinggopher.io",
		"password": "SecretPassword123!",
	}
	body, _ := json.Marshal(signupPayload)
	resp, _ := http.Post(ts.URL+"/v1/auth/signup", "application/json", bytes.NewBuffer(body))

	var authEnv struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&authEnv)
	token := authEnv.Data.Token
	resp.Body.Close()

	client := &http.Client{}

	createPayload := map[string]interface{}{
		"name":                   "Production API Target",
		"url":                    "https://httpbin.org/status/200",
		"check_interval_seconds": 30,
	}
	body, _ = json.Marshal(createPayload)

	req, _ := http.NewRequest("POST", ts.URL+"/v1/monitors", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Create monitor failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created for monitor, got %d", resp.StatusCode)
	}

	var monitorEnv struct {
		Data db.Monitor `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&monitorEnv)
	monitorID := monitorEnv.Data.ID.String()
	resp.Body.Close()

	req, _ = http.NewRequest("GET", ts.URL+"/v1/monitors", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for list monitors, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest("GET", ts.URL+"/v1/monitors/"+monitorID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for get monitor, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest("DELETE", ts.URL+"/v1/monitors/"+monitorID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for delete monitor, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
