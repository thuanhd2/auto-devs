package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteTaskAutomationFlow tests the complete happy path from task creation to PR merge
func TestCompleteTaskAutomationFlow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	// Set up happy path mock expectations
	SetupHappyPathMockExpectations(suite.services)

	// Create test data generator
	dataGen := NewTestDataGenerator(suite)

	// Step 1: Create project
	projectConfig := ProjectConfig{
		Name:          "E2E Test Project",
		Description:   "Project for end-to-end testing",
		RepositoryURL: "https://github.com/test/e2e-test.git",
		DefaultBranch: "main",
		AutoMerge:     true,
	}
	project := dataGen.GenerateProject(projectConfig)
	require.NotNil(t, project)
	require.NotEqual(t, uuid.Nil, project.ID)

	// Step 2: Create task in TODO status
	taskConfig := TaskConfig{
		Title:       "Implement user authentication",
		Description: "Add JWT-based user authentication to the API",
		Status:      entity.TaskStatusTODO,
		Priority:    entity.TaskPriorityHigh,
	}
	task := dataGen.GenerateTask(project.ID, taskConfig)
	require.NotNil(t, task)
	require.Equal(t, entity.TaskStatusTODO, task.Status)

	// Step 3: Start planning phase (TODO → PLANNING)
	planningResponse := suite.startTaskPlanning(t, task.ID)
	require.NotNil(t, planningResponse)

	// Wait for planning to complete
	assert.True(t, suite.WaitForTaskStatus(task.ID, entity.TaskStatusPLANREVIEWING, 10*time.Second))

	// Verify plan was created
	plans, err := suite.repositories.Plan.List(suite.ctx, entity.PlanFilters{TaskID: &task.ID})
	require.NoError(t, err)
	require.Len(t, plans, 1)
	plan := plans[0]
	assert.Equal(t, entity.PlanStatusDraft, plan.Status)
	assert.NotEmpty(t, plan.Content)

	// Step 4: Approve plan (PLAN_REVIEWING → IMPLEMENTING)
	suite.approvePlan(t, plan.ID)

	// Verify task status updated
	updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, updatedTask.Status)

	// Verify plan status updated
	updatedPlan, err := suite.repositories.Plan.GetByID(suite.ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PlanStatusApproved, updatedPlan.Status)

	// Step 5: Execute implementation
	implementationResponse := suite.startTaskImplementation(t, task.ID)
	require.NotNil(t, implementationResponse)

	// Wait for implementation to complete
	assert.True(t, suite.WaitForTaskStatus(task.ID, entity.TaskStatusCODEREVIEWING, 15*time.Second))

	// Verify worktree was created
	worktrees, err := suite.repositories.Worktree.List(suite.ctx, entity.WorktreeFilters{TaskID: &task.ID})
	require.NoError(t, err)
	require.Len(t, worktrees, 1)
	worktree := worktrees[0]
	assert.Equal(t, entity.WorktreeStatusActive, worktree.Status)
	assert.Contains(t, worktree.Branch, "task")

	// Step 6: Create pull request (automatically triggered)
	// Simulate PR creation after implementation
	prConfig := PullRequestConfig{
		Title:      fmt.Sprintf("Implement %s", task.Title),
		Body:       "This PR implements the planned changes for the task",
		HeadBranch: worktree.Branch,
		BaseBranch: "main",
	}
	pr := dataGen.GeneratePullRequest(task.ID, prConfig)
	require.NotNil(t, pr)

	// Step 7: Simulate PR merge and completion
	suite.services.GitHubService.SimulatePRMerge(pr.Number)

	// Step 8: Complete task (CODE_REVIEWING → DONE)
	suite.completeTask(t, task.ID)

	// Wait for task completion
	assert.True(t, suite.WaitForTaskStatus(task.ID, entity.TaskStatusDONE, 5*time.Second))

	// Verify final state
	finalTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusDONE, finalTask.Status)
	assert.Equal(t, entity.TaskGitStatusCompleted, finalTask.GitStatus)

	// Verify audit trail
	auditLogs, err := suite.repositories.Audit.List(suite.ctx, entity.AuditFilters{
		ResourceType: "task",
		ResourceID:   &task.ID,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(auditLogs), 4) // At least status changes logged
}

