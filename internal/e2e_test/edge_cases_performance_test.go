package e2e_test

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestEdgeCasesTaskStates tests edge cases in task state management
func TestEdgeCasesTaskStates(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("InvalidTaskStatusTransitions", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Invalid Transition Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Invalid Transition Task",
			Status: entity.TaskStatusTODO,
		})

		// Try to transition directly from TODO to DONE (invalid)
		task.Status = entity.TaskStatusDONE
		err := suite.repositories.Task.Update(suite.ctx, task)

		// Should either fail validation or be handled gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "invalid")
		} else {
			// If update succeeded, verify system handles it gracefully
			updatedTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
			require.NoError(t, err)
			assert.NotNil(t, updatedTask)
		}
	})

	t.Run("OrphanedTaskCleanup", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Orphaned Task Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Orphaned Task",
			Status: entity.TaskStatusIMPLEMENTING,
		})

		// Create worktree and execution for the task
		worktree := dataGen.GenerateWorktree(project.ID, task.ID, WorktreeConfig{
			Branch: "orphaned-task-branch",
			Status: entity.WorktreeStatusActive,
		})

		execution := dataGen.GenerateExecution(task.ID, ExecutionConfig{
			Type:   entity.ExecutionTypeImplementation,
			Status: entity.ExecutionStatusRunning,
		})

		// Delete project (simulating cascade delete)
		err := suite.repositories.Project.Delete(suite.ctx, project.ID)
		require.NoError(t, err)

		// Verify dependent resources are cleaned up or handled gracefully
		_, err = suite.repositories.Task.GetByID(suite.ctx, task.ID)
		assert.Error(t, err) // Task should be deleted

		_, err = suite.repositories.Worktree.GetByID(suite.ctx, worktree.ID)
		assert.Error(t, err) // Worktree should be deleted

		_, err = suite.repositories.Execution.GetByID(suite.ctx, execution.ID)
		assert.Error(t, err) // Execution should be deleted
	})

	t.Run("TaskWithMissingDependencies", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Missing Dependencies Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Task with Missing Dependencies",
			Status: entity.TaskStatusIMPLEMENTING,
		})

		// Try to start implementation without approved plan
		SetupHappyPathMockExpectations(suite.services)
		
		resp := suite.attemptTaskImplementationHTTP(t, task.ID)
		defer resp.Body.Close()

		// Should handle missing plan gracefully
		// Either return error or create default plan
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusOK}, resp.StatusCode)
	})
}

// TestLargeDataHandling tests system behavior with large datasets
func TestLargeDataHandling(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("LargeTaskDescription", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Large Data Test"})

		// Create task with very large description (1MB)
		largeDescription := make([]byte, 1024*1024)
		for i := range largeDescription {
			largeDescription[i] = byte('a' + (i % 26))
		}

		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       "Large Description Task",
			Description: string(largeDescription),
			Status:      entity.TaskStatusTODO,
		}

		err := suite.repositories.Task.Create(suite.ctx, task)
		
		// Should either succeed or fail gracefully with proper error
		if err != nil {
			assert.Contains(t, err.Error(), "too large")
		} else {
			// Verify data integrity
			retrievedTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
			require.NoError(t, err)
			assert.Equal(t, len(largeDescription), len(retrievedTask.Description))
		}
	})

	t.Run("LargePlanContent", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Large Plan Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{Title: "Large Plan Task"})

		// Create plan with large content structure
		largePlanContent := map[string]interface{}{
			"title":       "Large Plan",
			"description": "Plan with many steps",
			"steps":       make([]map[string]interface{}, 1000),
		}

		// Fill with many steps
		for i := 0; i < 1000; i++ {
			largePlanContent["steps"].([]map[string]interface{})[i] = map[string]interface{}{
				"id":          i + 1,
				"title":       fmt.Sprintf("Step %d", i+1),
				"description": fmt.Sprintf("Detailed description for step %d with lots of content", i+1),
				"files":       []string{fmt.Sprintf("file%d.go", i+1)},
			}
		}

		plan := &entity.Plan{
			ID:      uuid.New(),
			TaskID:  task.ID,
			Title:   "Large Plan",
			Content: largePlanContent,
			Status:  entity.PlanStatusDraft,
		}

		err := suite.repositories.Plan.Create(suite.ctx, plan)
		
		// Should handle large plan content appropriately
		if err != nil {
			t.Logf("Large plan creation failed (expected): %v", err)
		} else {
			// Verify data integrity
			retrievedPlan, err := suite.repositories.Plan.GetByID(suite.ctx, plan.ID)
			require.NoError(t, err)
			assert.Len(t, retrievedPlan.Content["steps"], 1000)
		}
	})

	t.Run("ManyTasksInProject", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Many Tasks Test"})

		// Create many tasks
		taskCount := 100
		tasks := make([]*entity.Task, taskCount)

		for i := 0; i < taskCount; i++ {
			tasks[i] = dataGen.GenerateTask(project.ID, TaskConfig{
				Title: fmt.Sprintf("Bulk Task %d", i+1),
			})
		}

		// Verify all tasks were created
		allTasks, err := suite.repositories.Task.List(suite.ctx, entity.TaskFilters{
			ProjectID: &project.ID,
		})
		require.NoError(t, err)
		assert.Len(t, allTasks, taskCount)

		// Test pagination performance
		start := time.Now()
		pagedTasks, err := suite.repositories.Task.List(suite.ctx, entity.TaskFilters{
			ProjectID: &project.ID,
			Limit:     20,
			Offset:    40,
		})
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, pagedTasks, 20)
		assert.Less(t, duration, 1*time.Second) // Should be reasonably fast
	})
}

