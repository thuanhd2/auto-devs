package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ExecutionStatus represents the current status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusPaused    ExecutionStatus = "paused"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// Execution represents an AI execution instance
type Execution struct {
	ID          string           `json:"id"`
	TaskID      string           `json:"task_id"`
	Plan        Plan             `json:"plan"`
	Status      ExecutionStatus  `json:"status"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Error       string           `json:"error,omitempty"`
	Progress    float64          `json:"progress"` // 0.0 to 1.0
	Logs        []string         `json:"logs"`
	Result      *ExecutionResult `json:"result,omitempty"`

	// Internal fields
	processID string
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	Output   string                 `json:"output"`
	Files    []string               `json:"files"`
	Metrics  map[string]interface{} `json:"metrics"`
	Duration time.Duration          `json:"duration"`
}

// ExecutionUpdate represents a real-time update for an execution
type ExecutionUpdate struct {
	ExecutionID string          `json:"execution_id"`
	Status      ExecutionStatus `json:"status"`
	Progress    float64         `json:"progress"`
	Log         string          `json:"log,omitempty"`
	Error       string          `json:"error,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// ExecutionService orchestrates the entire AI execution workflow
type ExecutionService struct {
	cliManager     *CLIManager
	processManager *ProcessManager
	executions     map[string]*Execution
	mu             sync.RWMutex

	// Callbacks for real-time updates
	onUpdate func(update ExecutionUpdate)
}

// NewExecutionService creates a new execution service
func NewExecutionService(cliManager *CLIManager, processManager *ProcessManager) *ExecutionService {
	return &ExecutionService{
		cliManager:     cliManager,
		processManager: processManager,
		executions:     make(map[string]*Execution),
	}
}

// SetUpdateCallback sets the callback for real-time updates
func (es *ExecutionService) SetUpdateCallback(callback func(update ExecutionUpdate)) {
	es.onUpdate = callback
}

// StartExecution starts a new AI execution
func (es *ExecutionService) StartExecution(taskID string, plan Plan) (*Execution, error) {
	executionID := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	execution := &Execution{
		ID:        executionID,
		TaskID:    taskID,
		Plan:      plan,
		Status:    ExecutionStatusPending,
		StartedAt: time.Now(),
		Progress:  0.0,
		Logs:      make([]string, 0),
		ctx:       ctx,
		cancel:    cancel,
	}

	es.mu.Lock()
	es.executions[executionID] = execution
	es.mu.Unlock()

	// Send initial update
	es.sendUpdate(executionID, ExecutionStatusPending, 0.0, "Execution started", "")

	// Start execution in background
	go es.runExecution(execution)

	return execution, nil
}

// GetExecution retrieves an execution by ID
func (es *ExecutionService) GetExecution(executionID string) (*Execution, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	execution, exists := es.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	return execution, nil
}

// CancelExecution cancels a running execution
func (es *ExecutionService) CancelExecution(executionID string) error {
	execution, err := es.GetExecution(executionID)
	if err != nil {
		return err
	}

	execution.mu.Lock()
	defer execution.mu.Unlock()

	if execution.Status == ExecutionStatusCompleted ||
		execution.Status == ExecutionStatusFailed ||
		execution.Status == ExecutionStatusCancelled {
		return fmt.Errorf("cannot cancel execution in status: %s", execution.Status)
	}

	// Cancel the context
	execution.cancel()

	// Update status
	execution.Status = ExecutionStatusCancelled
	now := time.Now()
	execution.CompletedAt = &now

	// Cleanup process if running
	if execution.processID != "" {
		if process, exists := es.processManager.GetProcess(execution.processID); exists {
			es.processManager.KillProcess(process)
		}
	}

	es.sendUpdate(executionID, ExecutionStatusCancelled, execution.Progress, "Execution cancelled", "")

	return nil
}

// PauseExecution pauses a running execution
func (es *ExecutionService) PauseExecution(executionID string) error {
	execution, err := es.GetExecution(executionID)
	if err != nil {
		return err
	}

	execution.mu.Lock()
	defer execution.mu.Unlock()

	if execution.Status != ExecutionStatusRunning {
		return fmt.Errorf("cannot pause execution in status: %s", execution.Status)
	}

	// Note: ProcessManager doesn't support pause/resume yet
	// For now, we'll just update the status
	execution.Status = ExecutionStatusPaused
	es.sendUpdate(executionID, ExecutionStatusPaused, execution.Progress, "Execution paused", "")

	return nil
}

// ResumeExecution resumes a paused execution
func (es *ExecutionService) ResumeExecution(executionID string) error {
	execution, err := es.GetExecution(executionID)
	if err != nil {
		return err
	}

	execution.mu.Lock()
	defer execution.mu.Unlock()

	if execution.Status != ExecutionStatusPaused {
		return fmt.Errorf("cannot resume execution in status: %s", execution.Status)
	}

	// Note: ProcessManager doesn't support pause/resume yet
	// For now, we'll just update the status
	execution.Status = ExecutionStatusRunning
	es.sendUpdate(executionID, ExecutionStatusRunning, execution.Progress, "Execution resumed", "")

	return nil
}

// runExecution runs the actual execution workflow
func (es *ExecutionService) runExecution(execution *Execution) {
	defer func() {
		// Cleanup on completion
		es.mu.Lock()
		delete(es.executions, execution.ID)
		es.mu.Unlock()
	}()

	execution.mu.Lock()
	execution.Status = ExecutionStatusRunning
	execution.mu.Unlock()

	es.sendUpdate(execution.ID, ExecutionStatusRunning, 0.0, "Starting AI execution", "")

	// Step 1: Prepare CLI command
	es.addLog(execution, "Preparing CLI command...")
	command, err := es.buildCommandFromPlan(execution.Plan)
	if err != nil {
		es.handleExecutionError(execution, fmt.Sprintf("Failed to build command: %v", err))
		return
	}

	es.updateProgress(execution, 0.1)

	// Step 2: Start process
	es.addLog(execution, "Starting AI process...")
	process, err := es.processManager.SpawnProcess(command, "")
	if err != nil {
		es.handleExecutionError(execution, fmt.Sprintf("Failed to start process: %v", err))
		return
	}

	execution.mu.Lock()
	execution.processID = process.ID
	execution.mu.Unlock()

	es.updateProgress(execution, 0.2)

	// Step 3: Monitor process
	es.addLog(execution, "Monitoring execution progress...")

	// Monitor process output
	go es.monitorProcessOutput(execution, process)

	// Wait for process completion
	select {
	case <-execution.ctx.Done():
		// Execution was cancelled
		return
	default:
		// Wait for process to complete
		for process.IsRunning() {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-execution.ctx.Done():
				return
			default:
			}
		}
		es.handleExecutionCompletion(execution, process)
	}
}

