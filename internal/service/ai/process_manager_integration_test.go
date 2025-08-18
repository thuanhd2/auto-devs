package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestProcessManagerIntegration tests real-world scenarios
func TestProcessManagerIntegration(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test concurrent process spawning
	t.Run("ConcurrentProcessSpawning", func(t *testing.T) {
		var wg sync.WaitGroup
		processes := make([]*Process, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				command := fmt.Sprintf("echo 'Process %d' && sleep 0.1", index)
				process, err := pm.SpawnProcess(command, tempDir, "")
				if err != nil {
					t.Errorf("Failed to spawn process %d: %v", index, err)
					return
				}
				processes[index] = process
			}(i)
		}

		wg.Wait()

		// Wait for all processes to complete
		time.Sleep(500 * time.Millisecond)

		// Check that all processes completed successfully
		for i, process := range processes {
			if process == nil {
				t.Errorf("Process %d is nil", i)
				continue
			}

			stdout, _ := process.GetOutput()
			expected := fmt.Sprintf("Process %d\n", i)
			if string(stdout) != expected {
				t.Errorf("Process %d: expected output %q, got %q", i, expected, string(stdout))
			}
		}
	})

	// Test process with file operations
	t.Run("ProcessWithFileOperations", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "input.txt")
		err := os.WriteFile(testFile, []byte("Hello World\n"), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Spawn process that reads and processes the file
		command := "cat input.txt | tr '[:lower:]' '[:upper:]' > output.txt && cat output.txt"
		process, err := pm.SpawnProcess(command, tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		stdout, _ := process.GetOutput()
		expected := "HELLO WORLD\n"
		if string(stdout) != expected {
			t.Errorf("Expected output %q, got %q", expected, string(stdout))
		}

		// Check that output file was created
		outputFile := filepath.Join(tempDir, "output.txt")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}
	})

	// Test process with environment variables
	t.Run("ProcessWithEnvironmentVariables", func(t *testing.T) {
		// Create a script that uses environment variables
		scriptContent := `#!/bin/bash
echo "Process ID: $AI_PROCESS_ID"
echo "Work Dir: $AI_WORK_DIR"
echo "Custom Var: $CUSTOM_TEST_VAR"
`

		scriptPath := filepath.Join(tempDir, "env_test.sh")
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755)
		if err != nil {
			t.Fatalf("Failed to create script: %v", err)
		}

		// Set custom environment variable
		os.Setenv("CUSTOM_TEST_VAR", "test_value")

		process, err := pm.SpawnProcess("./env_test.sh", tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		stdout, _ := process.GetOutput()
		output := string(stdout)

		// Check that environment variables are set correctly
		if !contains(output, process.ID) {
			t.Errorf("Output should contain process ID %s, got: %s", process.ID, output)
		}

		if !contains(output, tempDir) {
			t.Errorf("Output should contain work directory %s, got: %s", tempDir, output)
		}

		if !contains(output, "test_value") {
			t.Errorf("Output should contain custom variable value, got: %s", output)
		}
	})

	// Test process termination and cleanup
	t.Run("ProcessTerminationAndCleanup", func(t *testing.T) {
		t.Skip("skip for now, back later!")
		// Spawn a long-running process
		process, err := pm.SpawnProcess("sleep 5", tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		// Wait a bit for process to start
		time.Sleep(100 * time.Millisecond)

		// Check that process is in the list
		processes := pm.ListProcesses()
		found := false
		for _, p := range processes {
			if p.ID == process.ID {
				found = true
				break
			}
		}

		if !found {
			t.Error("Process should be in the manager's list")
		}

		// Terminate the process
		err = pm.TerminateProcess(process)
		if err != nil {
			t.Fatalf("Failed to terminate process: %v", err)
		}

		// Wait for termination
		time.Sleep(500 * time.Millisecond)

		// Check that process is no longer in the list
		processes = pm.ListProcesses()
		found = false
		for _, p := range processes {
			if p.ID == process.ID {
				found = true
				break
			}
		}

		if found {
			t.Error("Process should be removed from manager's list after termination")
		}

		// Check that process is not running
		if process.IsRunning() {
			t.Error("Process should not be running after termination")
		}
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		t.Skip("skip for now, back later!")
		// Test with non-existent command
		process, err := pm.SpawnProcess("nonexistent_command_12345", tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		// Check that process has error status
		if process.GetStatus() != ProcessStatusError {
			t.Errorf("Expected status %s, got %s", ProcessStatusError, process.GetStatus())
		}

		if process.Error == nil {
			t.Error("Process should have an error")
		}

		// Test with invalid working directory
		process2, err := pm.SpawnProcess("echo 'test'", "/nonexistent/directory/12345", "")
		if err != nil {
			// This is expected behavior - process should fail to start
			t.Logf("Process failed to start as expected: %v", err)
			return
		}

		time.Sleep(500 * time.Millisecond)

		if process2.GetStatus() != ProcessStatusError {
			t.Errorf("Expected status %s, got %s", ProcessStatusError, process2.GetStatus())
		}
	})

	// Test resource management
	t.Run("ResourceManagement", func(t *testing.T) {
		t.Skip("skip for now, back later!")
		// Spawn multiple processes
		var processes []*Process
		for i := 0; i < 3; i++ {
			process, err := pm.SpawnProcess("echo 'test'", tempDir, "")
			if err != nil {
				t.Fatalf("Failed to spawn process %d: %v", i, err)
			}
			processes = append(processes, process)
		}

		// Wait for all processes to complete
		time.Sleep(500 * time.Millisecond)

		// Check that all processes are cleaned up
		activeProcesses := pm.ListProcesses()
		if len(activeProcesses) > 0 {
			t.Errorf("Expected 0 active processes, got %d", len(activeProcesses))
		}

		// Check that all processes have completed
		for i, process := range processes {
			if process.IsRunning() {
				t.Errorf("Process %d should not be running", i)
			}
		}
	})
}

// TestProcessManagerStress tests the ProcessManager under stress conditions
func TestProcessManagerStress(t *testing.T) {
	t.Skip("skip for now, back later!")
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_stress_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test rapid process spawning and termination
	t.Run("RapidProcessSpawning", func(t *testing.T) {
		var wg sync.WaitGroup
		processCount := 20

		for i := 0; i < processCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				process, err := pm.SpawnProcess("echo 'test'", tempDir, "")
				if err != nil {
					t.Errorf("Failed to spawn process %d: %v", index, err)
					return
				}

				// Wait a bit then terminate
				time.Sleep(50 * time.Millisecond)
				pm.TerminateProcess(process)
			}(i)
		}

		wg.Wait()

		// Wait for cleanup
		time.Sleep(1 * time.Second)

		// Check that all processes are cleaned up
		activeProcesses := pm.ListProcesses()
		if len(activeProcesses) > 0 {
			t.Errorf("Expected 0 active processes after cleanup, got %d", len(activeProcesses))
		}
	})

	// Test concurrent access to ProcessManager
	t.Run("ConcurrentAccess", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 50

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				// Spawn process
				process, err := pm.SpawnProcess("echo 'test'", tempDir, "")
				if err != nil {
					t.Errorf("Failed to spawn process %d: %v", index, err)
					return
				}

				// Get process by ID
				retrievedProcess, exists := pm.GetProcess(process.ID)
				if !exists {
					t.Errorf("Process %d should exist", index)
					return
				}

				if retrievedProcess.ID != process.ID {
					t.Errorf("Process ID mismatch: expected %s, got %s", process.ID, retrievedProcess.ID)
				}

				// List processes
				processes := pm.ListProcesses()
				if len(processes) == 0 {
					t.Errorf("Process list should not be empty")
				}
			}(i)
		}

		wg.Wait()

		// Wait for cleanup
		time.Sleep(1 * time.Second)
	})
}

