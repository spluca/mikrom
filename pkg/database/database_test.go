package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabase_InvalidConnectionString(t *testing.T) {
	// Test with empty connection string
	db, err := NewDatabase("")

	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestNewDatabase_MalformedConnectionString(t *testing.T) {
	// Test with malformed connection string
	db, err := NewDatabase("invalid connection string format")

	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestNewDatabase_InvalidHost(t *testing.T) {
	// Test with invalid host (non-existent)
	connStr := "host=nonexistent-host-123456.invalid port=5432 user=test password=test dbname=test sslmode=disable"
	db, err := NewDatabase(connStr)

	assert.Error(t, err)
	assert.Nil(t, db)
	// Error should be about connection failure
	assert.Contains(t, err.Error(), "failed to")
}

func TestNewDatabase_InvalidPort(t *testing.T) {
	// Test with invalid port
	connStr := "host=localhost port=99999 user=test password=test dbname=test sslmode=disable"
	db, err := NewDatabase(connStr)

	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewDatabase_MissingRequiredFields(t *testing.T) {
	testCases := []struct {
		name    string
		connStr string
	}{
		{
			name:    "missing host",
			connStr: "port=5432 user=test password=test dbname=test sslmode=disable",
		},
		{
			name:    "missing user",
			connStr: "host=localhost port=5432 password=test dbname=test sslmode=disable",
		},
		{
			name:    "missing dbname",
			connStr: "host=localhost port=5432 user=test password=test sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewDatabase(tc.connStr)

			// Should fail to connect due to missing fields
			assert.Error(t, err)
			assert.Nil(t, db)
		})
	}
}

// Note: Testing AutoMigrate, successful NewDatabase, and Close with valid connection
// requires an actual database connection (PostgreSQL or test database).
// These are better suited for integration tests rather than unit tests.
//
// The tests above focus on error handling and validation that can be tested
// without a real database connection.
