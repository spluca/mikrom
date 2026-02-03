package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: These tests verify the server structure and configuration
// without requiring a real Redis instance.

func TestNewServer_Creation(t *testing.T) {
	cfg := ServerConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		Concurrency:   10,
	}

	// Create a mock handler (nil is ok for structure testing)
	handler := &TaskHandler{}

	server := NewServer(cfg, handler)

	assert.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.mux)
	assert.NotNil(t, server.handler)
	assert.Equal(t, handler, server.handler)
}

func TestNewServer_WithPassword(t *testing.T) {
	cfg := ServerConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "secret",
		RedisDB:       1,
		Concurrency:   20,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	assert.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.mux)
	assert.NotNil(t, server.handler)
}

func TestNewServer_DifferentConcurrency(t *testing.T) {
	testCases := []struct {
		name        string
		concurrency int
	}{
		{"low concurrency", 1},
		{"default concurrency", 10},
		{"high concurrency", 50},
		{"very high concurrency", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ServerConfig{
				RedisAddr:     "localhost:6379",
				RedisPassword: "",
				RedisDB:       0,
				Concurrency:   tc.concurrency,
			}

			handler := &TaskHandler{}
			server := NewServer(cfg, handler)

			assert.NotNil(t, server)
			assert.NotNil(t, server.server)
			assert.NotNil(t, server.mux)
		})
	}
}

func TestNewServer_DifferentRedisDB(t *testing.T) {
	testCases := []int{0, 1, 2, 5, 15}

	for _, db := range testCases {
		t.Run("DB_"+string(rune(db+'0')), func(t *testing.T) {
			cfg := ServerConfig{
				RedisAddr:     "localhost:6379",
				RedisPassword: "",
				RedisDB:       db,
				Concurrency:   10,
			}

			handler := &TaskHandler{}
			server := NewServer(cfg, handler)

			assert.NotNil(t, server)
		})
	}
}

func TestServerConfig_DefaultValues(t *testing.T) {
	// Test with minimal config
	cfg := ServerConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 10,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	assert.NotNil(t, server)
}

func TestServerConfig_AllFields(t *testing.T) {
	// Test with all fields populated
	cfg := ServerConfig{
		RedisAddr:     "redis.example.com:6379",
		RedisPassword: "super-secret",
		RedisDB:       3,
		Concurrency:   25,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	assert.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.mux)
	assert.NotNil(t, server.handler)
}

func TestServer_RegisterHandlers(t *testing.T) {
	cfg := ServerConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 10,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	// RegisterHandlers should not panic
	assert.NotPanics(t, func() {
		server.RegisterHandlers()
	})
}

func TestServer_Stop(t *testing.T) {
	cfg := ServerConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 10,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	// Stop should not panic even if server hasn't started
	assert.NotPanics(t, func() {
		server.Stop()
	})
}

func TestServer_MultipleStops(t *testing.T) {
	cfg := ServerConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 10,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	// Multiple stops should not panic
	assert.NotPanics(t, func() {
		server.Stop()
		server.Stop()
		server.Stop()
	})
}

func TestServer_StructureValidation(t *testing.T) {
	cfg := ServerConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 10,
	}

	handler := &TaskHandler{}
	server := NewServer(cfg, handler)

	// Verify the server has all expected fields
	assert.NotNil(t, server.server, "server.server should not be nil")
	assert.NotNil(t, server.mux, "server.mux should not be nil")
	assert.NotNil(t, server.handler, "server.handler should not be nil")
	assert.Equal(t, handler, server.handler, "handler should match")
}

// Note about Start() tests:
// Testing Start() requires either:
// 1. A real Redis instance (integration test)
// 2. Complex mocking of asynq internals
// 3. Using asynqtest package (if available)
//
// Start() blocks until the server is shut down, making it difficult
// to test in a unit test without goroutines and proper cleanup.
// This is better suited for integration tests.
//
// The tests above verify:
// - Server creation with various configs
// - RegisterHandlers behavior
// - Stop behavior
// - Structure validation