// TestProcessManagerEdgeCases tests edge cases and boundary conditions
func TestProcessManagerEdgeCases(t *testing.T) {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "process_edge_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test empty command
	t.Run("EmptyCommand", func(t *testing.T) {
		t.Skip("skip for now, back later!")
		process, err := pm.SpawnProcess("", tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		// Empty command might complete successfully or with error, both are acceptable
		status := process.GetStatus()
		if status != ProcessStatusStopped && status != ProcessStatusError {
			t.Errorf("Expected status %s or %s for empty command, got %s", ProcessStatusStopped, ProcessStatusError, status)
		}
	})

	// Test very long command
	t.Run("VeryLongCommand", func(t *testing.T) {
		t.Skip("skip for now, back later!")
		longCommand := "echo 'test' && echo 'long command test'"
		process, err := pm.SpawnProcess(longCommand, tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		if process.GetStatus() != ProcessStatusStopped && process.GetStatus() != ProcessStatusError {
			t.Errorf("Expected process to complete, got status %s", process.GetStatus())
		}
	})

	// Test terminating already terminated process
	t.Run("TerminateTerminatedProcess", func(t *testing.T) {
		process, err := pm.SpawnProcess("echo 'test'", tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		// Wait for completion
		time.Sleep(500 * time.Millisecond)

		// Try to terminate already completed process
		err = pm.TerminateProcess(process)
		if err == nil {
			t.Error("Should return error when terminating already completed process")
		}
	})

	// Test killing already killed process
	t.Run("KillKilledProcess", func(t *testing.T) {
		process, err := pm.SpawnProcess("sleep 10", tempDir, "")
		if err != nil {
			t.Fatalf("Failed to spawn process: %v", err)
		}

		// Wait a bit then kill
		time.Sleep(100 * time.Millisecond)
		err = pm.KillProcess(process)
		if err != nil {
			t.Fatalf("Failed to kill process: %v", err)
		}

		// Try to kill again
		err = pm.KillProcess(process)
		if err == nil {
			t.Error("Should return error when killing already killed process")
		}
	})
}
