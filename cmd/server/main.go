package main

import (
	"log"
	"path/filepath"

	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/handler"
	"github.com/auto-devs/auto-devs/pkg/database"
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
		if err := app.DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Run database migrations
	if err := runMigrations(app); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Setup health check endpoint
	handler.SetupHealthRoutes(router, app.DB)

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

// runMigrations runs database migrations
func runMigrations(app *di.App) error {
	migrationsPath, err := filepath.Abs("./migrations")
	if err != nil {
		return err
	}
	
	return database.RunMigrations(app.DB, migrationsPath)
}