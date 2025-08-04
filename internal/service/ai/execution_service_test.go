package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionService(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()

	es := NewExecutionService(cliManager, processManager)

	assert.NotNil(t, es)
	assert.NotNil(t, es.cliManager)
	assert.NotNil(t, es.processManager)
	assert.NotNil(t, es.executions)
}

func TestExecutionService_StartExecution(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	plan := Plan{
		ID:          "test-plan-1",
		TaskID:      "test-task-1",
		Description: "Test plan",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	execution, err := es.StartExecution("test-task-1", plan)
	require.NoError(t, err)

	assert.NotNil(t, execution)
	assert.Equal(t, "test-task-1", execution.TaskID)
	assert.Equal(t, ExecutionStatusPending, execution.Status)
	assert.Equal(t, 0.0, execution.Progress)
	assert.NotEmpty(t, execution.ID)
	assert.NotNil(t, execution.StartedAt)

	// Wait a bit for execution to start
	time.Sleep(100 * time.Millisecond)

	// Check that execution is now in memory
	retrieved, err := es.GetExecution(execution.ID)
	require.NoError(t, err)
	assert.Equal(t, execution.ID, retrieved.ID)
}

func TestExecutionService_GetExecution(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	// Test getting non-existent execution
	_, err = es.GetExecution("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution not found")
}

func TestExecutionService_CancelExecution(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	plan := Plan{
		ID:          "test-plan-2",
		TaskID:      "test-task-2",
		Description: "Test plan for cancellation",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	execution, err := es.StartExecution("test-task-2", plan)
	require.NoError(t, err)

	// Wait a bit for execution to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the execution
	err = es.CancelExecution(execution.ID)
	require.NoError(t, err)

	// Check that execution was cancelled
	retrieved, err := es.GetExecution(execution.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusCancelled, retrieved.Status)
	assert.NotNil(t, retrieved.CompletedAt)
}

func TestExecutionService_PauseResumeExecution(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	plan := Plan{
		ID:          "test-plan-3",
		TaskID:      "test-task-3",
		Description: "Test plan for pause/resume",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	execution, err := es.StartExecution("test-task-3", plan)
	require.NoError(t, err)

	// Wait a bit for execution to start
	time.Sleep(100 * time.Millisecond)

	// Try to pause (should work even though ProcessManager doesn't support it yet)
	err = es.PauseExecution(execution.ID)
	require.NoError(t, err)

	retrieved, err := es.GetExecution(execution.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusPaused, retrieved.Status)

	// Resume the execution
	err = es.ResumeExecution(execution.ID)
	require.NoError(t, err)

	retrieved, err = es.GetExecution(execution.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusRunning, retrieved.Status)
}

func TestExecutionService_ListExecutions(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	// Start multiple executions
	plan1 := Plan{
		ID:          "test-plan-4",
		TaskID:      "test-task-4",
		Description: "Test plan 1",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	plan2 := Plan{
		ID:          "test-plan-5",
		TaskID:      "test-task-5",
		Description: "Test plan 2",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	execution1, err := es.StartExecution("test-task-4", plan1)
	require.NoError(t, err)

	execution2, err := es.StartExecution("test-task-5", plan2)
	require.NoError(t, err)

	// Wait a bit for executions to start
	time.Sleep(100 * time.Millisecond)

	// List executions
	executions := es.ListExecutions()
	assert.Len(t, executions, 2)

	// Check that both executions are in the list
	executionIDs := make(map[string]bool)
	for _, exec := range executions {
		executionIDs[exec.ID] = true
	}

	assert.True(t, executionIDs[execution1.ID])
	assert.True(t, executionIDs[execution2.ID])
}

func TestExecutionService_RealTimeUpdates(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	// Track updates
	var updates []ExecutionUpdate
	es.SetUpdateCallback(func(update ExecutionUpdate) {
		updates = append(updates, update)
	})

	plan := Plan{
		ID:          "test-plan-6",
		TaskID:      "test-task-6",
		Description: "Test plan for real-time updates",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	execution, err := es.StartExecution("test-task-6", plan)
	require.NoError(t, err)

	// Wait a bit for updates to be sent
	time.Sleep(200 * time.Millisecond)

	// Check that we received updates
	assert.Greater(t, len(updates), 0)

	// Check that first update is for pending status
	if len(updates) > 0 {
		assert.Equal(t, execution.ID, updates[0].ExecutionID)
		assert.Equal(t, ExecutionStatusPending, updates[0].Status)
		assert.Equal(t, 0.0, updates[0].Progress)
	}
}

func TestExecutionService_BuildCommandFromPlan(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	plan := Plan{
		ID:          "test-plan-7",
		TaskID:      "test-task-7",
		Description: "Test plan for command building",
		Steps:       []PlanStep{},
		Context:     map[string]string{},
		CreatedAt:   time.Now(),
	}

	command, err := es.buildCommandFromPlan(plan)
	require.NoError(t, err)

	assert.Contains(t, command, "claude-code")
	assert.Contains(t, command, plan.ID)
	assert.Contains(t, command, plan.TaskID)
}

func TestExecutionService_EstimateProgress(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	// Test progress estimation
	assert.Equal(t, 1.0, es.estimateProgress("Task completed successfully"))
	assert.Equal(t, 1.0, es.estimateProgress("All done"))
	assert.Equal(t, 0.5, es.estimateProgress("Processing data"))
	assert.Equal(t, 0.5, es.estimateProgress("Running analysis"))
	assert.Equal(t, 0.2, es.estimateProgress("Starting initialization"))
	assert.Equal(t, 0.2, es.estimateProgress("Initializing system"))
	assert.Equal(t, 0.0, es.estimateProgress("Random output"))
}
