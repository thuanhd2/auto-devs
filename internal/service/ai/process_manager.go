package ai

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Process represents an AI execution process
type Process struct {
	ID          string
	Command     string
	WorkDir     string
	PID         int
	Status      ProcessStatus
	StartTime   time.Time
	EndTime     *time.Time
	ExitCode    *int
	Stdout      []byte
	Stderr      []byte
	Error       error
	ctx         context.Context
	cancel      context.CancelFunc
	cmd         *exec.Cmd
	mu          sync.RWMutex
	resourceMu  sync.RWMutex
	CPUUsage    float64
	MemoryUsage uint64
}

// ProcessStatus represents the current status of a process
type ProcessStatus string

const (
	ProcessStatusStarting ProcessStatus = "starting"
	ProcessStatusRunning  ProcessStatus = "running"
	ProcessStatusStopped  ProcessStatus = "stopped"
	ProcessStatusKilled   ProcessStatus = "killed"
	ProcessStatusError    ProcessStatus = "error"
)

// ProcessManager manages AI execution processes
type ProcessManager struct {
	processes map[string]*Process
	mu        sync.RWMutex
}

// NewProcessManager creates a new ProcessManager instance
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

// SpawnProcess creates and starts a new AI execution process
func (pm *ProcessManager) SpawnProcess(command string, workDir string) (*Process, error) {
	// Generate unique process ID
	processID := generateProcessID()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create process instance
	process := &Process{
		ID:        processID,
		Command:   command,
		WorkDir:   workDir,
		Status:    ProcessStatusStarting,
		StartTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Parse command and arguments
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir

	// Setup environment variables
	cmd.Env = append(os.Environ(),
		"AI_PROCESS_ID="+processID,
		"AI_WORK_DIR="+workDir,
	)

	// Setup stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		process.Status = ProcessStatusError
		process.Error = fmt.Errorf("failed to create stdout pipe: %w", err)
		return process, process.Error
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		process.Status = ProcessStatusError
		process.Error = fmt.Errorf("failed to create stderr pipe: %w", err)
		return process, process.Error
	}

	process.cmd = cmd

	// Start the process
	if err := cmd.Start(); err != nil {
		process.Status = ProcessStatusError
		process.Error = fmt.Errorf("failed to start process: %w", err)
		return process, process.Error
	}

	// Set PID and update status
	process.PID = cmd.Process.Pid
	process.Status = ProcessStatusRunning

	// Add to process manager
	pm.mu.Lock()
	pm.processes[processID] = process
	pm.mu.Unlock()

	// Start monitoring in background
	go pm.MonitorProcess(process)

	// Start output collection in background
	go pm.collectOutput(process, stdout, stderr)

	return process, nil
}

// MonitorProcess monitors the status and resource usage of a process
func (pm *ProcessManager) MonitorProcess(process *Process) error {
	// Wait for process to complete
	err := process.cmd.Wait()

	process.mu.Lock()
	defer process.mu.Unlock()

	// Update process status based on result
	if err != nil {
		process.Status = ProcessStatusError
		process.Error = err
	} else {
		process.Status = ProcessStatusStopped
	}

	// Set end time and exit code
	now := time.Now()
	process.EndTime = &now
	if process.cmd.ProcessState != nil {
		exitCode := process.cmd.ProcessState.ExitCode()
		process.ExitCode = &exitCode
	}

	// Cleanup process from manager when done
	pm.mu.Lock()
	delete(pm.processes, process.ID)
	pm.mu.Unlock()

	return err
}

// TerminateProcess gracefully terminates a process using SIGTERM
func (pm *ProcessManager) TerminateProcess(process *Process) error {
	process.mu.Lock()
	defer process.mu.Unlock()

	if process.Status != ProcessStatusRunning {
		return fmt.Errorf("process %s is not running (status: %s)", process.ID, process.Status)
	}

	// Send SIGTERM signal
	if process.cmd.Process != nil {
		if err := process.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to send SIGTERM to process %s: %w", process.ID, err)
		}
	}

	// Cancel context to stop the process
	process.cancel()

	return nil
}

// KillProcess forcefully kills a process using SIGKILL
func (pm *ProcessManager) KillProcess(process *Process) error {
	process.mu.Lock()
	defer process.mu.Unlock()

	if process.Status == ProcessStatusStopped || process.Status == ProcessStatusKilled {
		return fmt.Errorf("process %s is already stopped (status: %s)", process.ID, process.Status)
	}

	// Send SIGKILL signal
	if process.cmd.Process != nil {
		if err := process.cmd.Process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to send SIGKILL to process %s: %w", process.ID, err)
		}
	}

	// Update status
	process.Status = ProcessStatusKilled

	// Cancel context
	process.cancel()

	// Set end time
	now := time.Now()
	process.EndTime = &now

	// Cleanup from manager immediately
	pm.mu.Lock()
	delete(pm.processes, process.ID)
	pm.mu.Unlock()

	return nil
}

// GetProcess retrieves a process by ID
func (pm *ProcessManager) GetProcess(processID string) (*Process, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	process, exists := pm.processes[processID]
	return process, exists
}

// ListProcesses returns all active processes
func (pm *ProcessManager) ListProcesses() []*Process {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	processes := make([]*Process, 0, len(pm.processes))
	for _, process := range pm.processes {
		processes = append(processes, process)
	}

	return processes
}

// collectOutput collects stdout and stderr from the process
func (pm *ProcessManager) collectOutput(process *Process, stdout, stderr io.ReadCloser) {
	var wg sync.WaitGroup

	// Collect stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stdout.Close()
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if n > 0 {
				process.mu.Lock()
				process.Stdout = append(process.Stdout, buffer[:n]...)
				process.mu.Unlock()
			}
			if err != nil {
				break
			}
		}
	}()

	// Collect stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stderr.Close()
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if n > 0 {
				process.mu.Lock()
				process.Stderr = append(process.Stderr, buffer[:n]...)
				process.mu.Unlock()
			}
			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
}

// generateProcessID generates a unique process ID
func generateProcessID() string {
	return fmt.Sprintf("ai_process_%d", time.Now().UnixNano())
}

// GetStatus returns the current status of the process
func (p *Process) GetStatus() ProcessStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status
}

// GetOutput returns the collected stdout and stderr
func (p *Process) GetOutput() ([]byte, []byte) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Stdout, p.Stderr
}

// GetResourceUsage returns the current resource usage
func (p *Process) GetResourceUsage() (float64, uint64) {
	p.resourceMu.RLock()
	defer p.resourceMu.RUnlock()
	return p.CPUUsage, p.MemoryUsage
}

// IsRunning checks if the process is currently running
func (p *Process) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status == ProcessStatusRunning
}

// GetDuration returns the duration the process has been running
func (p *Process) GetDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.EndTime != nil {
		return p.EndTime.Sub(p.StartTime)
	}
	return time.Since(p.StartTime)
}
