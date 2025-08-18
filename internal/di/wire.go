//go:build wireinject
// +build wireinject

package di

import (
	"time"

	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/jobs"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/repository/postgres"
	"github.com/auto-devs/auto-devs/internal/service/ai"
	"github.com/auto-devs/auto-devs/internal/service/git"
	"github.com/auto-devs/auto-devs/internal/service/github"
	worktreesvc "github.com/auto-devs/auto-devs/internal/service/worktree"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/auto-devs/auto-devs/internal/websocket"
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
	postgres.NewPlanRepository,
	ProvideWorktreeRepository,
	postgres.NewAuditRepository,
	postgres.NewExecutionRepository,
	postgres.NewExecutionLogRepository,
	postgres.NewPullRequestRepository,
	// Service providers
	ProvideGitManager,
	ProvideProjectGitService,
	ProvideGitHubService,
	ProvidePRCreator,
	ProvideIntegratedWorktreeService,
	ProvideWorktreeManager,
	// WebSocket service provider
	ProvideWebSocketService,
	// AI Service providers
	ProvideCLIManager,
	ProvideProcessManager,
	ProvideExecutionService,
	ProvidePlanningService,
	// Job providers
	ProvideJobClient,
	ProvideJobClientAdapter,
	ProvideJobProcessor,
	// Usecase providers
	usecase.NewNotificationUsecase,
	ProvideAuditUsecase,
	ProvideProjectUsecase,
	ProvideWorktreeUsecase,
	ProvideTaskUsecase,
	ProvideExecutionUsecase,
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
	PlanRepo            repository.PlanRepository
	WorktreeRepo        repository.WorktreeRepository
	AuditRepo           repository.AuditRepository
	ExecutionRepo       repository.ExecutionRepository
	ExecutionLogRepo    repository.ExecutionLogRepository
	PullRequestRepo     repository.PullRequestRepository
	AuditUsecase        usecase.AuditUsecase
	ProjectUsecase      usecase.ProjectUsecase
	TaskUsecase         usecase.TaskUsecase
	WorktreeUsecase     usecase.WorktreeUsecase
	NotificationUsecase usecase.NotificationUsecase
	ExecutionUsecase    usecase.ExecutionUsecase
	// WebSocket Service
	WebSocketService *websocket.Service
	// AI Services
	CLIManager       *ai.CLIManager
	ProcessManager   *ai.ProcessManager
	ExecutionService *ai.ExecutionService
	PlanningService  *ai.PlanningService
	// Git Services
	GitManager      *git.GitManager
	WorktreeManager *worktreesvc.WorktreeManager
	// GitHub Services
	GitHubService *github.GitHubServiceV2
	PRCreator     *github.PRCreator
	// Job Services
	JobClient        *jobs.Client
	JobClientAdapter usecase.JobClientInterface
	JobProcessor     *jobs.Processor
}

// NewApp creates a new App instance
func NewApp(
	cfg *config.Config,
	gormDB *database.GormDB,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	planRepo repository.PlanRepository,
	worktreeRepo repository.WorktreeRepository,
	auditRepo repository.AuditRepository,
	executionRepo repository.ExecutionRepository,
	executionLogRepo repository.ExecutionLogRepository,
	pullRequestRepo repository.PullRequestRepository,
	auditUsecase usecase.AuditUsecase,
	projectUsecase usecase.ProjectUsecase,
	taskUsecase usecase.TaskUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
	notificationUsecase usecase.NotificationUsecase,
	executionUsecase usecase.ExecutionUsecase,
	wsService *websocket.Service,
	cliManager *ai.CLIManager,
	processManager *ai.ProcessManager,
	executionService *ai.ExecutionService,
	planningService *ai.PlanningService,
	gitManager *git.GitManager,
	worktreeManager *worktreesvc.WorktreeManager,
	prCreator *github.PRCreator,
	jobClient *jobs.Client,
	jobClientAdapter usecase.JobClientInterface,
	jobProcessor *jobs.Processor,
) *App {
	return &App{
		Config:              cfg,
		GormDB:              gormDB,
		ProjectRepo:         projectRepo,
		TaskRepo:            taskRepo,
		PlanRepo:            planRepo,
		WorktreeRepo:        worktreeRepo,
		AuditRepo:           auditRepo,
		ExecutionRepo:       executionRepo,
		ExecutionLogRepo:    executionLogRepo,
		PullRequestRepo:     pullRequestRepo,
		AuditUsecase:        auditUsecase,
		ProjectUsecase:      projectUsecase,
		TaskUsecase:         taskUsecase,
		WorktreeUsecase:     worktreeUsecase,
		NotificationUsecase: notificationUsecase,
		ExecutionUsecase:    executionUsecase,
		WebSocketService:    wsService,
		CLIManager:          cliManager,
		ProcessManager:      processManager,
		ExecutionService:    executionService,
		PlanningService:     planningService,
		GitManager:          gitManager,
		WorktreeManager:     worktreeManager,
		PRCreator:           prCreator,
		JobClient:           jobClient,
		JobClientAdapter:    jobClientAdapter,
		JobProcessor:        jobProcessor,
	}
}

// ProvideGormDB provides a GORM database connection
func ProvideGormDB(cfg *config.Config) (*database.GormDB, error) {
	return database.NewGormDB(cfg)
}

// ProvideWorktreeRepository provides a WorktreeRepository instance
func ProvideWorktreeRepository(gormDB *database.GormDB) repository.WorktreeRepository {
	return postgres.NewWorktreeRepository(gormDB)
}

