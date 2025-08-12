package jobs

import (
	"log/slog"

	"github.com/hibiken/asynq"
)

// Scheduler wraps asynq.Scheduler for periodic job scheduling
type Scheduler struct {
	scheduler *asynq.Scheduler
	logger    *slog.Logger
}

// NewScheduler creates a new job scheduler
func NewScheduler(redisAddr, redisPassword string, redisDB int) *Scheduler {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		LogLevel: asynq.InfoLevel,
	})

	return &Scheduler{
		scheduler: scheduler,
		logger:    slog.Default().With("component", "job-scheduler"),
	}
}

// RegisterPeriodicTasks registers all periodic tasks
func (s *Scheduler) RegisterPeriodicTasks() error {
	s.logger.Info("Registering periodic tasks")

	// Create PR status sync job
	prStatusSyncJob, err := NewPRStatusSyncJob()
	if err != nil {
		s.logger.Error("Failed to create PR status sync job", "error", err)
		return err
	}

	// Register PR status sync to run every 30 seconds in monitoring queue
	_, err = s.scheduler.Register("@every 30s", prStatusSyncJob, asynq.Queue("monitoring"))
	if err != nil {
		s.logger.Error("Failed to register PR status sync job", "error", err)
		return err
	}

	s.logger.Info("PR status sync job registered to run every 30 seconds")
	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.logger.Info("Starting job scheduler")
	if err := s.RegisterPeriodicTasks(); err != nil {
		return err
	}
	return s.scheduler.Run()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping job scheduler")
	s.scheduler.Shutdown()
}