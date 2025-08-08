package e2e_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/service/ai"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// GitManagerMock mocks Git operations for testing
type GitManagerMock struct {
	mock.Mock
	repositories map[string]*MockRepository
	mu           sync.RWMutex
}

// MockRepository represents a mock Git repository
type MockRepository struct {
	URL      string
	Branches map[string]*MockBranch
	Commits  []*MockCommit
}

// MockBranch represents a mock Git branch
type MockBranch struct {
	Name     string
	Commits  []*MockCommit
	IsActive bool
}

// MockCommit represents a mock Git commit
type MockCommit struct {
	Hash    string
	Message string
	Author  string
	Time    time.Time
}

// NewGitManagerMock creates a new Git manager mock
func NewGitManagerMock() GitManagerMock {
	return GitManagerMock{
		repositories: make(map[string]*MockRepository),
	}
}

// CloneRepository mocks repository cloning
func (m *GitManagerMock) CloneRepository(ctx context.Context, url, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, url, path)
	
	// Create mock repository
	m.repositories[url] = &MockRepository{
		URL:      url,
		Branches: map[string]*MockBranch{
			"main": {
				Name:     "main",
				Commits:  []*MockCommit{},
				IsActive: true,
			},
		},
		Commits: []*MockCommit{},
	}

	return args.Error(0)
}

// CreateBranch mocks branch creation
func (m *GitManagerMock) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, repoPath, branchName)
	
	// Find repository and create branch
	for _, repo := range m.repositories {
		if repo.Branches == nil {
			repo.Branches = make(map[string]*MockBranch)
		}
		repo.Branches[branchName] = &MockBranch{
			Name:     branchName,
			Commits:  []*MockCommit{},
			IsActive: false,
		}
		break
	}

	return args.Error(0)
}

// CheckoutBranch mocks branch checkout
func (m *GitManagerMock) CheckoutBranch(ctx context.Context, repoPath, branchName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, repoPath, branchName)
	
	// Find repository and set active branch
	for _, repo := range m.repositories {
		for name, branch := range repo.Branches {
			branch.IsActive = (name == branchName)
		}
		break
	}

	return args.Error(0)
}

// WorktreeServiceMock mocks worktree operations
type WorktreeServiceMock struct {
	mock.Mock
	worktrees map[uuid.UUID]*MockWorktree
	mu        sync.RWMutex
}

// MockWorktree represents a mock Git worktree
type MockWorktree struct {
	ID        uuid.UUID
	Path      string
	Branch    string
	IsActive  bool
	CreatedAt time.Time
}

// NewWorktreeServiceMock creates a new worktree service mock
func NewWorktreeServiceMock() WorktreeServiceMock {
	return WorktreeServiceMock{
		worktrees: make(map[uuid.UUID]*MockWorktree),
	}
}

// CreateWorktree mocks worktree creation
func (m *WorktreeServiceMock) CreateWorktree(ctx context.Context, projectID uuid.UUID, taskID uuid.UUID, branchName string) (*entity.Worktree, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, projectID, taskID, branchName)
	
	worktreeID := uuid.New()
	mockWorktree := &MockWorktree{
		ID:        worktreeID,
		Path:      fmt.Sprintf("/tmp/worktree-%s", worktreeID.String()),
		Branch:    branchName,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	
	m.worktrees[worktreeID] = mockWorktree
	
	worktree := &entity.Worktree{
		ID:        worktreeID,
		TaskID:    taskID,
		ProjectID: projectID,
		Branch:    branchName,
		Path:      mockWorktree.Path,
		Status:    entity.WorktreeStatusActive,
		CreatedAt: mockWorktree.CreatedAt,
	}

	return worktree, args.Error(0)
}

// DeleteWorktree mocks worktree deletion
func (m *WorktreeServiceMock) DeleteWorktree(ctx context.Context, worktreeID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, worktreeID)
	
	if mockWorktree, exists := m.worktrees[worktreeID]; exists {
		mockWorktree.IsActive = false
		delete(m.worktrees, worktreeID)
	}

	return args.Error(0)
}

// GitHubServiceMock mocks GitHub API operations
type GitHubServiceMock struct {
	mock.Mock
	pullRequests map[int]*MockPullRequest
	webhooks     []MockWebhookEvent
	mu           sync.RWMutex
}

