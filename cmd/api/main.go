package main

import (
	"log"

	"github.com/apardo/mikrom-go/config"
	"github.com/apardo/mikrom-go/internal/handlers"
	"github.com/apardo/mikrom-go/internal/middleware"
	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/apardo/mikrom-go/internal/service"
	"github.com/apardo/mikrom-go/pkg/database"
	"github.com/apardo/mikrom-go/pkg/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	// Conectar a la base de datos
	db, err := database.NewDatabase(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run auto migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.VM{},
		&models.IPPool{},
		&models.IPAllocation{},
	); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Inicializar repositorios
	userRepo := repository.NewUserRepository(db.DB)
	vmRepo := repository.NewVMRepository(db.DB)

	// Initialize worker client
	workerClient := worker.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer workerClient.Close()

	// Inicializar servicios
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	vmService := service.NewVMService(vmRepo, workerClient)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	vmHandler := handlers.NewVMHandler(vmService)

	// Configurar Gin
	router := gin.Default()

	// Rutas públicas
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Rutas protegidas (requieren autenticación)
		protected := api.Group("/auth")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/profile", authHandler.GetProfile)
		}

		// VM routes (protected)
		vms := api.Group("/vms")
		vms.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			vms.POST("", vmHandler.CreateVM)
			vms.GET("", vmHandler.ListVMs)
			vms.GET("/:vm_id", vmHandler.GetVM)
			vms.PATCH("/:vm_id", vmHandler.UpdateVM)
			vms.DELETE("/:vm_id", vmHandler.DeleteVM)
			vms.POST("/:vm_id/start", vmHandler.StartVM)
			vms.POST("/:vm_id/stop", vmHandler.StopVM)
			vms.POST("/:vm_id/restart", vmHandler.RestartVM)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Iniciar servidor
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
