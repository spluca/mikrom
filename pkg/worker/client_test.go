package worker

import (
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

// Note: These tests verify the client logic without requiring a real Redis instance.
// Full integration tests with actual Redis would require a running Redis instance
// and are better suited for integration test suites.

func TestNewClient_Creation(t *testing.T) {
	// Test that NewClient creates a client without panicking
	client := NewClient("localhost:6379", "", 0)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)

	// Close should not error even without actual connection
	err := client.Close()
	assert.NoError(t, err)
}

func TestNewClient_WithPassword(t *testing.T) {
	// Test that NewClient accepts password and DB parameters
	client := NewClient("localhost:6379", "mypassword", 2)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)

	err := client.Close()
	assert.NoError(t, err)
}

func TestNewClient_DifferentAddress(t *testing.T) {
	// Test with different Redis addresses
	testCases := []struct {
		name     string
		addr     string
		password string
		db       int
	}{
		{"localhost default", "localhost:6379", "", 0},
		{"custom port", "localhost:6380", "", 0},
		{"with password", "localhost:6379", "secret", 0},
		{"different DB", "localhost:6379", "", 5},
		{"remote host", "redis.example.com:6379", "password", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(tc.addr, tc.password, tc.db)
			assert.NotNil(t, client)
			assert.NotNil(t, client.client)
			err := client.Close()
			assert.NoError(t, err)
		})
	}
}

func TestClient_Close(t *testing.T) {
	// Test that Close works without panic
	client := NewClient("localhost:6379", "", 0)

	err := client.Close()
	assert.NoError(t, err)

	// Note: Closing again will return an error "redis: client is closed"
	// which is expected behavior from the underlying Redis client.
	// We don't test double-close as it's not a supported operation.
}

func TestClient_StructureValidation(t *testing.T) {
	// Verify the Client struct has the expected fields
	client := NewClient("localhost:6379", "", 0)
	defer client.Close()

	assert.NotNil(t, client.client)
	// Verify it's an asynq.Client
	assert.IsType(t, &asynq.Client{}, client.client)
}

// Note about enqueue tests:
// Testing EnqueueCreateVM, EnqueueDeleteVM, etc. requires either:
// 1. A real Redis instance (integration test)
// 2. Mocking the asynq.Client.Enqueue method (complex, not worth it for unit tests)
// 3. Using asynqtest package (if available)
//
// These tests focus on what can be tested without Redis:
// - Client creation
// - Parameter validation
// - Close behavior
//
// The actual enqueue logic is tested indirectly through:
// - Service layer tests (using mock worker)
// - Handler tests (using mock worker)
// - Integration tests (with real Redis - not in this file)
