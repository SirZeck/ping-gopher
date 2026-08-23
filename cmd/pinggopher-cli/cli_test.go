package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadCredentials(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "credentials.json")

	creds := &Credentials{
		ServerURL: "http://localhost:8080",
		Email:     "cli_user@pinggopher.io",
		Token:     "test-jwt-token-xyz-123",
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal creds: %v", err)
	}

	if err := os.WriteFile(testPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	readData, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read test credentials: %v", err)
	}

	var loaded Credentials
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal loaded creds: %v", err)
	}

	if loaded.Email != creds.Email {
		t.Fatalf("Expected email '%s', got '%s'", creds.Email, loaded.Email)
	}
	if loaded.Token != creds.Token {
		t.Fatalf("Expected token '%s', got '%s'", creds.Token, loaded.Token)
	}
}

func TestDoAPIRequestSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"status": "healthy"}}`))
	}))
	defer ts.Close()

	env, err := DoAPIRequest("GET", ts.URL, "/health", nil, "")
	if err != nil {
		t.Fatalf("DoAPIRequest failed: %v", err)
	}

	if !env.Success {
		t.Fatalf("Expected success true, got false")
	}
}
