package ai

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// ExampleProcessManager demonstrates how to use ProcessManager
func ExampleProcessManager() {
	// Create a new ProcessManager
	pm := NewProcessManager()

	// Create temporary working directory
	tempDir, err := os.MkdirTemp("", "ai_process_example")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Example 1: Spawn a simple command
	fmt.Println("=== Example 1: Simple Command ===")
	process1, err := pm.SpawnProcess("echo 'Hello from AI Process!'", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	fmt.Printf("Process started: ID=%s, PID=%d\n", process1.ID, process1.PID)

	// Wait for completion
	time.Sleep(500 * time.Millisecond)

	// Get output
	stdout, stderr := process1.GetOutput()
	fmt.Printf("Stdout: %s", string(stdout))
	if len(stderr) > 0 {
		fmt.Printf("Stderr: %s", string(stderr))
	}

	fmt.Printf("Final status: %s\n", process1.GetStatus())
	if process1.ExitCode != nil {
		fmt.Printf("Exit code: %d\n", *process1.ExitCode)
	}

	// Example 2: Long-running process with monitoring
	fmt.Println("\n=== Example 2: Long-running Process ===")
	process2, err := pm.SpawnProcess("sleep 3 && echo 'Process completed after 3 seconds'", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	fmt.Printf("Process started: ID=%s, PID=%d\n", process2.ID, process2.PID)

	// Monitor the process
	go func() {
		for process2.IsRunning() {
			status := process2.GetStatus()
			duration := process2.GetDuration()
			fmt.Printf("  Status: %s, Duration: %v\n", status, duration)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Wait for completion
	time.Sleep(4 * time.Second)

	stdout, _ = process2.GetOutput()
	fmt.Printf("Final output: %s", string(stdout))

	// Example 3: Process with error handling
	fmt.Println("\n=== Example 3: Process with Error ===")
	process3, err := pm.SpawnProcess("nonexistent_command", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Printf("Process status: %s\n", process3.GetStatus())
	if process3.Error != nil {
		fmt.Printf("Error: %v\n", process3.Error)
	}

	// Example 4: Process termination
	fmt.Println("\n=== Example 4: Process Termination ===")
	process4, err := pm.SpawnProcess("sleep 10", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	fmt.Printf("Process started: ID=%s, PID=%d\n", process4.ID, process4.PID)

	// Wait a bit then terminate
	time.Sleep(1 * time.Second)

	fmt.Println("Terminating process...")
	err = pm.TerminateProcess(process4)
	if err != nil {
		log.Printf("Failed to terminate process: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Process status after termination: %s\n", process4.GetStatus())

	// Example 5: Process killing
	fmt.Println("\n=== Example 5: Process Killing ===")
	process5, err := pm.SpawnProcess("sleep 10", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	fmt.Printf("Process started: ID=%s, PID=%d\n", process5.ID, process5.PID)

	// Wait a bit then kill
	time.Sleep(1 * time.Second)

	fmt.Println("Killing process...")
	err = pm.KillProcess(process5)
	if err != nil {
		log.Printf("Failed to kill process: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Process status after killing: %s\n", process5.GetStatus())

	// Example 6: List all processes
	fmt.Println("\n=== Example 6: List Processes ===")
	processes := pm.ListProcesses()
	fmt.Printf("Active processes: %d\n", len(processes))
	for _, p := range processes {
		fmt.Printf("  - %s: %s (PID: %d)\n", p.ID, p.GetStatus(), p.PID)
	}
}

// ExampleProcessManagerWithContext demonstrates using ProcessManager with context
func ExampleProcessManagerWithContext() {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "ai_process_context")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Spawn a long-running process
	process, err := pm.SpawnProcess("sleep 10", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	fmt.Printf("Process started: ID=%s, PID=%d\n", process.ID, process.PID)

	// Monitor with context
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Context cancelled, stopping monitoring")
				return
			case <-ticker.C:
				if !process.IsRunning() {
					fmt.Println("Process completed")
					return
				}
				duration := process.GetDuration()
				fmt.Printf("Process running for: %v\n", duration)
			}
		}
	}()

	// Wait for context to be cancelled
	<-ctx.Done()

	// Terminate the process if still running
	if process.IsRunning() {
		fmt.Println("Terminating process due to context cancellation")
		pm.TerminateProcess(process)
	}

	fmt.Printf("Final status: %s\n", process.GetStatus())
}

// ExampleProcessManagerWithEnvironment demonstrates environment variable handling
func ExampleProcessManagerWithEnvironment() {
	pm := NewProcessManager()

	tempDir, err := os.MkdirTemp("", "ai_process_env")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a script that uses environment variables
	scriptContent := `#!/bin/bash
echo "AI Process ID: $AI_PROCESS_ID"
echo "AI Work Directory: $AI_WORK_DIR"
echo "Current working directory: $(pwd)"
echo "Custom variable: $CUSTOM_VAR"
`

	scriptPath := tempDir + "/test_script.sh"
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0o755)
	if err != nil {
		log.Fatalf("Failed to write script: %v", err)
	}

	// Set custom environment variable
	os.Setenv("CUSTOM_VAR", "custom_value")

	// Spawn process with the script
	process, err := pm.SpawnProcess("./test_script.sh", tempDir)
	if err != nil {
		log.Fatalf("Failed to spawn process: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	stdout, _ := process.GetOutput()
	fmt.Printf("Script output:\n%s", string(stdout))
}
