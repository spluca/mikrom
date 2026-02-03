package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear all environment variables
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	config := LoadConfig()

	// Test database defaults
	assert.Equal(t, "localhost", config.DBHost)
	assert.Equal(t, "5432", config.DBPort)
	assert.Equal(t, "postgres", config.DBUser)
	assert.Equal(t, "postgres", config.DBPassword)
	assert.Equal(t, "mikrom", config.DBName)

	// Test server defaults
	assert.Equal(t, "8080", config.ServerPort)
	assert.Equal(t, "your-secret-key-change-this", config.JWTSecret)

	// Test Redis defaults
	assert.Equal(t, "localhost:6379", config.RedisAddr)
	assert.Equal(t, "", config.RedisPassword)
	assert.Equal(t, 0, config.RedisDB)

	// Test Firecracker defaults
	assert.Equal(t, "/path/to/firecracker-deploy", config.FirecrackerDeployPath)
	assert.Equal(t, "", config.FirecrackerDefaultHost)

	// Test Worker defaults
	assert.Equal(t, 10, config.WorkerConcurrency)
}

func TestLoadConfig_CustomDatabaseValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "customuser")
	os.Setenv("DB_PASSWORD", "custompass")
	os.Setenv("DB_NAME", "customdb")

	config := LoadConfig()

	assert.Equal(t, "custom-host", config.DBHost)
	assert.Equal(t, "3306", config.DBPort)
	assert.Equal(t, "customuser", config.DBUser)
	assert.Equal(t, "custompass", config.DBPassword)
	assert.Equal(t, "customdb", config.DBName)
}

func TestLoadConfig_CustomServerValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("JWT_SECRET", "super-secret-key")

	config := LoadConfig()

	assert.Equal(t, "9090", config.ServerPort)
	assert.Equal(t, "super-secret-key", config.JWTSecret)
}

func TestLoadConfig_CustomRedisValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	os.Setenv("REDIS_ADDR", "redis.example.com:6379")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("REDIS_DB", "2")

	config := LoadConfig()

	assert.Equal(t, "redis.example.com:6379", config.RedisAddr)
	assert.Equal(t, "redispass", config.RedisPassword)
	assert.Equal(t, 2, config.RedisDB)
}

func TestLoadConfig_CustomFirecrackerValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	os.Setenv("FIRECRACKER_DEPLOY_PATH", "/custom/firecracker/path")
	os.Setenv("FIRECRACKER_DEFAULT_HOST", "fc-host.example.com")

	config := LoadConfig()

	assert.Equal(t, "/custom/firecracker/path", config.FirecrackerDeployPath)
	assert.Equal(t, "fc-host.example.com", config.FirecrackerDefaultHost)
}

func TestLoadConfig_CustomWorkerValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	os.Setenv("WORKER_CONCURRENCY", "20")

	config := LoadConfig()

	assert.Equal(t, 20, config.WorkerConcurrency)
}

func TestGetDBConnectionString(t *testing.T) {
	config := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "testdb",
	}

	connStr := config.GetDBConnectionString()

	assert.Contains(t, connStr, "host=localhost")
	assert.Contains(t, connStr, "port=5432")
	assert.Contains(t, connStr, "user=testuser")
	assert.Contains(t, connStr, "password=testpass")
	assert.Contains(t, connStr, "dbname=testdb")
	assert.Contains(t, connStr, "sslmode=disable")

	// Verify exact format
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, connStr)
}

func TestGetDBConnectionString_WithSpecialCharacters(t *testing.T) {
	config := &Config{
		DBHost:     "db.example.com",
		DBPort:     "5433",
		DBUser:     "user@domain",
		DBPassword: "p@ssw0rd!#$",
		DBName:     "my-database",
	}

	connStr := config.GetDBConnectionString()

	assert.Contains(t, connStr, "host=db.example.com")
	assert.Contains(t, connStr, "port=5433")
	assert.Contains(t, connStr, "user=user@domain")
	assert.Contains(t, connStr, "password=p@ssw0rd!#$")
	assert.Contains(t, connStr, "dbname=my-database")
}

func TestGetEnv_WithValue(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")

	assert.Equal(t, "test-value", result)
}

func TestGetEnv_WithoutValue(t *testing.T) {
	os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")

	assert.Equal(t, "default", result)
}

func TestGetEnv_WithEmptyValue(t *testing.T) {
	os.Setenv("TEST_VAR", "")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")

	// Empty string should return default
	assert.Equal(t, "default", result)
}

func TestGetEnvInt_WithValidValue(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)

	assert.Equal(t, 42, result)
}

func TestGetEnvInt_WithoutValue(t *testing.T) {
	os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)

	assert.Equal(t, 10, result)
}

func TestGetEnvInt_WithInvalidValue(t *testing.T) {
	os.Setenv("TEST_INT", "not-a-number")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)

	// Should return default when parsing fails
	assert.Equal(t, 10, result)
}

func TestGetEnvInt_WithEmptyValue(t *testing.T) {
	os.Setenv("TEST_INT", "")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)

	assert.Equal(t, 10, result)
}

func TestGetEnvInt_WithZero(t *testing.T) {
	os.Setenv("TEST_INT", "0")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)

	assert.Equal(t, 0, result)
}

func TestGetEnvInt_WithNegativeValue(t *testing.T) {
	os.Setenv("TEST_INT", "-5")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 10)

	assert.Equal(t, -5, result)
}

func TestLoadConfig_AllCustomValues(t *testing.T) {
	clearConfigEnvVars()
	defer clearConfigEnvVars()

	// Set all environment variables
	os.Setenv("DB_HOST", "prod-db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "produser")
	os.Setenv("DB_PASSWORD", "prodpass")
	os.Setenv("DB_NAME", "proddb")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("JWT_SECRET", "production-secret")
	os.Setenv("REDIS_ADDR", "redis.prod.com:6379")
	os.Setenv("REDIS_PASSWORD", "redisprodpass")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("FIRECRACKER_DEPLOY_PATH", "/prod/firecracker")
	os.Setenv("FIRECRACKER_DEFAULT_HOST", "fc.prod.com")
	os.Setenv("WORKER_CONCURRENCY", "50")

	config := LoadConfig()

	// Verify all values
	assert.Equal(t, "prod-db.example.com", config.DBHost)
	assert.Equal(t, "5433", config.DBPort)
	assert.Equal(t, "produser", config.DBUser)
	assert.Equal(t, "prodpass", config.DBPassword)
	assert.Equal(t, "proddb", config.DBName)
	assert.Equal(t, "3000", config.ServerPort)
	assert.Equal(t, "production-secret", config.JWTSecret)
	assert.Equal(t, "redis.prod.com:6379", config.RedisAddr)
	assert.Equal(t, "redisprodpass", config.RedisPassword)
	assert.Equal(t, 1, config.RedisDB)
	assert.Equal(t, "/prod/firecracker", config.FirecrackerDeployPath)
	assert.Equal(t, "fc.prod.com", config.FirecrackerDefaultHost)
	assert.Equal(t, 50, config.WorkerConcurrency)
}

// Helper function to clear all config-related environment variables
func clearConfigEnvVars() {
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"SERVER_PORT", "JWT_SECRET",
		"REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB",
		"FIRECRACKER_DEPLOY_PATH", "FIRECRACKER_DEFAULT_HOST",
		"WORKER_CONCURRENCY",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
