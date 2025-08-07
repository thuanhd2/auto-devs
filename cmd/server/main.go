package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/handler"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)
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

	// Initialize Enhanced WebSocket service
	wsService, err := websocket.NewEnhancedService(&app.Config.Centrifuge)
	if err != nil {
		log.Fatal("Failed to initialize WebSocket service:", err)
	}
	log.Printf("Enhanced WebSocket service initialized")

	// Setup Gin router
	router := gin.Default()

	// Setup all routes with middleware
	handler.SetupRoutes(router, app.ProjectUsecase, app.TaskUsecase, app.GormDB, wsService)

	// Start server
	port := app.Config.Server.Port
	if port == "" {
		port = "8098"
	}

	// Create server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown WebSocket connections gracefully
	if wsHandler := wsService.GetHandler(); wsHandler != nil {
		wsHandler.Shutdown()
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