// MockPullRequest represents a mock GitHub pull request
type MockPullRequest struct {
	Number      int
	Title       string
	Body        string
	HeadBranch  string
	BaseBranch  string
	State       string
	Merged      bool
	MergedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MockWebhookEvent represents a mock GitHub webhook event
type MockWebhookEvent struct {
	Type      string
	Action    string
	PR        *MockPullRequest
	Timestamp time.Time
}

// NewGitHubServiceMock creates a new GitHub service mock
func NewGitHubServiceMock() GitHubServiceMock {
	return GitHubServiceMock{
		pullRequests: make(map[int]*MockPullRequest),
		webhooks:     make([]MockWebhookEvent, 0),
	}
}

// CreatePullRequest mocks PR creation
func (m *GitHubServiceMock) CreatePullRequest(ctx context.Context, repoOwner, repoName, title, body, headBranch, baseBranch string) (*entity.PullRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, repoOwner, repoName, title, body, headBranch, baseBranch)
	
	prNumber := len(m.pullRequests) + 1
	mockPR := &MockPullRequest{
		Number:     prNumber,
		Title:      title,
		Body:       body,
		HeadBranch: headBranch,
		BaseBranch: baseBranch,
		State:      "open",
		Merged:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	m.pullRequests[prNumber] = mockPR
	
	pr := &entity.PullRequest{
		ID:         uuid.New(),
		Number:     prNumber,
		Title:      title,
		Body:       body,
		HeadBranch: headBranch,
		BaseBranch: baseBranch,
		State:      "open",
		HTMLURL:    fmt.Sprintf("https://github.com/%s/%s/pull/%d", repoOwner, repoName, prNumber),
		CreatedAt:  mockPR.CreatedAt,
		UpdatedAt:  mockPR.UpdatedAt,
	}

	return pr, args.Error(0)
}

// SimulatePRMerge simulates a PR merge event for testing
func (m *GitHubServiceMock) SimulatePRMerge(prNumber int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pr, exists := m.pullRequests[prNumber]; exists {
		now := time.Now()
		pr.State = "closed"
		pr.Merged = true
		pr.MergedAt = &now
		pr.UpdatedAt = now

		// Generate webhook event
		event := MockWebhookEvent{
			Type:      "pull_request",
			Action:    "closed",
			PR:        pr,
			Timestamp: now,
		}
		m.webhooks = append(m.webhooks, event)
	}
}

// AIServiceMock mocks AI service operations
type AIServiceMock struct {
	mock.Mock
	serviceType string // "planning" or "execution"
	processes   map[uuid.UUID]*MockAIProcess
	mu          sync.RWMutex
}

// MockAIProcess represents a mock AI process
type MockAIProcess struct {
	ID          uuid.UUID
	Type        string
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
	Result      interface{}
	Error       error
}

// NewAIServiceMock creates a new AI service mock
func NewAIServiceMock(serviceType string) AIServiceMock {
	return AIServiceMock{
		serviceType: serviceType,
		processes:   make(map[uuid.UUID]*MockAIProcess),
	}
}

// StartPlanning mocks planning process start
func (m *AIServiceMock) StartPlanning(ctx context.Context, task *entity.Task) (*entity.Execution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, task)
	
	processID := uuid.New()
	mockProcess := &MockAIProcess{
		ID:        processID,
		Type:      "planning",
		Status:    "running",
		StartedAt: time.Now(),
	}
	
	m.processes[processID] = mockProcess
	
	execution := &entity.Execution{
		ID:        processID,
		TaskID:    task.ID,
		Type:      entity.ExecutionTypePlanning,
		Status:    entity.ExecutionStatusRunning,
		StartedAt: mockProcess.StartedAt,
	}

	// Simulate async completion
	go m.simulateProcessCompletion(processID, 2*time.Second)

	return execution, args.Error(0)
}

// StartImplementation mocks implementation process start
func (m *AIServiceMock) StartImplementation(ctx context.Context, task *entity.Task) (*entity.Execution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, task)
	
	processID := uuid.New()
	mockProcess := &MockAIProcess{
		ID:        processID,
		Type:      "implementation",
		Status:    "running",
		StartedAt: time.Now(),
	}
	
	m.processes[processID] = mockProcess
	
	execution := &entity.Execution{
		ID:        processID,
		TaskID:    task.ID,
		Type:      entity.ExecutionTypeImplementation,
		Status:    entity.ExecutionStatusRunning,
		StartedAt: mockProcess.StartedAt,
	}

	// Simulate async completion
	go m.simulateProcessCompletion(processID, 5*time.Second)

	return execution, args.Error(0)
}

