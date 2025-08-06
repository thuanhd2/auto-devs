package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// Server wraps asynq.Server for job processing
type Server struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	processor *Processor
	logger    *slog.Logger
}

// NewServer creates a new job server
func NewServer(redisAddr, redisPassword string, redisDB int, processor *Processor) *Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				"critical":       6, // High priority queue
				"planning":       4, // Planning jobs queue
				"implementation": 4, // Implementing jobs queue
				"default":        1, // Default queue
			},
			// Concurrency settings
			Concurrency: 4,
			// Retry settings
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				// Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (max)
				delay := time.Duration(1<<uint(n)) * time.Second
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
				return delay
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				// Safely get task ID from ResultWriter
				taskID := "unknown"
				if resultWriter := task.ResultWriter(); resultWriter != nil {
					taskID = resultWriter.TaskID()
				}

				slog.Error("Task processing failed",
					"task_type", task.Type(),
					"task_id", taskID,
					"error", err,
				)
			}),
		},
	)

	mux := asynq.NewServeMux()

	return &Server{
		server:    server,
		mux:       mux,
		processor: processor,
		logger:    slog.Default().With("component", "job-server"),
	}
}

// RegisterHandlers registers job handlers
func (s *Server) RegisterHandlers() {
	s.mux.HandleFunc(TypeTaskPlanning, s.processor.ProcessTaskPlanning)
	s.mux.HandleFunc(TypeTaskImplementation, s.processor.ProcessTaskImplementation)
}

// Start starts the job server
func (s *Server) Start() error {
	s.RegisterHandlers()
	s.logger.Info("Starting job server")
	return s.server.Run(s.mux)
}

// Stop gracefully stops the job server
func (s *Server) Stop() {
	s.logger.Info("Stopping job server")
	s.server.Stop()
	s.server.Shutdown()
}
