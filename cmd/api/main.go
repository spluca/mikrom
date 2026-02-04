package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spluca/mikrom/config"
	"github.com/spluca/mikrom/internal/handlers"
	"github.com/spluca/mikrom/internal/middleware"
	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/internal/repository"
	"github.com/spluca/mikrom/internal/service"
	"github.com/spluca/mikrom/pkg/database"
	"github.com/spluca/mikrom/pkg/worker"
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
	ipPoolRepo := repository.NewIPPoolRepository(db.DB)

	// Initialize worker client
	workerClient := worker.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer workerClient.Close()

	// Inicializar servicios
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	vmService := service.NewVMService(vmRepo, workerClient)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	vmHandler := handlers.NewVMHandler(vmService)
	ipPoolHandler := handlers.NewIPPoolHandler(ipPoolService)

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

		// IP Pool routes (protected - admin only)
		ippools := api.Group("/ippools")
		ippools.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			ippools.POST("", ipPoolHandler.CreateIPPool)
			ippools.GET("", ipPoolHandler.ListIPPools)
			ippools.GET("/stats", ipPoolHandler.GetAllPoolStats)
			ippools.POST("/suggest-range", ipPoolHandler.SuggestIPRange)
			ippools.GET("/:id", ipPoolHandler.GetIPPool)
			ippools.PATCH("/:id", ipPoolHandler.UpdateIPPool)
			ippools.DELETE("/:id", ipPoolHandler.DeleteIPPool)
			ippools.GET("/:id/stats", ipPoolHandler.GetPoolStats)
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
