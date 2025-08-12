package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/jobs"
)

func main() {
	savePidToFile()
	defer removePidFromFile()

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

	// Use job processor from DI container
	processor := app.JobProcessor

	// Create job server
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	server := jobs.NewServer(redisAddr, cfg.Redis.Password, cfg.Redis.DB, processor)

	// Create scheduler for periodic tasks
	scheduler := jobs.NewScheduler(redisAddr, cfg.Redis.Password, cfg.Redis.DB)

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

	// Start the scheduler
	logger.Info("Starting job scheduler",
		"redis_addr", redisAddr,
		"worker_name", *workerName)

	go func() {
		if err := scheduler.Start(); err != nil {
			logger.Error("Job scheduler failed", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Shutting down job worker...")
	server.Stop()
	scheduler.Stop()
	logger.Info("Job worker stopped")
}

var pidsFolder = "/private/var/folders/tv/531lt6yx3ss28h1b7bcpb1900000gn/T/autodevs"

func savePidToFile() {
	pid := os.Getpid()
	pidFile := fmt.Sprintf("%s/worker_%d.pid", pidsFolder, pid)
	// create the file
	os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0o644)
}

func removePidFromFile() {
	pid := os.Getpid()
	pidFile := fmt.Sprintf("%s/worker_%d.pid", pidsFolder, pid)
	os.Remove(pidFile)
}