// TestMultiTaskProjectWorkflow tests handling multiple tasks in a single project
func TestMultiTaskProjectWorkflow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	SetupHappyPathMockExpectations(suite.services)
	dataGen := NewTestDataGenerator(suite)

	// Create project
	project := dataGen.GenerateProject(ProjectConfig{
		Name:               "Multi-Task Project",
		MaxConcurrentTasks: 3,
	})

	// Create multiple tasks
	tasks := make([]*entity.Task, 5)
	for i := 0; i < 5; i++ {
		taskConfig := TaskConfig{
			Title:       fmt.Sprintf("Feature %d Implementation", i+1),
			Description: fmt.Sprintf("Implement feature %d", i+1),
			Status:      entity.TaskStatusTODO,
			Priority:    entity.TaskPriorityMedium,
		}
		tasks[i] = dataGen.GenerateTask(project.ID, taskConfig)
	}

	// Start planning for first 3 tasks (respecting concurrent limit)
	for i := 0; i < 3; i++ {
		response := suite.startTaskPlanning(t, tasks[i].ID)
		require.NotNil(t, response)
	}

	// Wait for planning to complete
	for i := 0; i < 3; i++ {
		assert.True(t, suite.WaitForTaskStatus(tasks[i].ID, entity.TaskStatusPLANREVIEWING, 10*time.Second))
	}

	// Approve plans and start implementation
	for i := 0; i < 3; i++ {
		plans, err := suite.repositories.Plan.List(suite.ctx, entity.PlanFilters{TaskID: &tasks[i].ID})
		require.NoError(t, err)
		require.Len(t, plans, 1)

		suite.approvePlan(t, plans[0].ID)
		assert.True(t, suite.WaitForTaskStatus(tasks[i].ID, entity.TaskStatusIMPLEMENTING, 5*time.Second))

		suite.startTaskImplementation(t, tasks[i].ID)
	}

	// Wait for implementations to complete
	for i := 0; i < 3; i++ {
		assert.True(t, suite.WaitForTaskStatus(tasks[i].ID, entity.TaskStatusCODEREVIEWING, 15*time.Second))
	}

	// Complete first task, which should allow starting the 4th task
	suite.completeTask(t, tasks[0].ID)
	assert.True(t, suite.WaitForTaskStatus(tasks[0].ID, entity.TaskStatusDONE, 5*time.Second))

	// Start planning for 4th task
	response := suite.startTaskPlanning(t, tasks[3].ID)
	require.NotNil(t, response)

	// Verify task sequencing and concurrency limits are respected
	activeTasks := 0
	for _, task := range tasks {
		updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
		require.NoError(t, err)

		if updatedTask.Status == entity.TaskStatusIMPLEMENTING || 
		   updatedTask.Status == entity.TaskStatusCODEREVIEWING {
			activeTasks++
		}
	}
	assert.LessOrEqual(t, activeTasks, 3) // Should not exceed concurrent limit
}

// TestPlanGenerationAndApproval tests the planning workflow
func TestPlanGenerationAndApproval(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	SetupHappyPathMockExpectations(suite.services)
	dataGen := NewTestDataGenerator(suite)

	// Create task requiring detailed planning
	flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
		Project: ProjectConfig{
			Name: "Planning Test Project",
		},
		Task: TaskConfig{
			Title:       "Complex Feature Implementation",
			Description: "Implement a complex feature requiring detailed planning",
			Status:      entity.TaskStatusTODO,
		},
	})

	// Start planning
	planningResponse := suite.startTaskPlanning(t, flow.Task.ID)
	require.NotNil(t, planningResponse)

	// Verify planning execution was created
	execution, err := suite.repositories.Execution.GetByID(suite.ctx, planningResponse.ExecutionID)
	require.NoError(t, err)
	assert.Equal(t, entity.ExecutionTypePlanning, execution.Type)
	assert.Equal(t, entity.ExecutionStatusRunning, execution.Status)

	// Wait for planning completion
	assert.True(t, suite.WaitForExecutionStatus(execution.ID, entity.ExecutionStatusCompleted, 10*time.Second))
	assert.True(t, suite.WaitForTaskStatus(flow.Task.ID, entity.TaskStatusPLANREVIEWING, 10*time.Second))

	// Verify plan was generated
	plans, err := suite.repositories.Plan.List(suite.ctx, entity.PlanFilters{TaskID: &flow.Task.ID})
	require.NoError(t, err)
	require.Len(t, plans, 1)

	plan := plans[0]
	assert.Equal(t, entity.PlanStatusDraft, plan.Status)
	assert.NotEmpty(t, plan.Content)
	assert.Contains(t, plan.Content, "steps") // Should contain planning steps

	// Test plan approval
	suite.approvePlan(t, plan.ID)

	// Verify plan and task status updates
	updatedPlan, err := suite.repositories.Plan.GetByID(suite.ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PlanStatusApproved, updatedPlan.Status)

	updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, flow.Task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskStatusIMPLEMENTING, updatedTask.Status)
}

