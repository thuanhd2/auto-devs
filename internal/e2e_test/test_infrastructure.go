package e2e_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/internal/di"
	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/websocket"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// E2ETestSuite represents a complete end-to-end test environment
type E2ETestSuite struct {
	t           *testing.T
	server      *httptest.Server
	db          *database.GormDB
	wsServer    *websocket.Server
	repositories *TestRepositories
	services     *TestServices
	handlers     *TestHandlers
	cleanup     []func() error
	ctx         context.Context
	cancel      context.CancelFunc
}

// TestRepositories contains all repository instances for testing
type TestRepositories struct {
	Project      repository.ProjectRepository
	Task         repository.TaskRepository
	Execution    repository.ExecutionRepository
	ExecutionLog repository.ExecutionLogRepository
	Plan         repository.PlanRepository
	PullRequest  repository.PullRequestRepository
	Worktree     repository.WorktreeRepository
	Process      repository.ProcessRepository
	Audit        repository.AuditRepository
}

// TestServices contains all service instances for testing
type TestServices struct {
	GitManager        GitManagerMock
	WorktreeService   WorktreeServiceMock
	GitHubService     GitHubServiceMock
	AIPlanning        AIServiceMock
	AIExecution       AIServiceMock
	ProcessManager    ProcessManagerMock
	NotificationHub   *websocket.Hub
}

// TestHandlers contains all handler instances for testing
type TestHandlers struct {
	ProjectHandler   *handler.ProjectHandler
	TaskHandler      *handler.TaskHandler
	ExecutionHandler *handler.ExecutionHandler
	WorktreeHandler  *handler.WorktreeHandler
	WSHandler        *websocket.Handler
}

// TestData contains commonly used test data
type TestData struct {
	Projects []*entity.Project
	Tasks    []*entity.Task
	Users    []*TestUser
}

// TestUser represents a test user for authentication scenarios
type TestUser struct {
	ID       uuid.UUID
	Username string
	Email    string
	Token    string
}

// NewE2ETestSuite creates a new end-to-end test suite
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	suite := &E2ETestSuite{
		t:       t,
		cleanup: make([]func() error, 0),
	}

	// Set up context with timeout
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Minute)

	// Set up test database
	suite.setupTestDatabase()

	// Set up repositories
	suite.setupRepositories()

	// Set up services with mocks
	suite.setupServices()

	// Set up handlers
	suite.setupHandlers()

	// Set up HTTP server
	suite.setupHTTPServer()

	// Set up WebSocket server
	suite.setupWebSocketServer()

	return suite
}

// setupTestDatabase initializes a test database with migrations
func (s *E2ETestSuite) setupTestDatabase() {
	// Get the absolute path to the project root directory
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../../")

	// Load test environment
	envFile := filepath.Join(projectRoot, ".env.test")
	if _, err := os.Stat(envFile); err == nil {
		godotenv.Load(envFile)
	}

	// Configure test database
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbUser := getEnvOrDefault("DB_USERNAME", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "password")
	dbName := getEnvOrDefault("DB_NAME", "autodevs_test")
	dbPort := getEnvOrDefault("DB_PORT", "5432")

	role := pgtestdb.DefaultRole()
	role.Capabilities = "SUPERUSER"

	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       dbHost,
		Port:       dbPort,
		User:       dbUser,
		Password:   dbPassword,
		Database:   dbName,
		Options:    "sslmode=disable",
		TestRole:   &role,
	}

	// Set up migrator
	migrationsPath := filepath.Join(projectRoot, "migrations")
	migrator := golangmigrator.New(migrationsPath)

	// Create test database
	sqlDB := pgtestdb.New(s.t, pgxConf, migrator)
	s.addCleanup(func() error {
		return sqlDB.Close()
	})

	// Set up GORM
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(s.t, err)

	s.db = &database.GormDB{DB: gormDB}
}

