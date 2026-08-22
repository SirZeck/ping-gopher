package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pinggopher/ping-gopher/internal/config"
	"github.com/pinggopher/ping-gopher/internal/db"
)

const banner = `
=====================================================
  PingGopher — Uptime & Synthetic Monitoring SaaS
  Powered by gopher-queue | Role: %-10s
=====================================================
`

func main() {
	cfg := config.LoadConfig()

	fmt.Printf(banner, cfg.Role)
	fmt.Printf("[INFO] Initializing PingGopher (Role: %s, DB: %s)\n", cfg.Role, cfg.DatabasePath)

	// Initialize Database Connection & Migrations
	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		fmt.Printf("[FATAL] Database initialization failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[SUCCESS] Database initialized and GORM models auto-migrated successfully.")

	// Boot active role modules
	switch cfg.Role {
	case "all":
		startAPIRole(cfg, database)
		startWorkerRole(cfg, database)
		startSchedulerRole(cfg, database)
	case "api":
		startAPIRole(cfg, database)
	case "worker":
		startWorkerRole(cfg, database)
	case "scheduler":
		startSchedulerRole(cfg, database)
	default:
		fmt.Printf("[FATAL] Unknown role: '%s'. Valid roles are: all, api, worker, scheduler\n", cfg.Role)
		os.Exit(1)
	}

	// Wait for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	fmt.Printf("\n[INFO] Received signal '%v'. Gracefully shutting down PingGopher...\n", sig)
	fmt.Println("[INFO] PingGopher stopped cleanly.")
}

func startAPIRole(cfg *config.Config, database interface{}) {
	fmt.Printf("[ROLE: API] Starting REST API Server on port %s...\n", cfg.Port)
}

func startWorkerRole(cfg *config.Config, database interface{}) {
	fmt.Printf("[ROLE: WORKER] Starting gopher-queue task worker pool connected to Redis at %s...\n", cfg.RedisAddr)
}

func startSchedulerRole(cfg *config.Config, database interface{}) {
	fmt.Println("[ROLE: SCHEDULER] Starting monitor check scheduler loop...")
}