// ProvideAuditService provides an AuditService instance
func ProvideAuditUsecase(auditRepo repository.AuditRepository) usecase.AuditUsecase {
	return usecase.NewAuditUsecase(auditRepo)
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
	integratedConfig := &worktreesvc.IntegratedConfig{
		Worktree: &cfg.Worktree,
		Git:      &git.ManagerConfig{},
	}
	return worktreesvc.NewIntegratedWorktreeService(integratedConfig)
}

// ProvideProjectGitService provides a ProjectGitService instance
func ProvideProjectGitService(gitManager *git.GitManager) git.ProjectGitServiceInterface {
	return git.NewProjectGitService(gitManager)
}

// ProvideProjectUsecase provides a ProjectUsecase instance
func ProvideProjectUsecase(projectRepo repository.ProjectRepository, auditUsecase usecase.AuditUsecase, gitService git.ProjectGitServiceInterface) usecase.ProjectUsecase {
	return usecase.NewProjectUsecase(projectRepo, auditUsecase, gitService)
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
	pullRequestRepo repository.PullRequestRepository,
	projectRepo repository.ProjectRepository,
	planRepo repository.PlanRepository,
	notificationUsecase usecase.NotificationUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
	jobClient usecase.JobClientInterface,
) usecase.TaskUsecase {
	return usecase.NewTaskUsecase(taskRepo, pullRequestRepo, projectRepo, planRepo, notificationUsecase, worktreeUsecase, jobClient)
}

// ProvideCLIManager provides a CLIManager instance
func ProvideCLIManager() (*ai.CLIManager, error) {
	config := &ai.CLIConfig{
		CLICommand:       "claude-code",
		Timeout:          300 * time.Second, // 5 minutes
		WorkingDirectory: "",
		EnableLogging:    true,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	}
	return ai.NewCLIManager(config)
}

// ProvideProcessManager provides a ProcessManager instance
func ProvideProcessManager() *ai.ProcessManager {
	return ai.NewProcessManager()
}

// ProvideExecutionService provides an ExecutionService instance
func ProvideExecutionService(cliManager *ai.CLIManager, processManager *ai.ProcessManager) *ai.ExecutionService {
	return ai.NewExecutionService(cliManager, processManager)
}

// ProvidePlanningService provides a PlanningService instance
func ProvidePlanningService(executionService *ai.ExecutionService, cliManager *ai.CLIManager) *ai.PlanningService {
	return ai.NewPlanningService(executionService, cliManager)
}

// ProvideJobClient provides a JobClient instance
func ProvideJobClient(cfg *config.Config) *jobs.Client {
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	return jobs.NewClient(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
}

// ProvideJobClientAdapter provides a JobClientAdapter instance
func ProvideJobClientAdapter(client *jobs.Client) usecase.JobClientInterface {
	return jobs.NewJobClientAdapter(client)
}

// ProvideWorktreeManager provides a WorktreeManager instance
func ProvideWorktreeManager(cfg *config.Config) (*worktreesvc.WorktreeManager, error) {
	return worktreesvc.NewWorktreeManager(&cfg.Worktree)
}

// ProvideJobProcessor provides a Processor instance
func ProvideJobProcessor(
	taskUsecase usecase.TaskUsecase,
	projectUsecase usecase.ProjectUsecase,
	worktreeUsecase usecase.WorktreeUsecase,
	planningService *ai.PlanningService,
	executionService *ai.ExecutionService,
	planRepo repository.PlanRepository,
	executionRepo repository.ExecutionRepository,
	executionLogRepo repository.ExecutionLogRepository,
	wsService *websocket.Service,
	gitManager *git.GitManager,
	prCreator *github.PRCreator,
	prRepo repository.PullRequestRepository,
	githubService github.GitHubServiceInterface,
) *jobs.Processor {
	return jobs.NewProcessor(taskUsecase, projectUsecase, worktreeUsecase, planningService, executionService, planRepo, executionRepo, executionLogRepo, wsService, gitManager, prCreator, prRepo, githubService)
}

// ProvideWebSocketService provides a WebSocket service instance
func ProvideWebSocketService(cfg *config.Config) *websocket.Service {
	return websocket.NewService(&cfg.CentrifugeRedisBroker)
}

func ProvideExecutionUsecase(executionRepo repository.ExecutionRepository, executionLogRepo repository.ExecutionLogRepository, taskRepo repository.TaskRepository) usecase.ExecutionUsecase {
	return usecase.NewExecutionUsecase(executionRepo, executionLogRepo, taskRepo)
}

// ProvideGitHubService provides a GitHub service instance
func ProvideGitHubService(cfg *config.Config) github.GitHubServiceInterface {
	githubConfig := &github.GitHubConfig{
		Token:     cfg.GitHub.Token,
		BaseURL:   cfg.GitHub.BaseURL,
		UserAgent: cfg.GitHub.UserAgent,
		Timeout:   cfg.GitHub.Timeout,
	}
	return github.NewGitHubServiceV2(githubConfig)
}

// ProvidePRCreator provides a PR creator instance
func ProvidePRCreator(githubService github.GitHubServiceInterface, cfg *config.Config) *github.PRCreator {
	baseURL := cfg.App.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8098" // fallback for development
	}
	return github.NewPRCreator(githubService, baseURL)
}

// ProvidePullRequestRepository provides a PullRequestRepository instance
func ProvidePullRequestRepository(gormDB *database.GormDB) repository.PullRequestRepository {
	return postgres.NewPullRequestRepository(gormDB)
}
