package repository

import (
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Suppress logs during tests
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run auto-migrations for all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.VM{},
		&models.IPPool{},
		&models.IPAllocation{},
	); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Logf("Failed to get underlying DB: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		t.Logf("Failed to close test database: %v", err)
	}
}

// TruncateTable truncates a table in the test database
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()

	if err := db.Exec("DELETE FROM " + tableName).Error; err != nil {
		t.Fatalf("Failed to truncate table %s: %v", tableName, err)
	}
}
