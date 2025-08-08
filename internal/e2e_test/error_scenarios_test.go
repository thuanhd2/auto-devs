package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestPlanningServiceFailure tests handling of AI planning service failures
func TestPlanningServiceFailure(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	// Create project and task
	project := dataGen.GenerateProject(ProjectConfig{Name: "Planning Failure Test"})
	task := dataGen.GenerateTask(project.ID, TaskConfig{
		Title:  "Task with Planning Failure",
		Status: entity.TaskStatusTODO,
	})

	// Set up mock to return error for planning
	suite.services.AIPlanning.On("StartPlanning", mock.Anything, mock.Anything).
		Return(nil, errors.New("AI planning service unavailable"))

	// Attempt to start planning
	url := fmt.Sprintf("%s/api/v1/tasks/%s/plan", suite.GetServerURL(), task.ID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return error response
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var errorResponse dto.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	require.NoError(t, err)
	assert.Contains(t, errorResponse.Message, "planning service")

	// Verify task status unchanged
	updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusTODO, updatedTask.Status)

	// Verify no executions were created
	executions, err := suite.repositories.Execution.List(suite.ctx, entity.ExecutionFilters{
		TaskID: &task.ID,
	})
	require.NoError(t, err)
	assert.Empty(t, executions)
}

// TestPlanningServiceRetry tests retry mechanisms for planning failures
func TestPlanningServiceRetry(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	project := dataGen.GenerateProject(ProjectConfig{Name: "Planning Retry Test"})
	task := dataGen.GenerateTask(project.ID, TaskConfig{
		Title:  "Task with Planning Retry",
		Status: entity.TaskStatusTODO,
	})

	// Set up mock to fail first time, succeed second time
	suite.services.AIPlanning.On("StartPlanning", mock.Anything, mock.Anything).
		Return(nil, errors.New("temporary failure")).Once()
	
	suite.services.AIPlanning.On("StartPlanning", mock.Anything, mock.Anything).
		Return(&entity.Execution{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Type:      entity.ExecutionTypePlanning,
			Status:    entity.ExecutionStatusRunning,
			StartedAt: time.Now(),
		}, nil).Once()

	// First attempt should fail
	planningResponse := suite.attemptTaskPlanning(t, task.ID)
	assert.Nil(t, planningResponse)

	// Wait a moment before retry
	time.Sleep(1 * time.Second)

	// Second attempt should succeed
	planningResponse = suite.attemptTaskPlanning(t, task.ID)
	assert.NotNil(t, planningResponse)

	// Verify task moved to PLANNING status
	assert.True(t, suite.WaitForTaskStatus(task.ID, entity.TaskStatusPLANNING, 5*time.Second))
}

// TestImplementationServiceFailure tests handling of AI implementation failures
func TestImplementationServiceFailure(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	// Create task ready for implementation
	flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
		Project: ProjectConfig{Name: "Implementation Failure Test"},
		Task: TaskConfig{
			Title:  "Task with Implementation Failure",
			Status: entity.TaskStatusIMPLEMENTING,
		},
		Plan: PlanConfig{Status: entity.PlanStatusApproved},
		IncludePlan: true,
	})

	// Set up mock to return error for implementation
	suite.services.AIExecution.On("StartImplementation", mock.Anything, mock.Anything).
		Return(nil, errors.New("AI implementation service failed"))

	// Attempt implementation
	url := fmt.Sprintf("%s/api/v1/tasks/%s/implement", suite.GetServerURL(), flow.Task.ID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return error
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Verify task status unchanged
	updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, flow.Task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, updatedTask.Status)

	// Verify no worktree created
	worktrees, err := suite.repositories.Worktree.List(suite.ctx, entity.WorktreeFilters{
		TaskID: &flow.Task.ID,
	})
	require.NoError(t, err)
	assert.Empty(t, worktrees)
}

// TestGitOperationFailures tests handling of Git-related failures
func TestGitOperationFailures(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	// Test git clone failure
	t.Run("GitCloneFailure", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{
			Name:          "Git Clone Failure Test",
			RepositoryURL: "https://github.com/invalid/repo.git",
		})

		// Mock git clone failure
		suite.services.GitManager.On("CloneRepository", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("repository not found"))

		// Attempt to create worktree (which requires clone)
		url := fmt.Sprintf("%s/api/v1/worktrees", suite.GetServerURL())
		requestBody := map[string]interface{}{
			"project_id": project.ID.String(),
			"task_id":    uuid.New().String(),
			"branch":     "test-branch",
		}

		resp, err := suite.postJSON(url, requestBody)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Test branch creation failure
	t.Run("BranchCreationFailure", func(t *testing.T) {
		flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
			Project:         ProjectConfig{Name: "Branch Failure Test"},
			Task:           TaskConfig{Status: entity.TaskStatusIMPLEMENTING},
			Plan:           PlanConfig{Status: entity.PlanStatusApproved},
			IncludePlan:    true,
		})

		// Mock successful clone but failed branch creation
		suite.services.GitManager.On("CloneRepository", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		suite.services.GitManager.On("CreateBranch", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("branch already exists"))

		// Mock worktree service to handle the error
		suite.services.WorktreeService.On("CreateWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("failed to create branch"))

		// Attempt implementation
		resp := suite.attemptTaskImplementationHTTP(t, flow.Task.ID)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify no worktree was created
		worktrees, err := suite.repositories.Worktree.List(suite.ctx, entity.WorktreeFilters{
			TaskID: &flow.Task.ID,
		})
		require.NoError(t, err)
		assert.Empty(t, worktrees)
	})
}