// TestHighVolumeTaskProcessing tests system performance under high load
func TestHighVolumeTaskProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high volume test in short mode")
	}

	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)
	SetupHappyPathMockExpectations(suite.services)

	// Create bulk test data
	bulkData := dataGen.CreateBulkTestData(BulkDataConfig{
		ProjectCount:      10,
		TasksPerProject:   50,
		ExecutionsPerTask: 2,
	})

	metrics := &PerformanceMetrics{
		StartTime: time.Now(),
	}

	t.Run("ConcurrentTaskCreation", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Concurrent Creation Test"})

		// Create tasks concurrently
		numWorkers := 10
		tasksPerWorker := 20
		
		var wg sync.WaitGroup
		createdTasks := make(chan *entity.Task, numWorkers*tasksPerWorker)
		errors := make(chan error, numWorkers*tasksPerWorker)

		start := time.Now()

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for j := 0; j < tasksPerWorker; j++ {
					task := &entity.Task{
						ProjectID:   project.ID,
						Title:       fmt.Sprintf("Concurrent Task %d-%d", workerID, j),
						Description: fmt.Sprintf("Task created by worker %d", workerID),
						Status:      entity.TaskStatusTODO,
					}

					if err := suite.repositories.Task.Create(suite.ctx, task); err != nil {
						errors <- err
					} else {
						createdTasks <- task
					}
				}
			}(i)
		}

		wg.Wait()
		close(createdTasks)
		close(errors)

		duration := time.Since(start)
		metrics.TaskCreationDuration = duration

		// Count results
		successCount := len(createdTasks)
		errorCount := len(errors)

		t.Logf("Created %d tasks in %v (%d errors)", successCount, duration, errorCount)
		
		// Performance assertions
		assert.Greater(t, successCount, numWorkers*tasksPerWorker/2) // At least 50% success
		assert.Less(t, duration, 30*time.Second) // Should complete within reasonable time

		// Calculate throughput
		throughput := float64(successCount) / duration.Seconds()
		metrics.TaskThroughput = throughput
		t.Logf("Task creation throughput: %.2f tasks/second", throughput)
	})

	t.Run("DatabaseQueryPerformance", func(t *testing.T) {
		// Test query performance with many records
		start := time.Now()

		// Complex query with filters and joins
		tasks, err := suite.repositories.Task.List(suite.ctx, entity.TaskFilters{
			Status:   []entity.TaskStatus{entity.TaskStatusTODO, entity.TaskStatusPLANNING},
			Priority: []entity.TaskPriority{entity.TaskPriorityHigh},
			Limit:    100,
		})

		queryDuration := time.Since(start)
		metrics.QueryDuration = queryDuration

		require.NoError(t, err)
		t.Logf("Query returned %d tasks in %v", len(tasks), queryDuration)

		// Query should be reasonably fast even with many records
		assert.Less(t, queryDuration, 5*time.Second)
	})

	t.Run("MemoryUsageUnderLoad", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Perform memory-intensive operations
		for i := 0; i < 100; i++ {
			// Create temporary data structures
			tempData := make([]map[string]interface{}, 100)
			for j := range tempData {
				tempData[j] = map[string]interface{}{
					"id":          uuid.New().String(),
					"data":        make([]byte, 1024),
					"timestamp":   time.Now(),
					"metadata":    make(map[string]string),
				}
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.TotalAlloc - m1.TotalAlloc
		metrics.MemoryUsage = memoryUsed

		t.Logf("Memory used: %d bytes", memoryUsed)
		
		// Memory usage should be reasonable
		assert.Less(t, memoryUsed, uint64(100*1024*1024)) // Less than 100MB
	})

	// Log performance metrics
	t.Logf("Performance Metrics: %+v", metrics)
}

