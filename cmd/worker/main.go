package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/apardo/mikrom-go/config"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/apardo/mikrom-go/pkg/database"
	"github.com/apardo/mikrom-go/pkg/firecracker"
	"github.com/apardo/mikrom-go/pkg/worker"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	log.Println("Connecting to database...")
	db, err := database.NewDatabase(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	vmRepo := repository.NewVMRepository(db.DB)
	ipPoolRepo := repository.NewIPPoolRepository(db.DB)

	// Initialize Firecracker client
	log.Printf("Initializing Firecracker client (deploy_path=%s, host=%s)...",
		cfg.FirecrackerDeployPath, cfg.FirecrackerDefaultHost)
	fcClient := firecracker.NewClient(cfg.FirecrackerDeployPath, cfg.FirecrackerDefaultHost)

	// Check Firecracker/Ansible health
	ctx := context.Background()
	if err := fcClient.CheckHealth(ctx); err != nil {
		log.Printf("WARNING: Firecracker/Ansible health check failed: %v", err)
		log.Println("Worker will start but VM operations may fail")
	} else {
		log.Println("Firecracker/Ansible health check passed")
	}

	// Initialize task handler
	taskHandler := worker.NewTaskHandler(
		db.DB,
		vmRepo,
		ipPoolRepo,
		fcClient,
		cfg.FirecrackerDeployPath,
	)

	// Initialize worker server
	workerCfg := worker.ServerConfig{
		RedisAddr:     cfg.RedisAddr,
		RedisPassword: cfg.RedisPassword,
		RedisDB:       cfg.RedisDB,
		Concurrency:   cfg.WorkerConcurrency,
	}

	log.Printf("Initializing worker server (redis=%s, concurrency=%d)...",
		cfg.RedisAddr, cfg.WorkerConcurrency)
	workerServer := worker.NewServer(workerCfg, taskHandler)

	// Register all task handlers
	workerServer.RegisterHandlers()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start worker in a goroutine
	go func() {
		log.Println("Worker server starting...")
		if err := workerServer.Start(); err != nil {
			log.Fatalf("Worker server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)
	log.Println("Shutting down worker server...")

	// Gracefully stop the worker
	workerServer.Stop()

	log.Println("Worker server stopped")
}
