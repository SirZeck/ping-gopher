package db

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the GORM database connection using CGO-free pure Go SQLite driver and runs automatic migrations.
func InitDB(dbPath string) (*gorm.DB, error) {
	// Enable SQLite foreign keys, Write-Ahead Logging (WAL), and 5s busy timeout using DSN params
	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	database, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Limit max open connections to 1 for SQLite to prevent SQLITE_BUSY locking
	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Auto-migrate domain schemas
	err = database.AutoMigrate(
		&User{},
		&Monitor{},
		&PingLog{},
		&Incident{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schemas: %w", err)
	}

	return database, nil
}