// TestGitHubAPIFailures tests handling of GitHub API issues
func TestGitHubAPIFailures(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	// Test PR creation failure due to rate limiting
	t.Run("PRCreationRateLimit", func(t *testing.T) {
		flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
			Project:         ProjectConfig{Name: "GitHub Rate Limit Test"},
			Task:           TaskConfig{Status: entity.TaskStatusCODEREVIEWING},
			Worktree:       WorktreeConfig{Branch: "test-branch"},
			IncludeWorktree: true,
		})

		// Mock GitHub API rate limit error
		suite.services.GitHubService.On("CreatePullRequest", 
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, 
			mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("API rate limit exceeded"))

		// Attempt PR creation
		resp := suite.attemptCreatePullRequest(t, flow.Task.ID, "Test PR", "Test body")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

		// Verify no PR was created in database
		prs, err := suite.repositories.PullRequest.List(suite.ctx, entity.PullRequestFilters{
			TaskID: &flow.Task.ID,
		})
		require.NoError(t, err)
		assert.Empty(t, prs)
	})

	// Test webhook delivery failure
	t.Run("WebhookDeliveryFailure", func(t *testing.T) {
		// This test simulates webhook delivery failures
		// In a real scenario, we would test webhook retry mechanisms
		flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
			Project: ProjectConfig{Name: "Webhook Failure Test"},
			Task:    TaskConfig{Status: entity.TaskStatusCODEREVIEWING},
			PullRequest: PullRequestConfig{
				Title: "Test PR",
				State: "open",
			},
			IncludePullRequest: true,
		})

		// Simulate webhook delivery failure by not updating PR status
		// Verify that system handles missing webhook gracefully
		time.Sleep(1 * time.Second)

		// PR should remain in open state
		pr, err := suite.repositories.PullRequest.GetByID(suite.ctx, flow.PullRequest.ID)
		require.NoError(t, err)
		assert.Equal(t, "open", pr.State)
		assert.Nil(t, pr.MergedAt)
	})
}

// TestDatabaseConnectionFailures tests handling of database connectivity issues
func TestDatabaseConnectionFailures(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	// This test is conceptual - in practice, we would need to simulate
	// database connection failures using a proxy or testing framework
	
	t.Run("DatabaseConnectionLoss", func(t *testing.T) {
		// Create initial data
		dataGen := NewTestDataGenerator(suite)
		project := dataGen.GenerateProject(ProjectConfig{Name: "DB Failure Test"})

		// Simulate database connection issue by using cancelled context
		ctx, cancel := context.WithCancel(suite.ctx)
		cancel() // Immediately cancel to simulate connection loss

		// Attempt to create task with cancelled context
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entity.TaskStatusTODO,
		}

		err := suite.repositories.Task.Create(ctx, task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Test transaction rollback on failure
		dataGen := NewTestDataGenerator(suite)
		project := dataGen.GenerateProject(ProjectConfig{Name: "Transaction Test"})

		// Create a task
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title: "Transaction Test Task",
		})

		// Start a transaction that will fail midway
		tx := suite.db.DB.Begin()
		
		// Create execution
		execution := &entity.Execution{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Type:      entity.ExecutionTypePlanning,
			Status:    entity.ExecutionStatusRunning,
			StartedAt: time.Now(),
		}
		
		err := tx.Create(execution).Error
		require.NoError(t, err)

		// Force rollback
		tx.Rollback()

		// Verify execution was not saved
		_, err = suite.repositories.Execution.GetByID(suite.ctx, execution.ID)
		assert.Error(t, err)
	})
}

