package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/SirZeck/ping-gopher/internal/db"
	"github.com/SirZeck/ping-gopher/internal/worker"
	"gorm.io/gorm"
)

// Scheduler manages periodic monitor polling and probe task dispatching.
type Scheduler struct {
	DB           *gorm.DB
	WorkerEngine *worker.WorkerEngine
	Ticker       *time.Ticker
	StopChan     chan struct{}
	workerPool   chan struct{}
	wg           sync.WaitGroup
}

// NewScheduler initializes a new monitor check scheduler.
func NewScheduler(database *gorm.DB, engine *worker.WorkerEngine) *Scheduler {
	return &Scheduler{
		DB:           database,
		WorkerEngine: engine,
		StopChan:     make(chan struct{}),
		workerPool:   make(chan struct{}, 50), // Max 50 concurrent probes
	}
}

// Start boots the monitor promoter loop ticking every interval (default 10s).
func (s *Scheduler) Start(pollInterval time.Duration) {
	if pollInterval <= 0 {
		pollInterval = 10 * time.Second
	}

	s.Ticker = time.NewTicker(pollInterval)
	pruneTicker := time.NewTicker(24 * time.Hour)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		defer pruneTicker.Stop()
		fmt.Printf("[SCHEDULER] Monitor scheduler promoter loop started (Interval: %v)\n", pollInterval)

		// Execute an initial check run on boot
		s.RunCheckCycle()

		for {
			select {
			case <-s.Ticker.C:
				s.RunCheckCycle()
			case <-pruneTicker.C:
				if deleted, err := db.PruneOldLogs(s.DB, 30); err == nil && deleted > 0 {
					fmt.Printf("[SCHEDULER] Pruned %d old telemetry logs older than 30 days.\n", deleted)
				}
			case <-s.StopChan:
				fmt.Println("[SCHEDULER] Stopping monitor scheduler promoter loop...")
				return
			}
		}
	}()
}

// Stop halts the scheduler promoter loop gracefully.
func (s *Scheduler) Stop() {
	if s.Ticker != nil {
		s.Ticker.Stop()
	}
	close(s.StopChan)
	s.wg.Wait()
	fmt.Println("[SCHEDULER] Monitor scheduler stopped cleanly.")
}

// RunCheckCycle queries active monitors from DB and dispatches probe tasks.
func (s *Scheduler) RunCheckCycle() {
	var monitors []db.Monitor
	// Fetch all active monitors (status != PAUSED)
	err := s.DB.Where("status != ?", db.StatusPaused).Find(&monitors).Error
	if err != nil {
		fmt.Printf("[SCHEDULER ERROR] Failed to query active monitors: %v\n", err)
		return
	}

	now := time.Now()
	for _, m := range monitors {
		// Respect monitor's configured check_interval_seconds unless it's a new monitor never checked
		isNewMonitor := m.UpdatedAt.IsZero() || m.UpdatedAt.Equal(m.CreatedAt)
		if !isNewMonitor && now.Sub(m.UpdatedAt) < time.Duration(m.CheckIntervalSeconds)*time.Second {
			continue
		}

		payload := worker.CheckPayload{
			MonitorID: m.ID.String(),
			TargetURL: m.URL,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("[SCHEDULER ERROR] Failed to marshal payload for monitor %s: %v\n", m.ID, err)
			continue
		}

		// Track probe execution goroutine in WaitGroup and workerPool semaphore for concurrency control
		s.wg.Add(1)
		s.workerPool <- struct{}{}
		go func(p []byte, monitorID string) {
			defer s.wg.Done()
			defer func() { <-s.workerPool }()
			if err := s.WorkerEngine.ProcessHTTPCheck(p); err != nil {
				fmt.Printf("[SCHEDULER ERROR] Probe execution failed for monitor %s: %v\n", monitorID, err)
			}
		}(payloadBytes, m.ID.String())
	}
}

// RunOnce executes a single check cycle synchronously (useful for tests).
func (s *Scheduler) RunOnce(ctx context.Context) error {
	s.RunCheckCycle()
	return nil
}
