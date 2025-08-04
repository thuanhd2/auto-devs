//go:build wireinject
// +build wireinject

package di

import (
	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/repository/postgres"
	"github.com/auto-devs/auto-devs/internal/service/git"
	worktreesvc "github.com/auto-devs/auto-devs/internal/service/worktree"
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
	ProvideWorktreeRepository,
	postgres.NewAuditRepository,
	// Service providers
	ProvideGitManager,
	ProvideIntegratedWorktreeService,
	// Usecase providers
	usecase.NewNotificationUsecase,
	ProvideAuditService,
	ProvideProjectUsecase,
	ProvideWorktreeUsecase,
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
	Config              *config.Config
	GormDB              *database.GormDB
	ProjectRepo         repository.ProjectRepository
	TaskRepo            repository.TaskRepository
	WorktreeRepo        repository.WorktreeRepository
	AuditRepo           repository.AuditRepository
	AuditService        usecase.AuditService
	ProjectUsecase      usecase.ProjectUsecase
	TaskUsecase         usecase.TaskUsecase
	WorktreeUsecase     usecase.WorktreeUsecase
	NotificationUsecase usecase.NotificationUsecase
}

// NewApp creates a new App instance
func NewApp(
	cfg *config.Config,
	gormDB *database.GormDB,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	worktreeRepo repository.WorktreeRepository,
	auditRepo repository.AuditRepository,
	auditService usecase.AuditService,
	projectUsecase usecase.ProjectUsecase,
	taskUsecase usecase.TaskUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
	notificationUsecase usecase.NotificationUsecase,
) *App {
	return &App{
		Config:              cfg,
		GormDB:              gormDB,
		ProjectRepo:         projectRepo,
		TaskRepo:            taskRepo,
		WorktreeRepo:        worktreeRepo,
		AuditRepo:           auditRepo,
		AuditService:        auditService,
		ProjectUsecase:      projectUsecase,
		TaskUsecase:         taskUsecase,
		WorktreeUsecase:     worktreeUsecase,
		NotificationUsecase: notificationUsecase,
	}
}

// ProvideGormDB provides a GORM database connection
func ProvideGormDB(cfg *config.Config) (*database.GormDB, error) {
	return database.NewGormDB(cfg)
}

// ProvideWorktreeRepository provides a WorktreeRepository instance
func ProvideWorktreeRepository(gormDB *database.GormDB) repository.WorktreeRepository {
	return postgres.NewWorktreeRepository(gormDB.DB)
}

// ProvideAuditService provides an AuditService instance
func ProvideAuditService(auditRepo repository.AuditRepository) usecase.AuditService {
	return usecase.NewAuditService(auditRepo)
}

// ProvideGitManager provides a GitManager instance
func ProvideGitManager(cfg *config.Config) (*git.GitManager, error) {
	gitConfig := &git.ManagerConfig{
		DefaultTimeout: 30,
		MaxRetries:     3,
		EnableLogging:  true,
	}
	return git.NewGitManager(gitConfig)
}

// ProvideIntegratedWorktreeService provides an IntegratedWorktreeService instance
func ProvideIntegratedWorktreeService(cfg *config.Config, gitManager *git.GitManager) (*worktreesvc.IntegratedWorktreeService, error) {
	worktreeConfig := &worktreesvc.WorktreeConfig{
		BaseDirectory: "/tmp/worktrees",
	}
	integratedConfig := &worktreesvc.IntegratedConfig{
		Worktree: worktreeConfig,
		Git:      &git.ManagerConfig{},
	}
	return worktreesvc.NewIntegratedWorktreeService(integratedConfig)
}

// ProvideProjectUsecase provides a ProjectUsecase instance
func ProvideProjectUsecase(projectRepo repository.ProjectRepository, auditService usecase.AuditService) usecase.ProjectUsecase {
	return usecase.NewProjectUsecase(projectRepo, auditService)
}

// ProvideWorktreeUsecase provides a WorktreeUsecase instance
func ProvideWorktreeUsecase(
	worktreeRepo repository.WorktreeRepository,
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	integratedWorktreeSvc *worktreesvc.IntegratedWorktreeService,
	gitManager *git.GitManager,
) usecase.WorktreeUsecase {
	return usecase.NewWorktreeUsecase(worktreeRepo, taskRepo, projectRepo, integratedWorktreeSvc, gitManager)
}

// ProvideTaskUsecase provides a TaskUsecase instance
func ProvideTaskUsecase(
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	notificationUsecase usecase.NotificationUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
) usecase.TaskUsecase {
	return usecase.NewTaskUsecase(taskRepo, projectRepo, notificationUsecase, worktreeUsecase)
}