// simulateProcessCompletion simulates async process completion
func (m *AIServiceMock) simulateProcessCompletion(processID uuid.UUID, delay time.Duration) {
	time.Sleep(delay)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if process, exists := m.processes[processID]; exists {
		now := time.Now()
		process.Status = "completed"
		process.CompletedAt = &now
		
		// Generate appropriate result based on process type
		if process.Type == "planning" {
			process.Result = map[string]interface{}{
				"plan": map[string]interface{}{
					"title":       "Generated Plan",
					"description": "This is a test plan generated by the mock AI service",
					"steps": []map[string]interface{}{
						{
							"title":       "Step 1",
							"description": "First implementation step",
							"files":       []string{"src/file1.go"},
						},
						{
							"title":       "Step 2", 
							"description": "Second implementation step",
							"files":       []string{"src/file2.go"},
						},
					},
				},
			}
		} else if process.Type == "implementation" {
			process.Result = map[string]interface{}{
				"implementation": map[string]interface{}{
					"status":        "completed",
					"files_changed": []string{"src/file1.go", "src/file2.go"},
					"tests_passed":  true,
					"summary":       "Implementation completed successfully",
				},
			}
		}
	}
}

// ProcessManagerMock mocks process management operations
type ProcessManagerMock struct {
	mock.Mock
	processes map[string]*ai.Process
	mu        sync.RWMutex
}

// NewProcessManagerMock creates a new process manager mock
func NewProcessManagerMock() ProcessManagerMock {
	return ProcessManagerMock{
		processes: make(map[string]*ai.Process),
	}
}

// SpawnProcess mocks process spawning
func (m *ProcessManagerMock) SpawnProcess(command, workingDir, environment string) (*ai.Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(command, workingDir, environment)
	
	processID := uuid.New().String()
	process := &ai.Process{
		ID:          processID,
		Command:     command,
		WorkingDir:  workingDir,
		Environment: environment,
		Status:      ai.ProcessStatusRunning,
		StartTime:   time.Now(),
		PID:         os.Getpid(), // Use current process PID for testing
	}
	
	m.processes[processID] = process
	
	// Simulate async process completion
	go func() {
		time.Sleep(1 * time.Second)
		m.mu.Lock()
		defer m.mu.Unlock()
		if p, exists := m.processes[processID]; exists {
			p.Status = ai.ProcessStatusStopped
			p.EndTime = time.Now()
		}
	}()

	return process, args.Error(0)
}

// GetProcess mocks process retrieval
func (m *ProcessManagerMock) GetProcess(processID string) (*ai.Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(processID)
	
	process, exists := m.processes[processID]
	return process, exists && args.Bool(1)
}

// ListProcesses mocks process listing
func (m *ProcessManagerMock) ListProcesses() []*ai.Process {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called()
	
	processes := make([]*ai.Process, 0, len(m.processes))
	for _, process := range m.processes {
		processes = append(processes, process)
	}
	
	if args.Error(0) == nil {
		return processes
	}
	return []*ai.Process{}
}

// TerminateProcess mocks process termination
func (m *ProcessManagerMock) TerminateProcess(process *ai.Process) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(process)
	
	if p, exists := m.processes[process.ID]; exists {
		p.Status = ai.ProcessStatusStopped
		p.EndTime = time.Now()
	}

	return args.Error(0)
}

// KillProcess mocks process killing
func (m *ProcessManagerMock) KillProcess(process *ai.Process) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(process)
	
	if p, exists := m.processes[process.ID]; exists {
		p.Status = ai.ProcessStatusError
		p.EndTime = time.Now()
	}

	return args.Error(0)
}

// SetupMockExpectations sets up default mock expectations for happy path scenarios
func SetupHappyPathMockExpectations(services *TestServices) {
	// Git Manager expectations
	services.GitManager.On("CloneRepository", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	services.GitManager.On("CreateBranch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	services.GitManager.On("CheckoutBranch", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Worktree Service expectations
	services.WorktreeService.On("CreateWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mock.AnythingOfType("*entity.Worktree"), nil)
	services.WorktreeService.On("DeleteWorktree", mock.Anything, mock.Anything).Return(nil)

	// GitHub Service expectations
	services.GitHubService.On("CreatePullRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mock.AnythingOfType("*entity.PullRequest"), nil)

	// AI Service expectations
	services.AIPlanning.On("StartPlanning", mock.Anything, mock.Anything).Return(mock.AnythingOfType("*entity.Execution"), nil)
	services.AIExecution.On("StartImplementation", mock.Anything, mock.Anything).Return(mock.AnythingOfType("*entity.Execution"), nil)

	// Process Manager expectations
	services.ProcessManager.On("SpawnProcess", mock.Anything, mock.Anything, mock.Anything).Return(mock.AnythingOfType("*ai.Process"), nil)
	services.ProcessManager.On("GetProcess", mock.Anything).Return(mock.AnythingOfType("*ai.Process"), true)
	services.ProcessManager.On("ListProcesses").Return(mock.AnythingOfType("[]*ai.Process"))
	services.ProcessManager.On("TerminateProcess", mock.Anything).Return(nil)
	services.ProcessManager.On("KillProcess", mock.Anything).Return(nil)
}