// buildCommandFromPlan builds a CLI command from a plan
func (es *ExecutionService) buildCommandFromPlan(plan Plan) (string, error) {
	// Simple command building - in a real implementation this would be more sophisticated
	command := fmt.Sprintf("claude-code --plan '%s' --task-id '%s'", plan.ID, plan.TaskID)
	return command, nil
}

// monitorProcessOutput monitors the process output and updates progress
func (es *ExecutionService) monitorProcessOutput(execution *Execution, process *Process) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-execution.ctx.Done():
			return
		case <-ticker.C:
			// Check if process is still running
			if !process.IsRunning() {
				return
			}

			// Get current output
			stdout, stderr := process.GetOutput()
			if len(stdout) > 0 {
				output := string(stdout)
				es.addLog(execution, output)

				// Update progress based on output patterns
				progress := es.estimateProgress(output)
				if progress > execution.Progress {
					es.updateProgress(execution, progress)
				}
			}

			if len(stderr) > 0 {
				errorOutput := string(stderr)
				es.addLog(execution, fmt.Sprintf("Error: %s", errorOutput))
			}
		}
	}
}

// estimateProgress estimates progress based on output patterns
func (es *ExecutionService) estimateProgress(output string) float64 {
	// Convert to lowercase for case-insensitive matching
	outputLower := strings.ToLower(output)

	// Simple progress estimation based on output keywords
	// In a real implementation, this would be more sophisticated
	if strings.Contains(outputLower, "completed") || strings.Contains(outputLower, "done") {
		return 1.0
	} else if strings.Contains(outputLower, "processing") || strings.Contains(outputLower, "running") {
		return 0.5
	} else if strings.Contains(outputLower, "starting") || strings.Contains(outputLower, "initializing") {
		return 0.2
	}
	return 0.0
}

