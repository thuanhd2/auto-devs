//go:build wireinject
// +build wireinject

package di

import (
	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/repository/postgres"
	"github.com/auto-devs/auto-devs/internal/usecase"
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
	// Usecase providers
	ProvideProjectUsecase,
	ProvideTaskUsecase,
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
	Config         *config.Config
	GormDB         *database.GormDB
	ProjectRepo    repository.ProjectRepository
	TaskRepo       repository.TaskRepository
	ProjectUsecase usecase.ProjectUsecase
	TaskUsecase    usecase.TaskUsecase
}

// NewApp creates a new App instance
func NewApp(
	cfg *config.Config,
	gormDB *database.GormDB,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	projectUsecase usecase.ProjectUsecase,
	taskUsecase usecase.TaskUsecase,
) *App {
	return &App{
		Config:         cfg,
		GormDB:         gormDB,
		ProjectRepo:    projectRepo,
		TaskRepo:       taskRepo,
		ProjectUsecase: projectUsecase,
		TaskUsecase:    taskUsecase,
	}
}

// ProvideGormDB provides a GORM database connection
func ProvideGormDB(cfg *config.Config) (*database.GormDB, error) {
	return database.NewGormDB(cfg)
}

// ProvideProjectUsecase provides a ProjectUsecase instance
func ProvideProjectUsecase(projectRepo repository.ProjectRepository) usecase.ProjectUsecase {
	return usecase.NewProjectUsecase(projectRepo)
}

// ProvideTaskUsecase provides a TaskUsecase instance
func ProvideTaskUsecase(taskRepo repository.TaskRepository) usecase.TaskUsecase {
	return usecase.NewTaskUsecase(taskRepo)
}
