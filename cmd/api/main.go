package main

import (
	"log"

	"github.com/apardo/mikrom-go/config"
	"github.com/apardo/mikrom-go/internal/handlers"
	"github.com/apardo/mikrom-go/internal/middleware"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/apardo/mikrom-go/internal/service"
	"github.com/apardo/mikrom-go/pkg/database"
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

	// Crear tablas si no existen
	if err := db.CreateTables(); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Inicializar repositorios
	userRepo := repository.NewUserRepository(db.DB)

	// Inicializar servicios
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)

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
