package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// ExecutionStatus represents the current status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "PENDING"
	ExecutionStatusRunning   ExecutionStatus = "RUNNING"
	ExecutionStatusPaused    ExecutionStatus = "PAUSED"
	ExecutionStatusCompleted ExecutionStatus = "COMPLETED"
	ExecutionStatusFailed    ExecutionStatus = "FAILED"
	ExecutionStatusCancelled ExecutionStatus = "CANCELLED"
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
	Command     string           `json:"command"`
	Input       string           `json:"input"`
	WorkingDir  string           `json:"working_dir"`

	// Internal fields
	processID     string
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	stdoutChannel chan string
	stderrChannel chan string
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

type AiCodingCli interface {
	GetPlanningCommand(context.Context, *entity.Task) (string, string, error)
	GetImplementationCommand(context.Context, *entity.Task) (string, string, error)
	ParseOutputToLogs(output string) []*entity.ExecutionLog
	ParseOutputToPlan(output string) (string, error)
}

// StartExecution starts a new AI execution
func (es *ExecutionService) StartExecution(task *entity.Task, cli AiCodingCli, isForPlanning bool) (*Execution, error) {
	executionID := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	var command, input string
	var err error
	if isForPlanning {
		command, input, err = cli.GetPlanningCommand(ctx, task)
	} else {
		command, input, err = cli.GetImplementationCommand(ctx, task)
	}
	if err != nil {
		return nil, err
	}

	if task.WorktreePath == nil {
		return nil, fmt.Errorf("worktree path is not set")
	}

	workingDir := *task.WorktreePath

	execution := &Execution{
		ID:         executionID,
		TaskID:     task.ID.String(),
		Status:     ExecutionStatusPending,
		StartedAt:  time.Now(),
		Progress:   0.0,
		Logs:       make([]string, 0),
		ctx:        ctx,
		cancel:     cancel,
		Command:    command,
		Input:      input,
		WorkingDir: workingDir,
	}

	es.mu.Lock()
	es.executions[executionID] = execution
	es.mu.Unlock()

	return execution, nil
}

func (es *ExecutionService) RunExecution(execution *Execution) (*Execution, error) {
	go es.runExecution(execution)
	return execution, nil
}

// RegisterStdoutChannel registers a channel for stdout output
func (exe *Execution) RegisterStdoutChannel(channel chan string) {
	exe.mu.Lock()
	defer exe.mu.Unlock()
	exe.stdoutChannel = channel
}

// RegisterStderrChannel registers a channel for stderr output
func (exe *Execution) RegisterStderrChannel(channel chan string) {
	exe.mu.Lock()
	defer exe.mu.Unlock()
	exe.stderrChannel = channel
}

// Get execution context Done channel
func (exe *Execution) GetContextDoneChannel() <-chan struct{} {
	return exe.ctx.Done()
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

	// Step 1: Prepare CLI command
	command := execution.Command

	// Step 2: Start process
	process, err := es.processManager.SpawnProcess(command, execution.WorkingDir, execution.Input)
	if err != nil {
		es.handleExecutionError(execution, fmt.Sprintf("Failed to start process: %v", err))
		return
	}

	execution.mu.Lock()
	execution.processID = process.ID
	execution.mu.Unlock()

	// Step 3: Monitor process
	// Monitor process output
	go es.monitorProcessOutput(execution, process)

	defer es.handleExecutionCompletion(execution, process)

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
				log.Println("Execution cancelled", execution.ID)
				return
			case <-process.ctx.Done():
				log.Println("Process cancelled", process.ID)
				return
			default:
			}
		}
	}
}

// buildCommandFromPlan builds a CLI command from a plan
func (es *ExecutionService) buildCommandFromPlan(plan Plan) (string, error) {
	// Simple command building - in a real implementation this would be more sophisticated
	// command := fmt.Sprintf("claude-code --plan '%s' --task-id '%s'", plan.ID, plan.TaskID)
	command := "echo 'Hello, world!' | /Users/thuanho/Documents/personal/auto-devs/fake-cli/fake.sh"
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
				// es.addLog(execution, output)

				// // Update progress based on output patterns
				// progress := es.estimateProgress(output)
				// if progress > execution.Progress {
				// 	es.updateProgress(execution, progress)
				// }
				execution.stdoutChannel <- output
			}

			if len(stderr) > 0 {
				errorOutput := string(stderr)
				// es.addLog(execution, fmt.Sprintf("Error: %s", errorOutput))
				execution.stderrChannel <- errorOutput
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
	defer func() {
		// should sleep 1 second to make sure the process is finished and logs are saved
		time.Sleep(1 * time.Second)
		execution.cancel()
		execution.mu.Unlock()
	}()

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
}

// addLog adds a log entry to the execution
func (es *ExecutionService) addLog(execution *Execution, message string) {
	execution.mu.Lock()
	defer execution.mu.Unlock()
	logMessage := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message)
	log.Println("addLog", execution.ID, logMessage)
	execution.Logs = append(execution.Logs, logMessage)
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