// TestImplementationExecution tests the implementation workflow
func TestImplementationExecution(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	SetupHappyPathMockExpectations(suite.services)
	dataGen := NewTestDataGenerator(suite)

	// Create approved task ready for implementation
	flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
		Project: ProjectConfig{
			Name: "Implementation Test Project",
		},
		Task: TaskConfig{
			Title:  "API Endpoint Implementation",
			Status: entity.TaskStatusIMPLEMENTING,
		},
		Plan: PlanConfig{
			Status: entity.PlanStatusApproved,
		},
		IncludePlan: true,
	})

	// Start implementation
	implementationResponse := suite.startTaskImplementation(t, flow.Task.ID)
	require.NotNil(t, implementationResponse)

	// Verify implementation execution was created
	execution, err := suite.repositories.Execution.GetByID(suite.ctx, implementationResponse.ExecutionID)
	require.NoError(t, err)
	assert.Equal(t, entity.ExecutionTypeImplementation, execution.Type)
	assert.Equal(t, entity.ExecutionStatusRunning, execution.Status)

	// Verify worktree was created
	worktrees, err := suite.repositories.Worktree.List(suite.ctx, entity.WorktreeFilters{TaskID: &flow.Task.ID})
	require.NoError(t, err)
	require.Len(t, worktrees, 1)

	worktree := worktrees[0]
	assert.Equal(t, entity.WorktreeStatusActive, worktree.Status)
	assert.NotEmpty(t, worktree.Branch)
	assert.NotEmpty(t, worktree.Path)

	// Wait for implementation completion
	assert.True(t, suite.WaitForExecutionStatus(execution.ID, entity.ExecutionStatusCompleted, 15*time.Second))
	assert.True(t, suite.WaitForTaskStatus(flow.Task.ID, entity.TaskStatusCODEREVIEWING, 15*time.Second))

	// Verify task Git status was updated
	updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, flow.Task.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.TaskGitStatusCompleted, updatedTask.GitStatus)

	// Verify execution logs were created
	logs, err := suite.repositories.ExecutionLog.List(suite.ctx, entity.ExecutionLogFilters{
		ExecutionID: &execution.ID,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
}

// TestPRCreationAndMonitoring tests pull request workflow
func TestPRCreationAndMonitoring(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	SetupHappyPathMockExpectations(suite.services)
	dataGen := NewTestDataGenerator(suite)

	// Create completed implementation ready for PR
	flow := dataGen.GenerateCompleteTaskFlow(CompleteFlowConfig{
		Project: ProjectConfig{
			Name: "PR Test Project",
		},
		Task: TaskConfig{
			Title:     "Feature Ready for PR",
			Status:    entity.TaskStatusCODEREVIEWING,
			GitStatus: entity.TaskGitStatusCompleted,
		},
		Worktree: WorktreeConfig{
			Branch: "task-feature-implementation",
			Status: entity.WorktreeStatusActive,
		},
		IncludeWorktree: true,
	})

	// Create pull request
	prResponse := suite.createPullRequest(t, flow.Task.ID, "Implement feature", "This PR implements the planned feature")
	require.NotNil(t, prResponse)

	// Verify PR was created in database
	prs, err := suite.repositories.PullRequest.List(suite.ctx, entity.PullRequestFilters{
		TaskID: &flow.Task.ID,
	})
	require.NoError(t, err)
	require.Len(t, prs, 1)

	pr := prs[0]
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, flow.Worktree.Branch, pr.HeadBranch)
	assert.NotEmpty(t, pr.HTMLURL)

	// Test PR monitoring - simulate external merge
	suite.services.GitHubService.SimulatePRMerge(pr.Number)

	// Simulate webhook processing or periodic sync
	time.Sleep(1 * time.Second)

	// Verify PR merge was detected
	updatedPR, err := suite.repositories.PullRequest.GetByID(suite.ctx, pr.ID)
	require.NoError(t, err)
	assert.Equal(t, "closed", updatedPR.State)
	assert.NotNil(t, updatedPR.MergedAt)

	// Complete task after PR merge
	suite.completeTask(t, flow.Task.ID)
	assert.True(t, suite.WaitForTaskStatus(flow.Task.ID, entity.TaskStatusDONE, 5*time.Second))
}

