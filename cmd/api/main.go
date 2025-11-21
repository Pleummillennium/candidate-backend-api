package main

import (
	"candidate-backend/internal/config"
	"candidate-backend/internal/database"
	"candidate-backend/internal/handlers"
	"candidate-backend/internal/middleware"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "candidate-backend/docs" // Import generated docs
)

// @title           Candidate Backend API
// @version         1.0
// @description     Task management system with authentication, comments, and change logs
// @description     API for managing tasks, comments, and tracking changes with user authentication

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, cfg.JWTSecret)
	taskHandler := handlers.NewTaskHandler(db.DB)
	commentHandler := handlers.NewCommentHandler(db.DB)

	// Setup router
	router := gin.Default()

	// Apply CORS middleware (must be before other routes)
	router.Use(middleware.CORSMiddleware())

	// Apply rate limiting to all routes
	router.Use(middleware.RateLimitMiddleware())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Task routes
		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/archived", taskHandler.GetArchivedTasks)
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.POST("/:id/archive", taskHandler.ArchiveTask)
			tasks.POST("/:id/unarchive", taskHandler.UnarchiveTask)
			tasks.GET("/:id/logs", taskHandler.GetTaskLogs)

			// Comment routes
			tasks.GET("/:id/comments", commentHandler.GetComments)
			tasks.POST("/:id/comments", commentHandler.CreateComment)
		}

		// Comment update/delete routes
		comments := api.Group("/comments")
		{
			comments.PUT("/:id", commentHandler.UpdateComment)
			comments.DELETE("/:id", commentHandler.DeleteComment)
		}
	}

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
