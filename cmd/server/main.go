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

	// Setup Gin router
	router := gin.Default()

	// Setup health check endpoint
	handler.SetupHealthRoutes(router)

	// Start server
	port := app.Config.Server.Port
	if port == "" {
		port = app.Config.Server.Port
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