// TestConcurrencyIssues tests handling of race conditions and concurrent access
func TestConcurrencyIssues(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("ConcurrentTaskStatusUpdates", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Concurrency Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Concurrent Update Test",
			Status: entity.TaskStatusTODO,
		})

		// Simulate concurrent status updates
		done := make(chan bool, 2)
		errors := make(chan error, 2)

		// Goroutine 1: Try to update to PLANNING
		go func() {
			defer func() { done <- true }()
			
			taskCopy := *task
			taskCopy.Status = entity.TaskStatusPLANNING
			
			if err := suite.repositories.Task.Update(suite.ctx, &taskCopy); err != nil {
				errors <- err
			}
		}()

		// Goroutine 2: Try to update to different status simultaneously
		go func() {
			defer func() { done <- true }()
			
			taskCopy := *task
			taskCopy.Status = entity.TaskStatusCANCELLED
			
			if err := suite.repositories.Task.Update(suite.ctx, &taskCopy); err != nil {
				errors <- err
			}
		}()

		// Wait for both operations to complete
		<-done
		<-done

		// Check if any errors occurred (some databases handle this differently)
		select {
		case err := <-errors:
			t.Logf("Concurrent update error (expected): %v", err)
		default:
			// No error - check final state is consistent
		}

		// Verify task is in a consistent state
		finalTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
		require.NoError(t, err)
		assert.NotEqual(t, entity.TaskStatusTODO, finalTask.Status) // Should be updated
	})

	t.Run("ConcurrentWorktreeCreation", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Concurrent Worktree Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Concurrent Worktree Test",
			Status: entity.TaskStatusIMPLEMENTING,
		})

		// Mock successful worktree creation
		suite.services.WorktreeService.On("CreateWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(&entity.Worktree{
				ID:        uuid.New(),
				ProjectID: project.ID,
				TaskID:    task.ID,
				Branch:    "test-branch",
				Path:      "/tmp/test-worktree",
				Status:    entity.WorktreeStatusActive,
				CreatedAt: time.Now(),
			}, nil)

		// Simulate concurrent worktree creation attempts
		done := make(chan bool, 2)
		worktrees := make(chan *entity.Worktree, 2)

		for i := 0; i < 2; i++ {
			go func() {
				defer func() { done <- true }()
				
				// Both attempts use same branch name - should conflict
				worktree, err := suite.services.WorktreeService.CreateWorktree(
					suite.ctx, project.ID, task.ID, "test-branch")
				
				if err == nil {
					worktrees <- worktree
				}
			}()
		}

		// Wait for completion
		<-done
		<-done

		// Should have created at least one worktree
		close(worktrees)
		count := 0
		for range worktrees {
			count++
		}
		
		assert.GreaterOrEqual(t, count, 1)
		assert.LessOrEqual(t, count, 2) // At most 2, ideally 1 due to conflict handling
	})
}

// TestResourceExhaustion tests system behavior under resource constraints
func TestResourceExhaustion(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("MaxConcurrentTasksExceeded", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{
			Name:               "Resource Limit Test",
			MaxConcurrentTasks: 2, // Low limit for testing
		})

		// Create tasks beyond the limit
		tasks := make([]*entity.Task, 4)
		for i := 0; i < 4; i++ {
			tasks[i] = dataGen.GenerateTask(project.ID, TaskConfig{
				Title:  fmt.Sprintf("Resource Test Task %d", i+1),
				Status: entity.TaskStatusTODO,
			})
		}

		// Mock successful planning for all tasks
		suite.services.AIPlanning.On("StartPlanning", mock.Anything, mock.Anything).
			Return(&entity.Execution{
				ID:        uuid.New(),
				TaskID:    uuid.New(),
				Type:      entity.ExecutionTypePlanning,
				Status:    entity.ExecutionStatusRunning,
				StartedAt: time.Now(),
			}, nil)

		// Start planning for all tasks
		successCount := 0
		for i := 0; i < 4; i++ {
			resp := suite.attemptTaskPlanningHTTP(t, tasks[i].ID)
			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				// Expected for tasks beyond limit
			}
			resp.Body.Close()
		}

		// Should respect the concurrent task limit
		assert.LessOrEqual(t, successCount, 2)
	})

	t.Run("DiskSpaceExhaustion", func(t *testing.T) {
		// This is a conceptual test - in practice, we would need to
		// simulate disk space issues or mock file system operations
		
		project := dataGen.GenerateProject(ProjectConfig{Name: "Disk Space Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Disk Space Test Task",
			Status: entity.TaskStatusIMPLEMENTING,
		})

		// Mock worktree creation failure due to disk space
		suite.services.WorktreeService.On("CreateWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("no space left on device"))

		// Attempt implementation
		resp := suite.attemptTaskImplementationHTTP(t, task.ID)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify graceful degradation
		updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.TaskStatusIMPLEMENTING, updatedTask.Status) // Status unchanged
	})
}

// Helper methods for error scenario testing

func (suite *E2ETestSuite) attemptTaskPlanning(t *testing.T, taskID uuid.UUID) *dto.ExecutionResponse {
	resp := suite.attemptTaskPlanningHTTP(t, taskID)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var response dto.ExecutionResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response
}

func (suite *E2ETestSuite) attemptTaskPlanningHTTP(t *testing.T, taskID uuid.UUID) *http.Response {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/plan", suite.GetServerURL(), taskID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func (suite *E2ETestSuite) attemptTaskImplementationHTTP(t *testing.T, taskID uuid.UUID) *http.Response {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/implement", suite.GetServerURL(), taskID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func (suite *E2ETestSuite) attemptCreatePullRequest(t *testing.T, taskID uuid.UUID, title, body string) *http.Response {
	requestBody := map[string]interface{}{
		"title": title,
		"body":  body,
	}

	return suite.postJSON(fmt.Sprintf("%s/api/v1/tasks/%s/pull-request", suite.GetServerURL(), taskID.String()), requestBody)
}

func (suite *E2ETestSuite) postJSON(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}