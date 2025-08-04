package ai

import (
	"fmt"
	"log"
	"time"
)

// ExampleExecutionService demonstrates how to use the ExecutionService
func ExampleExecutionService() {
	// Initialize services
	cliManager, err := NewCLIManager(DefaultCLIConfig())
	if err != nil {
		log.Fatalf("Failed to create CLI manager: %v", err)
	}

	processManager := NewProcessManager()
	executionService := NewExecutionService(cliManager, processManager)

	// Set up real-time update callback
	executionService.SetUpdateCallback(func(update ExecutionUpdate) {
		fmt.Printf("[%s] Execution %s: %s (%.1f%%)\n",
			update.Timestamp.Format("15:04:05"),
			update.ExecutionID[:8], // Show first 8 chars of ID
			update.Status,
			update.Progress*100)

		if update.Log != "" {
			fmt.Printf("  Log: %s\n", update.Log)
		}

		if update.Error != "" {
			fmt.Printf("  Error: %s\n", update.Error)
		}
	})

	// Create a sample plan
	plan := Plan{
		ID:          "example-plan-1",
		TaskID:      "example-task-1",
		Description: "Example AI execution plan",
		Steps: []PlanStep{
			{
				ID:          "step-1",
				Description: "Initialize AI model",
				Action:      "init_model",
				Parameters:  map[string]string{"model": "claude-3-sonnet"},
				Order:       1,
			},
			{
				ID:          "step-2",
				Description: "Process input data",
				Action:      "process_data",
				Parameters:  map[string]string{"input": "sample_data.json"},
				Order:       2,
			},
			{
				ID:          "step-3",
				Description: "Generate output",
				Action:      "generate_output",
				Parameters:  map[string]string{"format": "json"},
				Order:       3,
			},
		},
		Context: map[string]string{
			"project": "example-project",
			"user":    "demo-user",
		},
		CreatedAt: time.Now(),
	}

	// Start execution
	fmt.Println("Starting AI execution...")
	execution, err := executionService.StartExecution("example-task-1", plan)
	if err != nil {
		log.Fatalf("Failed to start execution: %v", err)
	}

	fmt.Printf("Execution started with ID: %s\n", execution.ID)

	// Monitor execution for a while
	time.Sleep(2 * time.Second)

	// Get current execution status
	currentExecution, err := executionService.GetExecution(execution.ID)
	if err != nil {
		log.Printf("Failed to get execution: %v", err)
	} else {
		fmt.Printf("Current status: %s, Progress: %.1f%%\n",
			currentExecution.Status,
			currentExecution.Progress*100)

		if len(currentExecution.Logs) > 0 {
			fmt.Println("Recent logs:")
			for _, log := range currentExecution.Logs {
				fmt.Printf("  %s\n", log)
			}
		}
	}

	// List all active executions
	executions := executionService.ListExecutions()
	fmt.Printf("Active executions: %d\n", len(executions))

	// Example of execution control (commented out to let execution complete)
	/*
		// Pause execution
		fmt.Println("Pausing execution...")
		err = executionService.PauseExecution(execution.ID)
		if err != nil {
			log.Printf("Failed to pause execution: %v", err)
		}

		time.Sleep(1 * time.Second)

		// Resume execution
		fmt.Println("Resuming execution...")
		err = executionService.ResumeExecution(execution.ID)
		if err != nil {
			log.Printf("Failed to resume execution: %v", err)
		}

		time.Sleep(1 * time.Second)

		// Cancel execution
		fmt.Println("Cancelling execution...")
		err = executionService.CancelExecution(execution.ID)
		if err != nil {
			log.Printf("Failed to cancel execution: %v", err)
		}
	*/

	// Wait for execution to complete (or timeout)
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			fmt.Println("Execution timeout reached")
			return
		case <-ticker.C:
			exec, err := executionService.GetExecution(execution.ID)
			if err != nil {
				// Execution completed and was cleaned up
				fmt.Println("Execution completed and cleaned up")
				return
			}

			if exec.Status == ExecutionStatusCompleted {
				fmt.Println("Execution completed successfully!")
				if exec.Result != nil {
					fmt.Printf("Output: %s\n", exec.Result.Output)
					fmt.Printf("Duration: %v\n", exec.Result.Duration)
				}
				return
			} else if exec.Status == ExecutionStatusFailed {
				fmt.Printf("Execution failed: %s\n", exec.Error)
				return
			} else if exec.Status == ExecutionStatusCancelled {
				fmt.Println("Execution was cancelled")
				return
			}
		}
	}
}

// ExampleExecutionServiceWithWebSocket demonstrates integration with WebSocket
func ExampleExecutionServiceWithWebSocket() {
	// This example shows how ExecutionService can be integrated with WebSocket
	// for real-time updates to frontend clients

	cliManager, err := NewCLIManager(DefaultCLIConfig())
	if err != nil {
		log.Fatalf("Failed to create CLI manager: %v", err)
	}

	processManager := NewProcessManager()
	executionService := NewExecutionService(cliManager, processManager)

	// Set up WebSocket integration
	executionService.SetUpdateCallback(func(update ExecutionUpdate) {
		// In a real implementation, this would send the update via WebSocket
		// to connected clients
		fmt.Printf("WebSocket Update: %+v\n", update)

		// Example WebSocket message structure:
		/*
			message := map[string]interface{}{
				"type": "execution_update",
				"data": update,
			}

			// Send to all connected clients
			websocketHub.Broadcast(message)
		*/
	})

	// Start multiple executions
	plans := []Plan{
		{
			ID:          "batch-plan-1",
			TaskID:      "batch-task-1",
			Description: "Batch processing task 1",
			Steps:       []PlanStep{},
			Context:     map[string]string{},
			CreatedAt:   time.Now(),
		},
		{
			ID:          "batch-plan-2",
			TaskID:      "batch-task-2",
			Description: "Batch processing task 2",
			Steps:       []PlanStep{},
			Context:     map[string]string{},
			CreatedAt:   time.Now(),
		},
	}

	executionIDs := make([]string, 0, len(plans))

	for i, plan := range plans {
		execution, err := executionService.StartExecution(fmt.Sprintf("batch-task-%d", i+1), plan)
		if err != nil {
			log.Printf("Failed to start execution %d: %v", i+1, err)
			continue
		}
		executionIDs = append(executionIDs, execution.ID)
		fmt.Printf("Started execution %d: %s\n", i+1, execution.ID)
	}

	// Monitor all executions
	time.Sleep(5 * time.Second)

	// List all active executions
	executions := executionService.ListExecutions()
	fmt.Printf("Active executions: %d\n", len(executions))

	// Cleanup completed executions
	executionService.CleanupCompletedExecutions()

	// Final count
	executions = executionService.ListExecutions()
	fmt.Printf("Remaining active executions: %d\n", len(executions))
}
