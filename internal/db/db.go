package db

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the GORM database connection using CGO-free pure Go SQLite driver and runs automatic migrations.
func InitDB(dbPath string) (*gorm.DB, error) {
	// Enable SQLite foreign keys using DSN param
	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)", dbPath)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	database, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
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
