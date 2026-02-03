package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	JWTSecret  string

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Firecracker configuration
	FirecrackerDeployPath  string
	FirecrackerDefaultHost string

	// Worker configuration
	WorkerConcurrency int
}

func LoadConfig() *Config {
	// Cargar archivo .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "mikrom"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key-change-this"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// Firecracker
		FirecrackerDeployPath:  getEnv("FIRECRACKER_DEPLOY_PATH", "/path/to/firecracker-deploy"),
		FirecrackerDefaultHost: getEnv("FIRECRACKER_DEFAULT_HOST", ""),

		// Worker
		WorkerConcurrency: getEnvInt("WORKER_CONCURRENCY", 10),
	}

	return config
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := fmt.Sscanf(value, "%d", new(int)); err == nil && intVal == 1 {
			var result int
			fmt.Sscanf(value, "%d", &result)
			return result
		}
	}
	return defaultValue
}
