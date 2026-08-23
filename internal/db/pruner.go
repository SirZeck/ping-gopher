package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PruneOldLogs purges PingLog entries older than retentionDays (default 30 days) to prevent database bloat.
func PruneOldLogs(database *gorm.DB, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	result := database.Where("created_at < ?", cutoffTime).Delete(&PingLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to prune old telemetry logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}