// setupRepositories initializes all repository instances
func (s *E2ETestSuite) setupRepositories() {
	// Initialize repositories with the test database
	repos, err := di.InitializeRepositories(s.db)
	require.NoError(s.t, err)

	s.repositories = &TestRepositories{
		Project:      repos.ProjectRepository,
		Task:         repos.TaskRepository,
		Execution:    repos.ExecutionRepository,
		ExecutionLog: repos.ExecutionLogRepository,
		Plan:         repos.PlanRepository,
		PullRequest:  repos.PullRequestRepository,
		Worktree:     repos.WorktreeRepository,
		Process:      repos.ProcessRepository,
		Audit:        repos.AuditRepository,
	}
}

// setupServices initializes all service instances with mocks
func (s *E2ETestSuite) setupServices() {
	// Create mock services
	gitManager := NewGitManagerMock()
	worktreeService := NewWorktreeServiceMock()
	githubService := NewGitHubServiceMock()
	aiPlanning := NewAIServiceMock("planning")
	aiExecution := NewAIServiceMock("execution")
	processManager := NewProcessManagerMock()

	// Set up WebSocket hub
	notificationHub := websocket.NewHub()
	go notificationHub.Run()
	s.addCleanup(func() error {
		notificationHub.Shutdown()
		return nil
	})

	s.services = &TestServices{
		GitManager:        gitManager,
		WorktreeService:   worktreeService,
		GitHubService:     githubService,
		AIPlanning:        aiPlanning,
		AIExecution:       aiExecution,
		ProcessManager:    processManager,
		NotificationHub:   notificationHub,
	}
}

// setupHandlers initializes all handler instances
func (s *E2ETestSuite) setupHandlers() {
	// Use dependency injection to create handlers with test services
	cfg := config.Load()
	
	handlers, err := di.InitializeHandlers(
		s.repositories,
		s.services,
		cfg,
	)
	require.NoError(s.t, err)

	s.handlers = &TestHandlers{
		ProjectHandler:   handlers.ProjectHandler,
		TaskHandler:      handlers.TaskHandler,
		ExecutionHandler: handlers.ExecutionHandler,
		WorktreeHandler:  handlers.WorktreeHandler,
		WSHandler:        handlers.WSHandler,
	}
}

