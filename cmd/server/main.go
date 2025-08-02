package main

import (
	"log"

	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize application with Wire dependency injection
	app, err := di.InitializeApp()
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Ensure database connection is closed on exit
	defer func() {
		if err := app.GormDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// TODO: think about auto migration later!
	// // Run database migrations using GORM AutoMigrate
	// if err := database.RunMigrations(app.GormDB); err != nil {
	// 	log.Printf("Warning: Failed to run migrations: %v", err)
	// }

	// Setup Gin router
	router := gin.Default()

	// Setup all routes with middleware
	handler.SetupRoutes(router, app.ProjectUsecase, app.TaskUsecase, app.GormDB)

	// Start server
	port := app.Config.Server.Port
	if port == "" {
		port = "8098"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
