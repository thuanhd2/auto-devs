package e2e_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestDataGenerator provides methods to generate test data
type TestDataGenerator struct {
	suite *E2ETestSuite
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(suite *E2ETestSuite) *TestDataGenerator {
	return &TestDataGenerator{suite: suite}
}

// GenerateProject creates a test project with specified configuration
func (g *TestDataGenerator) GenerateProject(config ProjectConfig) *entity.Project {
	if config.Name == "" {
		config.Name = fmt.Sprintf("Test Project %d", rand.Int())
	}
	if config.Description == "" {
		config.Description = fmt.Sprintf("Generated test project: %s", config.Name)
	}
	if config.RepositoryURL == "" {
		config.RepositoryURL = fmt.Sprintf("https://github.com/test/%s.git", 
			fmt.Sprintf("repo-%d", rand.Int()))
	}

	project := &entity.Project{
		Name:          config.Name,
		Description:   config.Description,
		RepositoryURL: config.RepositoryURL,
		Settings: entity.ProjectSettings{
			DefaultBranch:      config.DefaultBranch,
			AutoMerge:          config.AutoMerge,
			RequireApproval:    config.RequireApproval,
			MaxConcurrentTasks: config.MaxConcurrentTasks,
		},
	}

	if config.DefaultBranch == "" {
		project.Settings.DefaultBranch = "main"
	}
	if config.MaxConcurrentTasks == 0 {
		project.Settings.MaxConcurrentTasks = 3
	}

	err := g.suite.repositories.Project.Create(g.suite.ctx, project)
	require.NoError(g.suite.t, err)

	return project
}

// GenerateTask creates a test task with specified configuration
func (g *TestDataGenerator) GenerateTask(projectID uuid.UUID, config TaskConfig) *entity.Task {
	if config.Title == "" {
		config.Title = fmt.Sprintf("Test Task %d", rand.Int())
	}
	if config.Description == "" {
		config.Description = fmt.Sprintf("Generated test task: %s", config.Title)
	}
	if config.Status == "" {
		config.Status = entity.TaskStatusTODO
	}
	if config.Priority == "" {
		config.Priority = entity.TaskPriorityMedium
	}
	if config.GitStatus == "" {
		config.GitStatus = entity.TaskGitStatusNone
	}

	task := &entity.Task{
		ProjectID:   projectID,
		Title:       config.Title,
		Description: config.Description,
		Status:      config.Status,
		Priority:    config.Priority,
		GitStatus:   config.GitStatus,
		Tags:        config.Tags,
		Metadata:    config.Metadata,
	}

	if config.AssigneeID != uuid.Nil {
		task.AssigneeID = &config.AssigneeID
	}

	if !config.DueDate.IsZero() {
		task.DueDate = &config.DueDate
	}

	err := g.suite.repositories.Task.Create(g.suite.ctx, task)
	require.NoError(g.suite.t, err)

	return task
}

// GenerateExecution creates a test execution with specified configuration
func (g *TestDataGenerator) GenerateExecution(taskID uuid.UUID, config ExecutionConfig) *entity.Execution {
	if config.Type == "" {
		config.Type = entity.ExecutionTypePlanning
	}
	if config.Status == "" {
		config.Status = entity.ExecutionStatusRunning
	}

	execution := &entity.Execution{
		ID:        uuid.New(),
		TaskID:    taskID,
		Type:      config.Type,
		Status:    config.Status,
		StartedAt: time.Now(),
		Config:    config.Config,
		Metadata:  config.Metadata,
	}

	if !config.CompletedAt.IsZero() {
		execution.CompletedAt = &config.CompletedAt
		if config.Status == entity.ExecutionStatusRunning {
			execution.Status = entity.ExecutionStatusCompleted
		}
	}

	if config.Error != "" {
		execution.Error = &config.Error
		if config.Status == entity.ExecutionStatusRunning {
			execution.Status = entity.ExecutionStatusFailed
		}
	}

	err := g.suite.repositories.Execution.Create(g.suite.ctx, execution)
	require.NoError(g.suite.t, err)

	return execution
}

// GeneratePlan creates a test plan with specified configuration
func (g *TestDataGenerator) GeneratePlan(taskID uuid.UUID, config PlanConfig) *entity.Plan {
	if config.Title == "" {
		config.Title = fmt.Sprintf("Generated Plan for Task %s", taskID.String())
	}
	if config.Status == "" {
		config.Status = entity.PlanStatusDraft
	}

	// Generate default plan content if not provided
	if config.Content == nil {
		config.Content = g.generateDefaultPlanContent(config.Title)
	}

	plan := &entity.Plan{
		ID:       uuid.New(),
		TaskID:   taskID,
		Title:    config.Title,
		Content:  config.Content,
		Status:   config.Status,
		Metadata: config.Metadata,
	}

	if config.ApprovedBy != uuid.Nil {
		plan.ApprovedBy = &config.ApprovedBy
	}

	if !config.ApprovedAt.IsZero() {
		plan.ApprovedAt = &config.ApprovedAt
	}

	err := g.suite.repositories.Plan.Create(g.suite.ctx, plan)
	require.NoError(g.suite.t, err)

	return plan
}

// GenerateWorktree creates a test worktree with specified configuration
func (g *TestDataGenerator) GenerateWorktree(projectID, taskID uuid.UUID, config WorktreeConfig) *entity.Worktree {
	if config.Branch == "" {
		config.Branch = fmt.Sprintf("task-%s", taskID.String()[:8])
	}
	if config.Path == "" {
		config.Path = fmt.Sprintf("/tmp/worktree-%s", uuid.New().String())
	}
	if config.Status == "" {
		config.Status = entity.WorktreeStatusActive
	}

	worktree := &entity.Worktree{
		ID:        uuid.New(),
		ProjectID: projectID,
		TaskID:    taskID,
		Branch:    config.Branch,
		Path:      config.Path,
		Status:    config.Status,
		CreatedAt: time.Now(),
	}

	if !config.DeletedAt.IsZero() {
		worktree.DeletedAt = &config.DeletedAt
		if config.Status == entity.WorktreeStatusActive {
			worktree.Status = entity.WorktreeStatusInactive
		}
	}

	err := g.suite.repositories.Worktree.Create(g.suite.ctx, worktree)
	require.NoError(g.suite.t, err)

	return worktree
}

// GeneratePullRequest creates a test pull request with specified configuration
func (g *TestDataGenerator) GeneratePullRequest(taskID uuid.UUID, config PullRequestConfig) *entity.PullRequest {
	if config.Title == "" {
		config.Title = fmt.Sprintf("Pull Request for Task %s", taskID.String()[:8])
	}
	if config.Number == 0 {
		config.Number = rand.Intn(1000) + 1
	}
	if config.HeadBranch == "" {
		config.HeadBranch = fmt.Sprintf("task-%s", taskID.String()[:8])
	}
	if config.BaseBranch == "" {
		config.BaseBranch = "main"
	}
	if config.State == "" {
		config.State = "open"
	}

	pr := &entity.PullRequest{
		ID:         uuid.New(),
		TaskID:     &taskID,
		Number:     config.Number,
		Title:      config.Title,
		Body:       config.Body,
		HeadBranch: config.HeadBranch,
		BaseBranch: config.BaseBranch,
		State:      config.State,
		HTMLURL:    fmt.Sprintf("https://github.com/test/repo/pull/%d", config.Number),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if config.Merged {
		pr.State = "closed"
		now := time.Now()
		pr.MergedAt = &now
	}

	err := g.suite.repositories.PullRequest.Create(g.suite.ctx, pr)
	require.NoError(g.suite.t, err)

	return pr
}

// GenerateCompleteTaskFlow creates a complete task flow with all related entities
func (g *TestDataGenerator) GenerateCompleteTaskFlow(config CompleteFlowConfig) *CompleteTaskFlow {
	// Create project
	project := g.GenerateProject(config.Project)

	// Create task
	task := g.GenerateTask(project.ID, config.Task)

	// Create plan if specified
	var plan *entity.Plan
	if config.IncludePlan {
		plan = g.GeneratePlan(task.ID, config.Plan)
	}

	// Create worktree if specified
	var worktree *entity.Worktree
	if config.IncludeWorktree {
		worktree = g.GenerateWorktree(project.ID, task.ID, config.Worktree)
	}

	// Create executions if specified
	var executions []*entity.Execution
	if config.IncludeExecutions {
		for _, execConfig := range config.Executions {
			execution := g.GenerateExecution(task.ID, execConfig)
			executions = append(executions, execution)
		}
	}

	// Create pull request if specified
	var pullRequest *entity.PullRequest
	if config.IncludePullRequest {
		pullRequest = g.GeneratePullRequest(task.ID, config.PullRequest)
	}

	return &CompleteTaskFlow{
		Project:     project,
		Task:        task,
		Plan:        plan,
		Worktree:    worktree,
		Executions:  executions,
		PullRequest: pullRequest,
	}
}

// generateDefaultPlanContent generates default plan content for testing
func (g *TestDataGenerator) generateDefaultPlanContent(title string) map[string]interface{} {
	return map[string]interface{}{
		"title":       title,
		"description": "This is a generated test plan for end-to-end testing",
		"objective":   "Test the complete task automation workflow",
		"approach":    "Implement the required functionality step by step",
		"steps": []map[string]interface{}{
			{
				"id":          1,
				"title":       "Analysis and Design",
				"description": "Analyze requirements and design the solution",
				"files":       []string{"docs/design.md"},
				"estimated_time": "30 minutes",
			},
			{
				"id":          2,
				"title":       "Implementation",
				"description": "Implement the core functionality",
				"files":       []string{"src/main.go", "src/handler.go"},
				"estimated_time": "2 hours",
			},
			{
				"id":          3,
				"title":       "Testing",
				"description": "Write and run tests",
				"files":       []string{"src/main_test.go", "src/handler_test.go"},
				"estimated_time": "1 hour",
			},
			{
				"id":          4,
				"title":       "Documentation",
				"description": "Update documentation",
				"files":       []string{"README.md", "docs/api.md"},
				"estimated_time": "30 minutes",
			},
		},
		"dependencies": []string{},
		"risks": []string{
			"Potential integration issues with external APIs",
			"Performance concerns with large datasets",
		},
		"success_criteria": []string{
			"All tests pass",
			"Documentation is updated",
			"Code review is completed",
		},
	}
}

// Configuration structs for test data generation

// ProjectConfig configures project generation
type ProjectConfig struct {
	Name               string
	Description        string
	RepositoryURL      string
	DefaultBranch      string
	AutoMerge          bool
	RequireApproval    bool
	MaxConcurrentTasks int
}

// TaskConfig configures task generation
type TaskConfig struct {
	Title       string
	Description string
	Status      entity.TaskStatus
	Priority    entity.TaskPriority
	GitStatus   entity.TaskGitStatus
	AssigneeID  uuid.UUID
	DueDate     time.Time
	Tags        []string
	Metadata    map[string]interface{}
}

// ExecutionConfig configures execution generation
type ExecutionConfig struct {
	Type        entity.ExecutionType
	Status      entity.ExecutionStatus
	CompletedAt time.Time
	Error       string
	Config      map[string]interface{}
	Metadata    map[string]interface{}
}

// PlanConfig configures plan generation
type PlanConfig struct {
	Title      string
	Content    map[string]interface{}
	Status     entity.PlanStatus
	ApprovedBy uuid.UUID
	ApprovedAt time.Time
	Metadata   map[string]interface{}
}

// WorktreeConfig configures worktree generation
type WorktreeConfig struct {
	Branch    string
	Path      string
	Status    entity.WorktreeStatus
	DeletedAt time.Time
}

// PullRequestConfig configures pull request generation
type PullRequestConfig struct {
	Number     int
	Title      string
	Body       string
	HeadBranch string
	BaseBranch string
	State      string
	Merged     bool
}

// CompleteFlowConfig configures complete flow generation
type CompleteFlowConfig struct {
	Project            ProjectConfig
	Task               TaskConfig
	Plan               PlanConfig
	Worktree           WorktreeConfig
	PullRequest        PullRequestConfig
	Executions         []ExecutionConfig
	IncludePlan        bool
	IncludeWorktree    bool
	IncludePullRequest bool
	IncludeExecutions  bool
}

// CompleteTaskFlow represents a complete task flow with all related entities
type CompleteTaskFlow struct {
	Project     *entity.Project
	Task        *entity.Task
	Plan        *entity.Plan
	Worktree    *entity.Worktree
	Executions  []*entity.Execution
	PullRequest *entity.PullRequest
}

// CreateBulkTestData creates bulk test data for performance testing
func (g *TestDataGenerator) CreateBulkTestData(config BulkDataConfig) *BulkTestData {
	data := &BulkTestData{
		Projects: make([]*entity.Project, 0, config.ProjectCount),
		Tasks:    make([]*entity.Task, 0, config.ProjectCount*config.TasksPerProject),
	}

	for i := 0; i < config.ProjectCount; i++ {
		// Create project
		projectConfig := ProjectConfig{
			Name:        fmt.Sprintf("Bulk Test Project %d", i+1),
			Description: fmt.Sprintf("Bulk test project %d for performance testing", i+1),
		}
		project := g.GenerateProject(projectConfig)
		data.Projects = append(data.Projects, project)

		// Create tasks for project
		for j := 0; j < config.TasksPerProject; j++ {
			taskConfig := TaskConfig{
				Title:       fmt.Sprintf("Bulk Task %d-%d", i+1, j+1),
				Description: fmt.Sprintf("Bulk task %d for project %d", j+1, i+1),
				Status:      g.randomTaskStatus(),
				Priority:    g.randomTaskPriority(),
			}
			task := g.GenerateTask(project.ID, taskConfig)
			data.Tasks = append(data.Tasks, task)
		}
	}

	return data
}

// BulkDataConfig configures bulk data generation
type BulkDataConfig struct {
	ProjectCount     int
	TasksPerProject  int
	ExecutionsPerTask int
}

// BulkTestData represents bulk test data
type BulkTestData struct {
	Projects   []*entity.Project
	Tasks      []*entity.Task
	Executions []*entity.Execution
}

// Helper functions for random data generation

func (g *TestDataGenerator) randomTaskStatus() entity.TaskStatus {
	statuses := []entity.TaskStatus{
		entity.TaskStatusTODO,
		entity.TaskStatusPLANNING,
		entity.TaskStatusPLANREVIEWING,
		entity.TaskStatusIMPLEMENTING,
		entity.TaskStatusCODEREVIEWING,
		entity.TaskStatusDONE,
	}
	return statuses[rand.Intn(len(statuses))]
}

func (g *TestDataGenerator) randomTaskPriority() entity.TaskPriority {
	priorities := []entity.TaskPriority{
		entity.TaskPriorityLow,
		entity.TaskPriorityMedium,
		entity.TaskPriorityHigh,
		entity.TaskPriorityUrgent,
	}
	return priorities[rand.Intn(len(priorities))]
}

// Cleanup removes all test data created by the generator
func (g *TestDataGenerator) Cleanup(ctx context.Context) error {
	// Clean up in reverse dependency order
	
	// Delete pull requests
	if prs, err := g.suite.repositories.PullRequest.List(ctx, entity.PullRequestFilters{}); err == nil {
		for _, pr := range prs {
			g.suite.repositories.PullRequest.Delete(ctx, pr.ID)
		}
	}

	// Delete executions
	if executions, err := g.suite.repositories.Execution.List(ctx, entity.ExecutionFilters{}); err == nil {
		for _, execution := range executions {
			g.suite.repositories.Execution.Delete(ctx, execution.ID)
		}
	}

	// Delete plans
	if plans, err := g.suite.repositories.Plan.List(ctx, entity.PlanFilters{}); err == nil {
		for _, plan := range plans {
			g.suite.repositories.Plan.Delete(ctx, plan.ID)
		}
	}

	// Delete worktrees
	if worktrees, err := g.suite.repositories.Worktree.List(ctx, entity.WorktreeFilters{}); err == nil {
		for _, worktree := range worktrees {
			g.suite.repositories.Worktree.Delete(ctx, worktree.ID)
		}
	}

	// Delete tasks
	if tasks, err := g.suite.repositories.Task.List(ctx, entity.TaskFilters{}); err == nil {
		for _, task := range tasks {
			g.suite.repositories.Task.Delete(ctx, task.ID)
		}
	}

	// Delete projects
	if projects, err := g.suite.repositories.Project.List(ctx, entity.ProjectFilters{}); err == nil {
		for _, project := range projects {
			g.suite.repositories.Project.Delete(ctx, project.ID)
		}
	}

	return nil
}