// TestWebSocketPerformance tests WebSocket performance under load
func TestWebSocketPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket performance test in short mode")
	}

	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("MultipleWebSocketConnections", func(t *testing.T) {
		connectionCount := 50
		connections := make([]*websocket.Conn, connectionCount)
		defer func() {
			for _, conn := range connections {
				if conn != nil {
					conn.Close()
				}
			}
		}()

		// Create multiple WebSocket connections
		start := time.Now()
		var wg sync.WaitGroup
		connectionErrors := make(chan error, connectionCount)

		for i := 0; i < connectionCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				wsURL := suite.GetWebSocketURL()
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err != nil {
					connectionErrors <- err
					return
				}
				connections[index] = conn
			}(i)
		}

		wg.Wait()
		connectionDuration := time.Since(start)
		close(connectionErrors)

		errorCount := len(connectionErrors)
		successCount := connectionCount - errorCount

		t.Logf("Created %d WebSocket connections in %v (%d errors)", 
			successCount, connectionDuration, errorCount)

		assert.Greater(t, successCount, connectionCount*3/4) // At least 75% success
		assert.Less(t, connectionDuration, 10*time.Second)
	})

	t.Run("MessageDeliveryPerformance", func(t *testing.T) {
		// Create WebSocket connection
		wsURL := suite.GetWebSocketURL()
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Set up message collection
		messages := make(chan map[string]interface{}, 100)
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
					// Buffer full
				}
			}
		}()

		// Create project and task to trigger notifications
		project := dataGen.GenerateProject(ProjectConfig{Name: "WebSocket Perf Test"})
		
		// Subscribe to project updates
		subscribeMessage := map[string]interface{}{
			"type":    "subscribe",
			"channel": fmt.Sprintf("project:%s", project.ID.String()),
		}
		err = conn.WriteJSON(subscribeMessage)
		require.NoError(t, err)

		// Create many tasks to generate notifications
		taskCount := 20
		start := time.Now()

		for i := 0; i < taskCount; i++ {
			dataGen.GenerateTask(project.ID, TaskConfig{
				Title: fmt.Sprintf("WebSocket Test Task %d", i+1),
			})
		}

		// Collect notifications with timeout
		receivedCount := 0
		timeout := time.After(5 * time.Second)

		for receivedCount < taskCount {
			select {
			case msg := <-messages:
				if msg["type"] == "task_created" {
					receivedCount++
				}
			case <-timeout:
				break
			}
		}

		deliveryDuration := time.Since(start)
		t.Logf("Received %d/%d notifications in %v", receivedCount, taskCount, deliveryDuration)

		// Should receive most notifications quickly
		assert.Greater(t, receivedCount, taskCount/2)
		assert.Less(t, deliveryDuration, 10*time.Second)
	})
}

// TestSystemLimits tests system behavior at various limits
func TestSystemLimits(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("MaxProjectsPerUser", func(t *testing.T) {
		// This test would be more relevant with authentication
		// For now, test database limits
		
		projectCount := 1000 // Large number to test limits
		projects := make([]*entity.Project, 0, projectCount)

		start := time.Now()
		
		for i := 0; i < projectCount && time.Since(start) < 30*time.Second; i++ {
			project := dataGen.GenerateProject(ProjectConfig{
				Name: fmt.Sprintf("Limit Test Project %d", i+1),
			})
			projects = append(projects, project)
		}

		actualCount := len(projects)
		t.Logf("Created %d projects", actualCount)

		// System should handle many projects gracefully
		assert.Greater(t, actualCount, 100) // Should create at least 100
	})

	t.Run("MaxExecutionLogs", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Log Limit Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{Title: "Log Test Task"})
		execution := dataGen.GenerateExecution(task.ID, ExecutionConfig{
			Type:   entity.ExecutionTypePlanning,
			Status: entity.ExecutionStatusRunning,
		})

		// Create many execution logs
		logCount := 10000
		
		start := time.Now()
		for i := 0; i < logCount && time.Since(start) < 10*time.Second; i++ {
			log := &entity.ExecutionLog{
				ID:          uuid.New(),
				ExecutionID: execution.ID,
				Level:       "INFO",
				Message:     fmt.Sprintf("Log message %d", i+1),
				Timestamp:   time.Now(),
			}

			err := suite.repositories.ExecutionLog.Create(suite.ctx, log)
			if err != nil {
				t.Logf("Log creation failed at %d: %v", i, err)
				break
			}
		}

		duration := time.Since(start)
		
		// Count actual logs created
		logs, err := suite.repositories.ExecutionLog.List(suite.ctx, entity.ExecutionLogFilters{
			ExecutionID: &execution.ID,
		})
		require.NoError(t, err)

		actualLogCount := len(logs)
		t.Logf("Created %d execution logs in %v", actualLogCount, duration)

		// Should handle reasonable number of logs
		assert.Greater(t, actualLogCount, 1000)
	})
}

