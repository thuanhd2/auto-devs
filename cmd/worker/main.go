package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/jobs"
)

func main() {
	// Parse command line flags
	var (
		workerName = flag.String("worker", "default", "Worker name for identification")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting job worker", "worker_name", *workerName)

	// Load configuration
	cfg := config.Load()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	// Initialize application dependencies
	app, err := di.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Create job processor with dependencies
	processor := jobs.NewProcessor(
		app.TaskUsecase,
		app.ProjectUsecase,
		app.WorktreeUsecase,
	)

	// Create job server
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	server := jobs.NewServer(redisAddr, cfg.Redis.Password, cfg.Redis.DB, processor)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// Start the job server
	logger.Info("Starting job server",
		"redis_addr", redisAddr,
		"worker_name", *workerName)

	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Job server failed", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Shutting down job worker...")
	server.Stop()
	logger.Info("Job worker stopped")
}
