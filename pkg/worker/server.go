package worker

import (
	"log"

	"github.com/hibiken/asynq"
)

// Server wraps asynq.Server for processing tasks
type Server struct {
	server  *asynq.Server
	mux     *asynq.ServeMux
	handler *TaskHandler
}

// ServerConfig holds configuration for the worker server
type ServerConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
}

// NewServer creates a new worker server
func NewServer(cfg ServerConfig, handler *TaskHandler) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			// Concurrency is the maximum number of concurrent processing of tasks
			Concurrency: cfg.Concurrency,

			// Queues is a map of queue name to priority level
			// Tasks in "critical" queue will be processed first
			Queues: map[string]int{
				"critical": 6, // Highest priority for VM creation/deletion
				"default":  3, // Normal priority for start/stop/restart
				"low":      1, // Lowest priority for background tasks
			},
		},
	)

	mux := asynq.NewServeMux()

	return &Server{
		server:  srv,
		mux:     mux,
		handler: handler,
	}
}

// RegisterHandlers registers all task handlers
func (s *Server) RegisterHandlers() {
	// Register VM operation handlers
	s.mux.HandleFunc(TypeCreateVM, s.handler.HandleCreateVM)
	s.mux.HandleFunc(TypeDeleteVM, s.handler.HandleDeleteVM)
	s.mux.HandleFunc(TypeStartVM, s.handler.HandleStartVM)
	s.mux.HandleFunc(TypeStopVM, s.handler.HandleStopVM)
	s.mux.HandleFunc(TypeRestartVM, s.handler.HandleRestartVM)

	log.Println("Registered task handlers:")
	log.Printf("  - %s", TypeCreateVM)
	log.Printf("  - %s", TypeDeleteVM)
	log.Printf("  - %s", TypeStartVM)
	log.Printf("  - %s", TypeStopVM)
	log.Printf("  - %s", TypeRestartVM)
}

// Start starts the worker server
func (s *Server) Start() error {
	log.Println("Starting worker server...")
	return s.server.Run(s.mux)
}

// Stop gracefully stops the worker server
func (s *Server) Stop() {
	log.Println("Stopping worker server...")
	s.server.Shutdown()
}