// TestWebSocketRealTimeUpdates tests real-time updates via WebSocket
func TestWebSocketRealTimeUpdates(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	SetupHappyPathMockExpectations(suite.services)
	dataGen := NewTestDataGenerator(suite)

	// Create WebSocket connection
	wsURL := suite.GetWebSocketURL()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Set up message collection
	messages := make(chan map[string]interface{}, 10)
	go func() {
		defer close(messages)
		for {
			var message map[string]interface{}
			if err := conn.ReadJSON(&message); err != nil {
				return
			}
			select {
			case messages <- message:
			default:
				// Buffer full, skip message
			}
		}
	}()

	// Create project and task
	project := dataGen.GenerateProject(ProjectConfig{Name: "WebSocket Test Project"})
	task := dataGen.GenerateTask(project.ID, TaskConfig{
		Title:  "WebSocket Test Task",
		Status: entity.TaskStatusTODO,
	})

	// Subscribe to task updates
	subscribeMessage := map[string]interface{}{
		"type": "subscribe",
		"channel": fmt.Sprintf("task:%s", task.ID.String()),
	}
	err = conn.WriteJSON(subscribeMessage)
	require.NoError(t, err)

	// Start planning and wait for WebSocket notification
	suite.startTaskPlanning(t, task.ID)

	// Collect WebSocket messages with timeout
	timeout := time.After(5 * time.Second)
	receivedTaskUpdate := false

	for !receivedTaskUpdate {
		select {
		case message := <-messages:
			if message["type"] == "task_update" {
				taskData, ok := message["data"].(map[string]interface{})
				if ok && taskData["id"] == task.ID.String() {
					receivedTaskUpdate = true
				}
			}
		case <-timeout:
			t.Fatal("Timeout waiting for WebSocket task update")
		}
	}

	assert.True(t, receivedTaskUpdate, "Should receive task update via WebSocket")
}

// Helper methods for test operations

func (suite *E2ETestSuite) startTaskPlanning(t *testing.T, taskID uuid.UUID) *dto.ExecutionResponse {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/plan", suite.GetServerURL(), taskID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response dto.ExecutionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response
}

func (suite *E2ETestSuite) approvePlan(t *testing.T, planID uuid.UUID) {
	url := fmt.Sprintf("%s/api/v1/plans/%s/approve", suite.GetServerURL(), planID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (suite *E2ETestSuite) startTaskImplementation(t *testing.T, taskID uuid.UUID) *dto.ExecutionResponse {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/implement", suite.GetServerURL(), taskID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response dto.ExecutionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response
}

func (suite *E2ETestSuite) completeTask(t *testing.T, taskID uuid.UUID) {
	url := fmt.Sprintf("%s/api/v1/tasks/%s/complete", suite.GetServerURL(), taskID.String())
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (suite *E2ETestSuite) createPullRequest(t *testing.T, taskID uuid.UUID, title, body string) *dto.PullRequestResponse {
	requestBody := map[string]interface{}{
		"title": title,
		"body":  body,
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	url := fmt.Sprintf("%s/api/v1/tasks/%s/pull-request", suite.GetServerURL(), taskID.String())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response dto.PullRequestResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response
}