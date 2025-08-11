package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
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

type FakeAiCodingCli struct{}

func (f *FakeAiCodingCli) GetPlanningCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootPath := filepath.Join(projectPath, "../../../")
	fakeCliPath := filepath.Join(projectRootPath, "fake-cli", "fake.sh")

	return fakeCliPath, "hello world", nil
}

func (f *FakeAiCodingCli) GetImplementationCommand(ctx context.Context, task *entity.Task) (string, string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	projectRootPath := filepath.Join(projectPath, "../../../")
	fakeCliPath := filepath.Join(projectRootPath, "fake-cli", "fake.sh")
	return fakeCliPath, "hello world", nil
}

func (f *FakeAiCodingCli) ParseOutputToLogs(output string) []*entity.ExecutionLog {
	lines := strings.Split(output, "\n")
	logs := make([]*entity.ExecutionLog, len(lines))
	for i, line := range lines {
		logs[i] = &entity.ExecutionLog{
			Message: line,
			Level:   entity.LogLevelInfo,
			Source:  "stdout",
			Line:    i,
		}
	}
	return logs
}

func NewFakeAiCodingCli() AiCodingCli {
	return &FakeAiCodingCli{}
}

func TestExecutionService_StartExecution(t *testing.T) {
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	worktreePath := "testdata/worktree"
	task := entity.Task{
		ID:           uuid.New(),
		WorktreePath: &worktreePath,
	}

	execution, err := es.StartExecution(&task, NewFakeAiCodingCli(), true)
	require.NoError(t, err)

	assert.NotNil(t, execution)
	assert.Equal(t, task.ID.String(), execution.TaskID)
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

	worktreePath := "testdata/worktree"
	task := entity.Task{
		ID:           uuid.New(),
		WorktreePath: &worktreePath,
	}

	execution, err := es.StartExecution(&task, NewFakeAiCodingCli(), true)
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
	// TODO: skip for now, back later
	t.Skip("skip for now, back later!")
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	worktreePath := "testdata/worktree"
	task := entity.Task{
		ID:           uuid.New(),
		WorktreePath: &worktreePath,
	}

	execution, err := es.StartExecution(&task, NewFakeAiCodingCli(), true)
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
	// TODO: skip for now, back later
	t.Skip("skip for now, back later!")
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	// Start multiple executions
	task1 := entity.Task{
		ID: uuid.New(),
	}
	task2 := entity.Task{
		ID: uuid.New(),
	}

	execution1, err := es.StartExecution(&task1, NewFakeAiCodingCli(), true)
	require.NoError(t, err)

	execution2, err := es.StartExecution(&task2, NewFakeAiCodingCli(), true)
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
	// TODO: skip for now, back later
	t.Skip("skip for now, back later!")
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	require.NoError(t, err)

	processManager := NewProcessManager()
	es := NewExecutionService(cliManager, processManager)

	// Track updates
	var updates []ExecutionUpdate
	es.SetUpdateCallback(func(update ExecutionUpdate) {
		updates = append(updates, update)
	})

	worktreePath := "testdata/worktree"
	task := entity.Task{
		ID:           uuid.New(),
		WorktreePath: &worktreePath,
	}

	execution, err := es.StartExecution(&task, NewFakeAiCodingCli(), true)
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
	// TODO: skip for now, back later
	t.Skip("skip for now, back later!")
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
