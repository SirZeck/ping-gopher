package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

const version = "1.1.0"

const cliBanner = `
==================================================
  PingGopher CLI v%s — Uptime Monitoring Control
==================================================
`

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Monitor struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	Status               string `json:"status"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
}

type PingLog struct {
	StatusCode     int       `json:"status_code"`
	ResponseTimeMS int       `json:"response_time_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "signup":
		handleSignup(args)
	case "login":
		handleLogin(args)
	case "status":
		handleStatus(args)
	case "monitor":
		handleMonitor(args)
	case "logs":
		handleLogs(args)
	case "version", "-v", "--version":
		fmt.Printf("PingGopher CLI v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command '%s'. Run 'pinggopher-cli help' for usage.\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(cliBanner, version)
	fmt.Println("Usage: pinggopher-cli <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  signup    Register a new tenant account on a PingGopher server")
	fmt.Println("  login     Authenticate against a PingGopher API server")
	fmt.Println("  status    Query public system operational health status")
	fmt.Println("  monitor   Manage target monitors (list, add, delete)")
	fmt.Println("  logs      View response latency telemetry and execution logs")
	fmt.Println("  version   Display CLI version")
	fmt.Println("\nExamples:")
	fmt.Println("  pinggopher-cli signup --url http://localhost:8080 --email newuser@example.com --password secret")
	fmt.Println("  pinggopher-cli login --url http://localhost:8080 --email admin@example.com --password secret")
	fmt.Println("  pinggopher-cli monitor list")
	fmt.Println("  pinggopher-cli monitor add --name \"Production API\" --url https://api.example.com --interval 30")
	fmt.Println("  pinggopher-cli status")
	fmt.Printf("\nTip: Ensure $(go env GOPATH)/bin (%%USERPROFILE%%\\go\\bin on Windows) is in your system PATH.\n")
}

func handleSignup(args []string) {
	fs := flag.NewFlagSet("signup", flag.ExitOnError)
	serverURL := fs.String("url", "http://localhost:8080", "PingGopher API Server URL")
	email := fs.String("email", "", "Tenant Account Email")
	password := fs.String("password", "", "Tenant Account Password")
	fs.Parse(args)

	if *email == "" || *password == "" {
		fmt.Println("Error: --email and --password flags are required.")
		fs.Usage()
		os.Exit(1)
	}

	signupPayload := map[string]string{
		"email":    *email,
		"password": *password,
	}

	env, err := DoAPIRequest("POST", *serverURL, "/v1/auth/signup", signupPayload, "")
	if err != nil {
		fmt.Printf("[ERROR] Registration failed: %v\n", err)
		os.Exit(1)
	}

	var authData struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	}
	if err := json.Unmarshal(env.Data, &authData); err != nil {
		fmt.Printf("[ERROR] Failed to parse auth response: %v\n", err)
		os.Exit(1)
	}

	creds := &Credentials{
		ServerURL: *serverURL,
		Email:     *email,
		Token:     authData.Token,
	}

	if err := SaveCredentials(creds); err != nil {
		fmt.Printf("[ERROR] Failed to save session credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[SUCCESS] Account created and authenticated successfully as '%s' on %s\n", *email, *serverURL)
	fmt.Println("Session token saved to ~/.pinggopher/credentials.json")
}

func handleLogin(args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	serverURL := fs.String("url", "http://localhost:8080", "PingGopher API Server URL")
	email := fs.String("email", "", "Tenant Account Email")
	password := fs.String("password", "", "Tenant Account Password")
	fs.Parse(args)

	if *email == "" || *password == "" {
		fmt.Println("Error: --email and --password flags are required.")
		fs.Usage()
		os.Exit(1)
	}

	loginPayload := map[string]string{
		"email":    *email,
		"password": *password,
	}

	env, err := DoAPIRequest("POST", *serverURL, "/v1/auth/login", loginPayload, "")
	if err != nil {
		fmt.Printf("[ERROR] Authentication failed: %v\n", err)
		os.Exit(1)
	}

	var authData struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	}
	if err := json.Unmarshal(env.Data, &authData); err != nil {
		fmt.Printf("[ERROR] Failed to parse auth response: %v\n", err)
		os.Exit(1)
	}

	creds := &Credentials{
		ServerURL: *serverURL,
		Email:     *email,
		Token:     authData.Token,
	}

	if err := SaveCredentials(creds); err != nil {
		fmt.Printf("[ERROR] Failed to save session credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[SUCCESS] Authenticated successfully as '%s' on %s\n", *email, *serverURL)
	fmt.Println("Session token saved to ~/.pinggopher/credentials.json")
}

func handleStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	serverURL := fs.String("url", "", "PingGopher API Server URL (optional)")
	fs.Parse(args)

	targetURL := *serverURL
	if targetURL == "" {
		if creds, err := LoadCredentials(); err == nil {
			targetURL = creds.ServerURL
		} else {
			targetURL = "http://localhost:8080"
		}
	}

	env, err := DoAPIRequest("GET", targetURL, "/v1/status/public", nil, "")
	if err != nil {
		fmt.Printf("[ERROR] Failed to fetch system status: %v\n", err)
		os.Exit(1)
	}

	var statusData struct {
		SystemStatus  string `json:"system_status"`
		TotalMonitors int    `json:"total_monitors"`
		UpMonitors    int    `json:"up_monitors"`
		DownMonitors  int    `json:"down_monitors"`
		Monitors      []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"monitors"`
	}
	json.Unmarshal(env.Data, &statusData)

	fmt.Println("\n==================================================")
	fmt.Printf(" SYSTEM STATUS: %s\n", statusData.SystemStatus)
	fmt.Println("==================================================")
	fmt.Printf(" Monitored Targets : %d Total (%d UP, %d DOWN)\n\n", statusData.TotalMonitors, statusData.UpMonitors, statusData.DownMonitors)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "COMPONENT SERVICE\tSTATUS")
	fmt.Fprintln(w, "-----------------\t------")
	for _, m := range statusData.Monitors {
		fmt.Fprintf(w, "%s\t%s\n", m.Name, m.Status)
	}
	w.Flush()
	fmt.Println()
}

func handleMonitor(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pinggopher-cli monitor <list|add|delete> [options]")
		os.Exit(1)
	}

	subCmd := args[0]
	subArgs := args[1:]

	creds, err := LoadCredentials()
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}

	switch subCmd {
	case "list":
		env, err := DoAPIRequest("GET", creds.ServerURL, "/v1/monitors", nil, creds.Token)
		if err != nil {
			fmt.Printf("[ERROR] Failed to list monitors: %v\n", err)
			os.Exit(1)
		}

		var monitors []Monitor
		json.Unmarshal(env.Data, &monitors)

		if len(monitors) == 0 {
			fmt.Println("No target monitors configured yet.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTARGET URL\tSTATUS\tINTERVAL")
		fmt.Fprintln(w, "--\t----\t----------\t------\t--------")
		for _, m := range monitors {
			idDisplay := m.ID
			if len(m.ID) > 8 {
				idDisplay = m.ID[:8]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%ds\n", idDisplay, m.Name, m.URL, m.Status, m.CheckIntervalSeconds)
		}
		w.Flush()

	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		name := fs.String("name", "", "Target Monitor Name")
		url := fs.String("url", "", "Target Monitor URL")
		interval := fs.Int("interval", 60, "Check Interval in Seconds")
		fs.Parse(subArgs)

		if *name == "" || *url == "" {
			fmt.Println("Error: --name and --url flags are required.")
			fs.Usage()
			os.Exit(1)
		}

		payload := map[string]interface{}{
			"name":                   *name,
			"url":                    *url,
			"check_interval_seconds": *interval,
		}

		env, err := DoAPIRequest("POST", creds.ServerURL, "/v1/monitors", payload, creds.Token)
		if err != nil {
			fmt.Printf("[ERROR] Failed to create monitor: %v\n", err)
			os.Exit(1)
		}

		var created Monitor
		json.Unmarshal(env.Data, &created)
		fmt.Printf("[SUCCESS] Created target monitor '%s' (ID: %s)\n", created.Name, created.ID)

	case "delete":
		fs := flag.NewFlagSet("delete", flag.ExitOnError)
		id := fs.String("id", "", "Target Monitor UUID")
		fs.Parse(subArgs)

		if *id == "" {
			fmt.Println("Error: --id flag is required.")
			fs.Usage()
			os.Exit(1)
		}

		_, err := DoAPIRequest("DELETE", creds.ServerURL, "/v1/monitors/"+*id, nil, creds.Token)
		if err != nil {
			fmt.Printf("[ERROR] Failed to delete monitor: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("[SUCCESS] Deleted target monitor '%s'\n", *id)

	default:
		fmt.Printf("Unknown monitor subcommand '%s'. Use: list, add, delete\n", subCmd)
		os.Exit(1)
	}
}

func handleLogs(args []string) {
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	id := fs.String("id", "", "Target Monitor UUID")
	limit := fs.Int("limit", 20, "Number of recent logs to fetch")
	fs.Parse(args)

	if *id == "" {
		fmt.Println("Error: --id flag is required.")
		fs.Usage()
		os.Exit(1)
	}

	creds, err := LoadCredentials()
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}

	endpoint := fmt.Sprintf("/v1/monitors/%s/logs?limit=%d", *id, *limit)
	env, err := DoAPIRequest("GET", creds.ServerURL, endpoint, nil, creds.Token)
	if err != nil {
		fmt.Printf("[ERROR] Failed to fetch logs: %v\n", err)
		os.Exit(1)
	}

	var logs []PingLog
	json.Unmarshal(env.Data, &logs)

	if len(logs) == 0 {
		fmt.Printf("No probe execution logs found for monitor '%s'.\n", *id)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "STATUS CODE\tLATENCY (MS)\tTIMESTAMP")
	fmt.Fprintln(w, "-----------\t------------\t---------")
	for _, l := range logs {
		fmt.Fprintf(w, "%d\t%d ms\t%s\n", l.StatusCode, l.ResponseTimeMS, l.CreatedAt.Format(time.RFC3339))
	}
	w.Flush()
}