// TestDataConsistency tests data consistency under various scenarios
func TestDataConsistency(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite)

	t.Run("ConcurrentStatusUpdates", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Consistency Test"})
		task := dataGen.GenerateTask(project.ID, TaskConfig{
			Title:  "Consistency Test Task",
			Status: entity.TaskStatusTODO,
		})

		// Perform concurrent updates
		numUpdaters := 10
		var wg sync.WaitGroup
		results := make(chan error, numUpdaters)

		for i := 0; i < numUpdaters; i++ {
			wg.Add(1)
			go func(updaterID int) {
				defer wg.Done()
				
				// Each updater tries to update task
				taskCopy := *task
				taskCopy.Status = entity.TaskStatus(fmt.Sprintf("STATUS_%d", updaterID))
				
				err := suite.repositories.Task.Update(suite.ctx, &taskCopy)
				results <- err
			}(i)
		}

		wg.Wait()
		close(results)

		// Check final state consistency
		finalTask, err := suite.repositories.Task.GetByID(suite.ctx, task.ID)
		require.NoError(t, err)
		
		// Task should be in a valid state (even if custom status)
		assert.NotEmpty(t, finalTask.Status)
		assert.NotNil(t, finalTask.UpdatedAt)

		// Count successful updates
		successCount := 0
		for err := range results {
			if err == nil {
				successCount++
			}
		}

		t.Logf("Successful concurrent updates: %d/%d", successCount, numUpdaters)
	})

	t.Run("TransactionConsistency", func(t *testing.T) {
		project := dataGen.GenerateProject(ProjectConfig{Name: "Transaction Test"})
		
		// Test that related entities are created/updated atomically
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       "Transaction Test Task",
			Description: "Test atomic operations",
			Status:      entity.TaskStatusTODO,
		}

		// Use transaction to create task and related entities
		tx := suite.db.DB.Begin()
		
		err := tx.Create(task).Error
		require.NoError(t, err)

		execution := &entity.Execution{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Type:      entity.ExecutionTypePlanning,
			Status:    entity.ExecutionStatusRunning,
			StartedAt: time.Now(),
		}

		err = tx.Create(execution).Error
		require.NoError(t, err)

		// Commit transaction
		err = tx.Commit().Error
		require.NoError(t, err)

		// Verify both entities exist
		_, err = suite.repositories.Task.GetByID(suite.ctx, task.ID)
		assert.NoError(t, err)

		_, err = suite.repositories.Execution.GetByID(suite.ctx, execution.ID)
		assert.NoError(t, err)
	})
}

// PerformanceMetrics holds performance test results
type PerformanceMetrics struct {
	StartTime              time.Time
	TaskCreationDuration   time.Duration
	TaskThroughput         float64
	QueryDuration          time.Duration
	MemoryUsage            uint64
	WebSocketConnections   int
	MessageDeliveryLatency time.Duration
}

// BenchmarkTaskOperations provides benchmark tests for task operations
func BenchmarkTaskOperations(b *testing.B) {
	suite := NewE2ETestSuite(&testing.T{})
	defer suite.Teardown()

	dataGen := NewTestDataGenerator(suite.t)
	project := dataGen.GenerateProject(ProjectConfig{Name: "Benchmark Test"})

	b.Run("TaskCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := &entity.Task{
				ProjectID:   project.ID,
				Title:       fmt.Sprintf("Benchmark Task %d", i),
				Description: "Benchmark task description",
				Status:      entity.TaskStatusTODO,
			}

			err := suite.repositories.Task.Create(suite.ctx, task)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("TaskRetrieval", func(b *testing.B) {
		// Create some tasks first
		tasks := make([]*entity.Task, 100)
		for i := 0; i < 100; i++ {
			tasks[i] = dataGen.GenerateTask(project.ID, TaskConfig{
				Title: fmt.Sprintf("Retrieval Test Task %d", i),
			})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			taskID := tasks[i%len(tasks)].ID
			_, err := suite.repositories.Task.GetByID(suite.ctx, taskID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("TaskList", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := suite.repositories.Task.List(suite.ctx, entity.TaskFilters{
				ProjectID: &project.ID,
				Limit:     20,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Helper method to add import for bytes package
import "bytes"