// handleExecutionCompletion handles successful execution completion
func (es *ExecutionService) handleExecutionCompletion(execution *Execution, process *Process) {
	execution.mu.Lock()
	defer execution.mu.Unlock()

	now := time.Now()
	execution.CompletedAt = &now

	// Get process output
	stdout, stderr := process.GetOutput()

	// Check if process completed successfully
	if process.ExitCode != nil && *process.ExitCode == 0 {
		execution.Status = ExecutionStatusCompleted
		execution.Progress = 1.0

		// Parse result from process output
		result := &ExecutionResult{
			Output:   string(stdout),
			Files:    []string{}, // Parse generated files
			Metrics:  make(map[string]interface{}),
			Duration: now.Sub(execution.StartedAt),
		}
		execution.Result = result

		es.addLog(execution, "Execution completed successfully")
		es.sendUpdate(execution.ID, ExecutionStatusCompleted, 1.0, "Execution completed successfully", "")
	} else {
		exitCode := -1
		if process.ExitCode != nil {
			exitCode = *process.ExitCode
		}
		errorMsg := fmt.Sprintf("Process failed with exit code: %d", exitCode)
		if len(stderr) > 0 {
			errorMsg += fmt.Sprintf(" - Error: %s", string(stderr))
		}
		es.handleExecutionError(execution, errorMsg)
	}
}

// handleExecutionError handles execution errors
func (es *ExecutionService) handleExecutionError(execution *Execution, errorMsg string) {
	execution.mu.Lock()
	defer execution.mu.Unlock()

	now := time.Now()
	execution.CompletedAt = &now
	execution.Status = ExecutionStatusFailed
	execution.Error = errorMsg

	es.addLog(execution, fmt.Sprintf("Execution failed: %s", errorMsg))
	es.sendUpdate(execution.ID, ExecutionStatusFailed, execution.Progress, "", errorMsg)
}

// addLog adds a log entry to the execution
func (es *ExecutionService) addLog(execution *Execution, message string) {
	execution.mu.Lock()
	defer execution.mu.Unlock()

	execution.Logs = append(execution.Logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message))
}

// updateProgress updates the execution progress
func (es *ExecutionService) updateProgress(execution *Execution, progress float64) {
	execution.mu.Lock()
	defer execution.mu.Unlock()

	if progress > execution.Progress {
		execution.Progress = progress
		es.sendUpdate(execution.ID, execution.Status, progress, "", "")
	}
}

// sendUpdate sends a real-time update
func (es *ExecutionService) sendUpdate(executionID string, status ExecutionStatus, progress float64, log, error string) {
	if es.onUpdate != nil {
		update := ExecutionUpdate{
			ExecutionID: executionID,
			Status:      status,
			Progress:    progress,
			Log:         log,
			Error:       error,
			Timestamp:   time.Now(),
		}
		es.onUpdate(update)
	}
}

// ListExecutions returns all active executions
func (es *ExecutionService) ListExecutions() []*Execution {
	es.mu.RLock()
	defer es.mu.RUnlock()

	executions := make([]*Execution, 0, len(es.executions))
	for _, execution := range es.executions {
		executions = append(executions, execution)
	}

	return executions
}

// CleanupCompletedExecutions removes completed executions from memory
func (es *ExecutionService) CleanupCompletedExecutions() {
	es.mu.Lock()
	defer es.mu.Unlock()

	for id, execution := range es.executions {
		if execution.Status == ExecutionStatusCompleted ||
			execution.Status == ExecutionStatusFailed ||
			execution.Status == ExecutionStatusCancelled {
			delete(es.executions, id)
		}
	}
}
