package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SirZeck/ping-gopher/internal/api"
	"github.com/SirZeck/ping-gopher/internal/config"
	"github.com/SirZeck/ping-gopher/internal/db"
	"github.com/SirZeck/ping-gopher/internal/scheduler"
	"github.com/SirZeck/ping-gopher/internal/worker"
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

	workerEngine := worker.NewWorkerEngine(database)
	sched := scheduler.NewScheduler(database, workerEngine)
	apiHandler := api.NewAPIHandler(database, cfg)

	var server *http.Server

	// Boot active role modules
	switch cfg.Role {
	case "all":
		server = startAPIRole(cfg, apiHandler)
		startWorkerRole(cfg, workerEngine)
		startSchedulerRole(cfg, sched)
	case "api":
		server = startAPIRole(cfg, apiHandler)
	case "worker":
		startWorkerRole(cfg, workerEngine)
	case "scheduler":
		startSchedulerRole(cfg, sched)
	default:
		fmt.Printf("[FATAL] Unknown role: '%s'. Valid roles are: all, api, worker, scheduler\n", cfg.Role)
		os.Exit(1)
	}

	// Wait for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	fmt.Printf("\n[INFO] Received signal '%v'. Gracefully shutting down PingGopher...\n", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if server != nil {
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("[WARN] HTTP Server graceful shutdown error: %v\n", err)
		} else {
			fmt.Println("[INFO] HTTP Server stopped gracefully.")
		}
	}

	if cfg.Role == "all" || cfg.Role == "scheduler" {
		sched.Stop()
	}
	fmt.Println("[INFO] PingGopher stopped cleanly.")
}

func startAPIRole(cfg *config.Config, handler *api.APIHandler) *http.Server {
	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("[ROLE: API] Starting REST API Server on %s...\n", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      handler.SetupRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[FATAL] API Server failed: %v\n", err)
		}
	}()

	return server
}

func startWorkerRole(cfg *config.Config, engine *worker.WorkerEngine) {
	fmt.Printf("[ROLE: WORKER] Probe execution worker engine ready (Redis target: %s)\n", cfg.RedisAddr)
}

func startSchedulerRole(cfg *config.Config, sched *scheduler.Scheduler) {
	fmt.Println("[ROLE: SCHEDULER] Booting monitor check promoter loop...")
	sched.Start(10 * time.Second)
}
