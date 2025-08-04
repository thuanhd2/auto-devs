package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProcessManager_SpawnProcess(t *testing.T) {
	pm := NewProcessManager()

	// Create temporary working directory
	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test spawning a simple command
	command := "echo 'Hello World'"
	process, err := pm.SpawnProcess(command, tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	if process == nil {
		t.Fatal("Process should not be nil")
	}

	if process.ID == "" {
		t.Error("Process ID should not be empty")
	}

	if process.Command != command {
		t.Errorf("Expected command %s, got %s", command, process.Command)
	}

	if process.WorkDir != tempDir {
		t.Errorf("Expected workDir %s, got %s", tempDir, process.WorkDir)
	}

	if process.PID <= 0 {
		t.Error("Process PID should be positive")
	}

	// Wait for process to complete
	time.Sleep(200 * time.Millisecond)

	// Check if process is in the list (it should still be there during execution)
	processes := pm.ListProcesses()
	found := false
	for _, p := range processes {
		if p.ID == process.ID {
			found = true
			break
		}
	}

	// Process might be removed from list if it completed quickly
	// So we check if it was found OR if it's no longer running
	if !found && process.IsRunning() {
		t.Error("Process should be in the manager's process list if still running")
	}
}

func TestProcessManager_GetProcess(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	process, err := pm.SpawnProcess("echo 'test'", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Test getting existing process
	retrievedProcess, exists := pm.GetProcess(process.ID)
	if !exists {
		t.Error("Process should exist")
	}

	if retrievedProcess.ID != process.ID {
		t.Errorf("Expected process ID %s, got %s", process.ID, retrievedProcess.ID)
	}

	// Test getting non-existent process
	_, exists = pm.GetProcess("non_existent_id")
	if exists {
		t.Error("Non-existent process should not exist")
	}
}

func TestProcessManager_TerminateProcess(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Spawn a long-running process
	process, err := pm.SpawnProcess("sleep 10", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Wait a bit for process to start
	time.Sleep(100 * time.Millisecond)

	// Terminate the process
	err = pm.TerminateProcess(process)
	if err != nil {
		t.Fatalf("Failed to terminate process: %v", err)
	}

	// Wait for termination
	time.Sleep(100 * time.Millisecond)

	// Check if process is no longer running
	if process.IsRunning() {
		t.Error("Process should not be running after termination")
	}
}

func TestProcessManager_KillProcess(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Spawn a long-running process
	process, err := pm.SpawnProcess("sleep 10", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Wait a bit for process to start
	time.Sleep(100 * time.Millisecond)

	// Kill the process
	err = pm.KillProcess(process)
	if err != nil {
		t.Fatalf("Failed to kill process: %v", err)
	}

	// Wait for termination
	time.Sleep(200 * time.Millisecond)

	// Check if process status is killed or error (both are acceptable for killed process)
	status := process.GetStatus()
	if status != ProcessStatusKilled && status != ProcessStatusError {
		t.Errorf("Expected status %s or %s, got %s", ProcessStatusKilled, ProcessStatusError, status)
	}
}

func TestProcess_GetOutput(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Spawn a process that produces output
	process, err := pm.SpawnProcess("echo 'stdout message' && echo 'stderr message' >&2", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Wait for process to complete
	time.Sleep(200 * time.Millisecond)

	// Get output
	stdout, stderr := process.GetOutput()

	// Check stdout
	expectedStdout := "stdout message\n"
	if string(stdout) != expectedStdout {
		t.Errorf("Expected stdout %q, got %q", expectedStdout, string(stdout))
	}

	// Check stderr
	expectedStderr := "stderr message\n"
	if string(stderr) != expectedStderr {
		t.Errorf("Expected stderr %q, got %q", expectedStderr, string(stderr))
	}
}

func TestProcess_GetDuration(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	process, err := pm.SpawnProcess("echo 'test'", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Wait for process to complete
	time.Sleep(100 * time.Millisecond)

	duration := process.GetDuration()
	if duration <= 0 {
		t.Error("Process duration should be positive")
	}

	if duration > 1*time.Second {
		t.Error("Process duration should be reasonable (less than 1 second)")
	}
}

func TestProcessManager_EnvironmentVariables(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Spawn a process that checks environment variables
	process, err := pm.SpawnProcess("echo $AI_PROCESS_ID && echo $AI_WORK_DIR", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Wait for process to complete
	time.Sleep(100 * time.Millisecond)

	stdout, _ := process.GetOutput()
	output := string(stdout)

	// Check if AI_PROCESS_ID is set
	if !contains(output, process.ID) {
		t.Errorf("Output should contain process ID %s, got: %s", process.ID, output)
	}

	// Check if AI_WORK_DIR is set
	if !contains(output, tempDir) {
		t.Errorf("Output should contain work directory %s, got: %s", tempDir, output)
	}
}

func TestProcessManager_WorkingDirectory(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file in the temp directory
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Spawn a process that reads the file
	process, err := pm.SpawnProcess("cat test.txt", tempDir)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Wait for process to complete
	time.Sleep(100 * time.Millisecond)

	stdout, _ := process.GetOutput()
	output := string(stdout)

	expected := "test content"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			contains(s[1:len(s)-1], substr)))
}