// setupHTTPServer creates a test HTTP server
func (s *E2ETestSuite) setupHTTPServer() {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up routes
	v1 := router.Group("/api/v1")
	{
		// Project routes
		projects := v1.Group("/projects")
		{
			projects.GET("", s.handlers.ProjectHandler.List)
			projects.POST("", s.handlers.ProjectHandler.Create)
			projects.GET("/:id", s.handlers.ProjectHandler.GetByID)
			projects.PUT("/:id", s.handlers.ProjectHandler.Update)
			projects.DELETE("/:id", s.handlers.ProjectHandler.Delete)
		}

		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", s.handlers.TaskHandler.List)
			tasks.POST("", s.handlers.TaskHandler.Create)
			tasks.GET("/:id", s.handlers.TaskHandler.GetByID)
			tasks.PUT("/:id", s.handlers.TaskHandler.Update)
			tasks.DELETE("/:id", s.handlers.TaskHandler.Delete)
			tasks.POST("/:id/plan", s.handlers.TaskHandler.StartPlanning)
			tasks.POST("/:id/implement", s.handlers.TaskHandler.StartImplementation)
			tasks.POST("/:id/complete", s.handlers.TaskHandler.CompleteTask)
		}

		// Execution routes
		executions := v1.Group("/executions")
		{
			executions.GET("", s.handlers.ExecutionHandler.List)
			executions.GET("/:id", s.handlers.ExecutionHandler.GetByID)
			executions.POST("/:id/cancel", s.handlers.ExecutionHandler.Cancel)
		}

		// Worktree routes
		worktrees := v1.Group("/worktrees")
		{
			worktrees.GET("", s.handlers.WorktreeHandler.List)
			worktrees.POST("", s.handlers.WorktreeHandler.Create)
			worktrees.DELETE("/:id", s.handlers.WorktreeHandler.Delete)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", s.handlers.WSHandler.HandleWebSocket)

	s.server = httptest.NewServer(router)
	s.addCleanup(func() error {
		s.server.Close()
		return nil
	})
}

// setupWebSocketServer initializes WebSocket test capabilities
func (s *E2ETestSuite) setupWebSocketServer() {
	// WebSocket server is handled by the HTTP server
	// Additional WebSocket-specific test utilities can be added here
}

// GetServerURL returns the test server URL
func (s *E2ETestSuite) GetServerURL() string {
	return s.server.URL
}

// GetWebSocketURL returns the WebSocket endpoint URL
func (s *E2ETestSuite) GetWebSocketURL() string {
	return fmt.Sprintf("ws%s/ws", s.server.URL[4:]) // Replace http with ws
}

// GetDB returns the test database instance
func (s *E2ETestSuite) GetDB() *database.GormDB {
	return s.db
}

// GetContext returns the test context
func (s *E2ETestSuite) GetContext() context.Context {
	return s.ctx
}

// CreateTestData creates common test data
func (s *E2ETestSuite) CreateTestData() *TestData {
	data := &TestData{
		Projects: make([]*entity.Project, 0),
		Tasks:    make([]*entity.Task, 0),
		Users:    make([]*TestUser, 0),
	}

	// Create test projects
	for i := 0; i < 3; i++ {
		project := &entity.Project{
			Name:          fmt.Sprintf("Test Project %d", i+1),
			Description:   fmt.Sprintf("Test project description %d", i+1),
			RepositoryURL: fmt.Sprintf("https://github.com/test/repo%d.git", i+1),
			Settings: entity.ProjectSettings{
				DefaultBranch:    "main",
				AutoMerge:        true,
				RequireApproval:  false,
				MaxConcurrentTasks: 3,
			},
		}
		err := s.repositories.Project.Create(s.ctx, project)
		require.NoError(s.t, err)
		data.Projects = append(data.Projects, project)
	}

	// Create test tasks for each project
	for _, project := range data.Projects {
		for i := 0; i < 5; i++ {
			task := &entity.Task{
				ProjectID:   project.ID,
				Title:       fmt.Sprintf("Test Task %d", i+1),
				Description: fmt.Sprintf("Test task description %d", i+1),
				Status:      entity.TaskStatusTODO,
				Priority:    entity.TaskPriorityMedium,
				GitStatus:   entity.TaskGitStatusNone,
			}
			err := s.repositories.Task.Create(s.ctx, task)
			require.NoError(s.t, err)
			data.Tasks = append(data.Tasks, task)
		}
	}

	// Create test users
	for i := 0; i < 3; i++ {
		user := &TestUser{
			ID:       uuid.New(),
			Username: fmt.Sprintf("testuser%d", i+1),
			Email:    fmt.Sprintf("testuser%d@example.com", i+1),
			Token:    fmt.Sprintf("test-token-%d", i+1),
		}
		data.Users = append(data.Users, user)
	}

	return data
}

// WaitForTaskStatus waits for a task to reach a specific status
func (s *E2ETestSuite) WaitForTaskStatus(taskID uuid.UUID, expectedStatus entity.TaskStatus, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		task, err := s.repositories.Task.GetByID(s.ctx, taskID)
		if err == nil && task.Status == expectedStatus {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// WaitForExecutionStatus waits for an execution to reach a specific status
func (s *E2ETestSuite) WaitForExecutionStatus(executionID uuid.UUID, expectedStatus entity.ExecutionStatus, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		execution, err := s.repositories.Execution.GetByID(s.ctx, executionID)
		if err == nil && execution.Status == expectedStatus {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// addCleanup adds a cleanup function to be called during teardown
func (s *E2ETestSuite) addCleanup(cleanup func() error) {
	s.cleanup = append(s.cleanup, cleanup)
}

// Teardown cleans up all test resources
func (s *E2ETestSuite) Teardown() {
	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	// Run cleanup functions in reverse order
	for i := len(s.cleanup) - 1; i >= 0; i-- {
		if err := s.cleanup[i](); err != nil {
			s.t.Logf("Cleanup error: %v", err)
		}
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}