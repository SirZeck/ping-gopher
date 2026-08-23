package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Credentials stores local CLI authentication session details.
type Credentials struct {
	ServerURL string `json:"server_url"`
	Email     string `json:"email"`
	Token     string `json:"token"`
}

// ResponseEnvelope represents the standard API response format.
type ResponseEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".pinggopher")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return configDir, nil
}

func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// LoadCredentials loads saved CLI credentials from disk.
func LoadCredentials() (*Credentials, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run 'pinggopher-cli login' first")
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("corrupted credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentials writes CLI credentials to disk.
func SaveCredentials(creds *Credentials) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

// DoAPIRequest sends an HTTP request to the PingGopher API server.
func DoAPIRequest(method, serverURL, endpoint string, body interface{}, token string) (*ResponseEnvelope, error) {
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	url := serverURL + endpoint
	var reqBody io.Reader

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PingGopher-CLI/1.0")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server at %s: %w", url, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var env ResponseEnvelope
	if err := json.Unmarshal(respBytes, &env); err != nil {
		return nil, fmt.Errorf("invalid response from server (HTTP %d): %s", resp.StatusCode, string(respBytes))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !env.Success {
		errMsg := env.Error
		if errMsg == "" {
			errMsg = fmt.Sprintf("HTTP %d error", resp.StatusCode)
		}
		return &env, fmt.Errorf("%s", errMsg)
	}

	return &env, nil
}
