//go:build wireinject
// +build wireinject

package di

import (
	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/repository/postgres"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the entire application
var ProviderSet = wire.NewSet(
	config.Load,
	ProvideGormDB,
	// Repository providers
	postgres.NewProjectRepository,
	postgres.NewTaskRepository,
	// Add usecase providers here
	// Add handler providers here
)

// InitializeApp builds the entire dependency tree
func InitializeApp() (*App, error) {
	wire.Build(
		ProviderSet,
		NewApp,
	)
	return &App{}, nil
}

// App represents the initialized application with all dependencies
type App struct {
	Config      *config.Config
	GormDB      *database.GormDB
	ProjectRepo repository.ProjectRepository
	TaskRepo    repository.TaskRepository
	// Add other dependencies here as needed
}

// NewApp creates a new App instance
func NewApp(
	cfg *config.Config,
	gormDB *database.GormDB,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
) *App {
	return &App{
		Config:      cfg,
		GormDB:      gormDB,
		ProjectRepo: projectRepo,
		TaskRepo:    taskRepo,
	}
}

// ProvideGormDB provides a GORM database connection
func ProvideGormDB(cfg *config.Config) (*database.GormDB, error) {
	return database.NewGormDB(cfg)
}
