package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/handler"
	"github.com/gin-gonic/gin"
)

// isAPIRoute checks if the given path is an API route
func isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/swagger/") ||
		strings.HasPrefix(path, "/health") ||
		strings.HasPrefix(path, "/ws")
}

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

	// Initialize WebSocket service
	log.Printf("WebSocket service initialized")

	// Start WebSocket service
	if err := app.WebSocketService.Start(); err != nil {
		log.Printf("Warning: Failed to start WebSocket service: %v", err)
	} else {
		log.Printf("WebSocket service started successfully")
	}

	// Setup Gin router
	router := gin.Default()

	// Setup all routes with middleware
	handler.SetupRoutes(router, app.ProjectUsecase, app.TaskUsecase, app.ExecutionUsecase, app.GormDB, app.WebSocketService)

	runMode := app.Config.Server.RunMode

	if runMode == "production" {
		frontendPath := "./public"
		// Check if frontend is built
		if _, err := os.Stat(frontendPath + "/index.html"); os.IsNotExist(err) {
			log.Printf("Warning: Frontend not built. Please run 'make build-frontend' or 'make build-full' first")
			log.Printf("Serving API only. Frontend will not be available.")
		} else {
			// Serve static files from frontend build output
			router.Static("/assets", frontendPath+"/assets")
			router.Static("/images", frontendPath+"/images")
			router.Static("/sounds", frontendPath+"/sounds")
			router.GET("/", func(c *gin.Context) {
				c.File(frontendPath + "/index.html")
			})

			// Handle SPA routing - serve index.html for all non-API routes
			router.NoRoute(func(c *gin.Context) {
				// Check if the request is for an API route
				if c.Request.URL.Path == "/" || !isAPIRoute(c.Request.URL.Path) {
					c.File(frontendPath + "/index.html")
				} else {
					c.JSON(404, gin.H{"error": "Not found"})
				}
			})
		}
	}

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
	if wsHandler := app.WebSocketService.GetHandler(); wsHandler != nil {
		wsHandler.Shutdown()
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
