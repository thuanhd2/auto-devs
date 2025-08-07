package github

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitHubService is a mock implementation of GitHubServiceInterface for testing
type MockGitHubService struct {
	mock.Mock
}

func (m *MockGitHubService) CreatePullRequest(ctx context.Context, repo, base, head, title, body string) (*entity.PullRequest, error) {
	args := m.Called(ctx, repo, base, head, title, body)
	if pr := args.Get(0); pr != nil {
		return pr.(*entity.PullRequest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitHubService) UpdatePullRequest(ctx context.Context, repo string, prNumber int, updates map[string]interface{}) error {
	args := m.Called(ctx, repo, prNumber, updates)
	return args.Error(0)
}

func (m *MockGitHubService) GetPullRequest(ctx context.Context, repo string, prNumber int) (*entity.PullRequest, error) {
	args := m.Called(ctx, repo, prNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PullRequest), args.Error(1)
}

func TestNewPRCreator(t *testing.T) {
	mockGitHub := &MockGitHubService{}
	baseURL := "https://auto-devs.example.com"

	creator := NewPRCreator(mockGitHub, baseURL)

	assert.NotNil(t, creator)
	assert.Equal(t, mockGitHub, creator.githubService)
	assert.Equal(t, baseURL, creator.baseURL)
}

func TestPRCreator_GeneratePRTitle(t *testing.T) {
	creator := NewPRCreator(nil, "")

	tests := []struct {
		name     string
		task     entity.Task
		expected string
		hasError bool
	}{
		{
			name: "Feature task",
			task: entity.Task{
				ID:    uuid.New(),
				Title: "Add user authentication feature",
			},
			expected: "[feat] Add user authentication feature",
			hasError: false,
		},
		{
			name: "Bug fix task",
			task: entity.Task{
				ID:    uuid.New(),
				Title: "Fix login bug in authentication",
			},
			expected: "[fix] Fix login bug in authentication",
			hasError: false,
		},
		{
			name: "Empty title",
			task: entity.Task{
				ID:    uuid.New(),
				Title: "",
			},
			hasError: true,
		},
		{
			name: "Long title gets truncated",
			task: entity.Task{
				ID:    uuid.New(),
				Title: "This is a very long title that should be truncated when it exceeds the maximum length allowed for GitHub pull request titles to prevent API errors and maintain readability for developers",
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := creator.GeneratePRTitle(tt.task)

			if tt.hasError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Contains(t, title, tt.task.ID.String()[:8])

			if tt.expected != "" {
				assert.Contains(t, title, tt.expected[:6]) // Check prefix
			}

			// Ensure title isn't too long
			assert.LessOrEqual(t, len(title), 255)
		})
	}
}

func TestPRCreator_GeneratePRDescription(t *testing.T) {
	creator := NewPRCreator(nil, "https://auto-devs.example.com")

	taskID := uuid.New()
	projectID := uuid.New()
	planID := uuid.New()
	executionID := uuid.New()

	task := entity.Task{
		ID:          taskID,
		ProjectID:   projectID,
		Title:       "Test task",
		Description: "Test task description",
		Priority:    entity.TaskPriorityHigh,
		Status:      entity.TaskStatusIMPLEMENTING,
	}

	plan := &entity.Plan{
		ID:      planID,
		TaskID:  taskID,
		Status:  entity.PlanStatusAPPROVED,
		Content: "Test plan content",
	}

	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	execution := entity.Execution{
		ID:          executionID,
		TaskID:      taskID,
		Status:      entity.ExecutionStatusCompleted,
		StartedAt:   startTime,
		CompletedAt: &endTime,
		Result:      `{"status": "success", "files": ["test.go"]}`,
	}

	description, err := creator.GeneratePRDescription(task, plan, execution)

	assert.NoError(t, err)
	assert.NotEmpty(t, description)

	// Check that description contains expected sections
	assert.Contains(t, description, "## Task Information")
	assert.Contains(t, description, "## Implementation Plan")
	assert.Contains(t, description, "## Implementation Summary")
	assert.Contains(t, description, "## Testing Instructions")
	assert.Contains(t, description, "## Review Checklist")

	// Check that key information is included
	assert.Contains(t, description, task.Title)
	assert.Contains(t, description, task.Description)
	assert.Contains(t, description, taskID.String())
	assert.Contains(t, description, planID.String())
	assert.Contains(t, description, executionID.String())
}

func TestPRCreator_ValidateTaskForPRCreation(t *testing.T) {
	creator := NewPRCreator(nil, "")

	validTask := entity.Task{
		ID:         uuid.New(),
		Title:      "Valid task",
		BranchName: stringPtr("feature/valid-task"),
		Project: entity.Project{
			RepositoryURL: "https://github.com/owner/repo",
		},
	}

	validExecution := entity.Execution{
		Status: entity.ExecutionStatusCompleted,
	}

	tests := []struct {
		name      string
		task      entity.Task
		execution entity.Execution
		hasError  bool
	}{
		{
			name:      "Valid task and execution",
			task:      validTask,
			execution: validExecution,
			hasError:  false,
		},
		{
			name: "Task with empty title",
			task: entity.Task{
				ID:         uuid.New(),
				Title:      "",
				BranchName: stringPtr("feature/test"),
				Project: entity.Project{
					RepositoryURL: "https://github.com/owner/repo",
				},
			},
			execution: validExecution,
			hasError:  true,
		},
		{
			name: "Task without branch name",
			task: entity.Task{
				ID:    uuid.New(),
				Title: "Valid task",
				Project: entity.Project{
					RepositoryURL: "https://github.com/owner/repo",
				},
			},
			execution: validExecution,
			hasError:  true,
		},
		{
			name: "Execution not completed",
			task: validTask,
			execution: entity.Execution{
				Status: entity.ExecutionStatusRunning,
			},
			hasError: true,
		},
		{
			name: "Task without repository",
			task: entity.Task{
				ID:         uuid.New(),
				Title:      "Valid task",
				BranchName: stringPtr("feature/test"),
				Project:    entity.Project{},
			},
			execution: validExecution,
			hasError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := creator.ValidateTaskForPRCreation(tt.task, tt.execution)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPRCreator_getRepositoryFromTask(t *testing.T) {
	creator := NewPRCreator(nil, "")

	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "HTTPS URL",
			repoURL:  "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS URL with .git",
			repoURL:  "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH URL",
			repoURL:  "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTP URL",
			repoURL:  "http://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "Invalid URL",
			repoURL:  "invalid-url",
			expected: "",
		},
		{
			name:     "Empty URL",
			repoURL:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := entity.Task{
				Project: entity.Project{
					RepositoryURL: tt.repoURL,
				},
			}

			result := creator.getRepositoryFromTask(task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPRCreator_SanitizeForGitHub(t *testing.T) {
	creator := NewPRCreator(nil, "")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "This is normal text",
			expected: "This is normal text",
		},
		{
			name:     "Text with leading/trailing spaces",
			input:    "  spaced text  ",
			expected: "spaced text",
		},
		{
			name:     "Text with null bytes",
			input:    "text\x00with\x00nulls",
			expected: "textwithnulls", // null bytes removed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := creator.SanitizeForGitHub(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPRCreator_CreatePRFromImplementation(t *testing.T) {
	mockGitHub := &MockGitHubService{}
	creator := NewPRCreator(mockGitHub, "https://auto-devs.example.com")

	taskID := uuid.New()
	branchName := "feature/test-task"

	task := entity.Task{
		ID:         taskID,
		Title:      "Test task",
		BranchName: &branchName,
		Project: entity.Project{
			RepositoryURL: "https://github.com/owner/repo",
		},
	}

	execution := entity.Execution{
		ID:        uuid.New(),
		TaskID:    taskID,
		Status:    entity.ExecutionStatusCompleted,
		StartedAt: time.Now().Add(-1 * time.Hour),
	}

	expectedPR := &entity.PullRequest{
		ID:             uuid.New(),
		TaskID:         taskID,
		GitHubPRNumber: 123,
		Repository:     "owner/repo",
		Title:          "[feat] Test task",
		Status:         entity.PullRequestStatusOpen,
		HeadBranch:     branchName,
		BaseBranch:     "main",
	}

	// Set up mock expectations
	mockGitHub.On("CreatePullRequest",
		mock.Anything,                 // context
		"owner/repo",                  // repository
		"main",                        // base
		branchName,                    // head
		mock.AnythingOfType("string"), // title
		mock.AnythingOfType("string"), // body
	).Return(expectedPR, nil)

	mockGitHub.On("UpdatePullRequest",
		mock.Anything, // context
		"owner/repo",  // repository
		123,           // PR number
		mock.AnythingOfType("map[string]interface {}"), // updates
	).Return(nil)

	// Execute test
	ctx := context.Background()
	result, err := creator.CreatePRFromImplementation(ctx, task, execution, nil)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPR.ID, result.ID)
	assert.Equal(t, expectedPR.TaskID, result.TaskID)

	// Verify mock calls
	mockGitHub.AssertExpectations(t)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
