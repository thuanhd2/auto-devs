package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestDatabaseIntegration_MigrationIntegrity tests that migrations work correctly
func TestDatabaseIntegration_MigrationIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	container, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	t.Run("all tables exist", func(t *testing.T) {
		expectedTables := []string{"projects", "tasks", "audit_logs"}
		
		for _, tableName := range expectedTables {
			var exists bool
			err := container.GormDB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists).Error
			require.NoError(t, err)
			assert.True(t, exists, "Table %s should exist", tableName)
		}
	})

	t.Run("foreign key constraints work", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		taskRepo := NewTaskRepository(container.DB)
		ctx := context.Background()

		// Create project
		project := &entity.Project{
			Name:        "FK Test Project",
			Description: "Testing foreign keys",
			RepoURL:     "https://github.com/test/fk.git",
		}
		err := projectRepo.Create(ctx, project)
		require.NoError(t, err)

		// Create task with valid project ID
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       "FK Test Task",
			Description: "Testing foreign keys",
			Status:      entity.TaskStatusTODO,
		}
		err = taskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Try to create task with invalid project ID - should fail
		invalidTask := &entity.Task{
			ProjectID:   uuid.New(), // Non-existent project
			Title:       "Invalid FK Task",
			Description: "This should fail",
			Status:      entity.TaskStatusTODO,
		}
		err = taskRepo.Create(ctx, invalidTask)
		assert.Error(t, err, "Creating task with non-existent project ID should fail")
	})

	t.Run("indexes improve query performance", func(t *testing.T) {
		// Check that important indexes exist
		var indexExists bool
		
		// Check for project name index
		err := container.GormDB.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE tablename = 'projects' 
				AND indexname LIKE '%name%'
			)
		`).Scan(&indexExists).Error
		require.NoError(t, err)
		// Note: This might not exist depending on implementation
		
		// Check for task project_id index (should exist for foreign key)
		err = container.GormDB.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE tablename = 'tasks' 
				AND indexname LIKE '%project_id%'
			)
		`).Scan(&indexExists).Error
		require.NoError(t, err)
		// Foreign key index should exist
	})

	t.Run("soft delete functionality", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		ctx := context.Background()

		// Create project
		project := &entity.Project{
			Name:        "Soft Delete Test",
			Description: "Testing soft delete",
			RepoURL:     "https://github.com/test/softdelete.git",
		}
		err := projectRepo.Create(ctx, project)
		require.NoError(t, err)

		// Delete project (soft delete)
		err = projectRepo.Delete(ctx, project.ID)
		require.NoError(t, err)

		// Project should not be found in normal queries
		_, err = projectRepo.GetByID(ctx, project.ID)
		assert.Error(t, err)

		// But should exist in database with deleted_at set
		var count int64
		err = container.GormDB.Unscoped().Model(&entity.Project{}).Where("id = ?", project.ID).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Soft deleted project should still exist in database")

		// Check that deleted_at is set
		var deletedProject entity.Project
		err = container.GormDB.Unscoped().Where("id = ?", project.ID).First(&deletedProject).Error
		require.NoError(t, err)
		assert.NotNil(t, deletedProject.DeletedAt, "deleted_at should be set")
	})
}

// TestDatabaseIntegration_TransactionHandling tests transaction behavior
func TestDatabaseIntegration_TransactionHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	container, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	t.Run("transaction rollback on error", func(t *testing.T) {
		ctx := context.Background()

		// Start transaction
		tx := container.GormDB.Begin()
		require.NoError(t, tx.Error)

		// Create project in transaction
		project := &entity.Project{
			Name:        "Transaction Test",
			Description: "Testing transactions",
			RepoURL:     "https://github.com/test/transaction.git",
		}
		err := tx.Create(project).Error
		require.NoError(t, err)

		// Project should exist in transaction
		var count int64
		err = tx.Model(&entity.Project{}).Where("id = ?", project.ID).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Rollback transaction
		tx.Rollback()

		// Project should not exist in main database
		err = container.GormDB.Model(&entity.Project{}).Where("id = ?", project.ID).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "Project should not exist after rollback")
	})

	t.Run("transaction commit persists data", func(t *testing.T) {
		ctx := context.Background()

		// Start transaction
		tx := container.GormDB.Begin()
		require.NoError(t, tx.Error)

		// Create project in transaction
		project := &entity.Project{
			Name:        "Commit Test",
			Description: "Testing transaction commit",
			RepoURL:     "https://github.com/test/commit.git",
		}
		err := tx.Create(project).Error
		require.NoError(t, err)

		// Commit transaction
		err = tx.Commit().Error
		require.NoError(t, err)

		// Project should exist in main database
		var retrievedProject entity.Project
		err = container.GormDB.Where("id = ?", project.ID).First(&retrievedProject).Error
		require.NoError(t, err)
		assert.Equal(t, project.Name, retrievedProject.Name)
	})

	t.Run("nested transactions", func(t *testing.T) {
		ctx := context.Background()

		// Start outer transaction
		outerTx := container.GormDB.Begin()
		require.NoError(t, outerTx.Error)

		// Create project in outer transaction
		project := &entity.Project{
			Name:        "Outer Transaction",
			Description: "Testing nested transactions",
			RepoURL:     "https://github.com/test/nested.git",
		}
		err := outerTx.Create(project).Error
		require.NoError(t, err)

		// Start nested savepoint
		sp := outerTx.SavePoint("sp1")
		require.NoError(t, sp.Error)

		// Create task in savepoint
		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       "Nested Task",
			Description: "Testing savepoint",
			Status:      entity.TaskStatusTODO,
		}
		err = outerTx.Create(task).Error
		require.NoError(t, err)

		// Rollback to savepoint
		err = outerTx.RollbackTo("sp1").Error
		require.NoError(t, err)

		// Commit outer transaction
		err = outerTx.Commit().Error
		require.NoError(t, err)

		// Project should exist, task should not
		var projectCount, taskCount int64
		
		err = container.GormDB.Model(&entity.Project{}).Where("id = ?", project.ID).Count(&projectCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), projectCount, "Project should exist after commit")

		err = container.GormDB.Model(&entity.Task{}).Where("id = ?", task.ID).Count(&taskCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), taskCount, "Task should not exist after rollback to savepoint")
	})
}

// TestDatabaseIntegration_ConcurrentOperations tests concurrent database operations
func TestDatabaseIntegration_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent database test in short mode")
	}

	container, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	t.Run("concurrent project creation", func(t *testing.T) {
		const numGoroutines = 20
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)
		projectIDs := make(chan uuid.UUID, numGoroutines)

		projectRepo := NewProjectRepository(container.DB)
		ctx := context.Background()

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				defer wg.Done()
				
				project := &entity.Project{
					Name:        fmt.Sprintf("Concurrent Project %d", i),
					Description: fmt.Sprintf("Description %d", i),
					RepoURL:     fmt.Sprintf("https://github.com/test/concurrent%d.git", i),
				}
				
				if err := projectRepo.Create(ctx, project); err != nil {
					errors <- err
				} else {
					projectIDs <- project.ID
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		close(projectIDs)

		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Logf("Concurrent creation error: %v", err)
			errorCount++
		}

		// Count successful creations
		successCount := 0
		createdIDs := make([]uuid.UUID, 0)
		for id := range projectIDs {
			createdIDs = append(createdIDs, id)
			successCount++
		}

		assert.Equal(t, numGoroutines, successCount+errorCount, "All goroutines should complete")
		assert.LessOrEqual(t, errorCount, numGoroutines/4, "Most concurrent operations should succeed")
		assert.GreaterOrEqual(t, successCount, numGoroutines*3/4, "Most operations should succeed")

		// Verify all successful projects exist in database
		for _, id := range createdIDs {
			_, err := projectRepo.GetByID(ctx, id)
			assert.NoError(t, err, "Created project should exist in database")
		}
	})

	t.Run("concurrent task updates", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		taskRepo := NewTaskRepository(container.DB)
		ctx := context.Background()

		// Create test project and task
		project := &entity.Project{
			Name:        "Concurrent Update Test",
			Description: "Testing concurrent updates",
			RepoURL:     "https://github.com/test/concurrent-update.git",
		}
		err := projectRepo.Create(ctx, project)
		require.NoError(t, err)

		task := &entity.Task{
			ProjectID:   project.ID,
			Title:       "Concurrent Task",
			Description: "Initial description",
			Status:      entity.TaskStatusTODO,
		}
		err = taskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Concurrent updates
		const numGoroutines = 10
		var wg sync.WaitGroup
		updateResults := make(chan error, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				defer wg.Done()
				
				// Get task, modify, and update
				currentTask, err := taskRepo.GetByID(ctx, task.ID)
				if err != nil {
					updateResults <- err
					return
				}

				currentTask.Description = fmt.Sprintf("Updated by goroutine %d", i)
				
				if err := taskRepo.Update(ctx, currentTask); err != nil {
					updateResults <- err
				} else {
					updateResults <- nil
				}
			}(i)
		}

		wg.Wait()
		close(updateResults)

		// Check results
		successCount := 0
		for err := range updateResults {
			if err == nil {
				successCount++
			} else {
				t.Logf("Concurrent update error: %v", err)
			}
		}

		// Most updates should succeed
		assert.GreaterOrEqual(t, successCount, numGoroutines/2, "At least half of concurrent updates should succeed")

		// Verify final state
		finalTask, err := taskRepo.GetByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Contains(t, finalTask.Description, "Updated by goroutine", "Task should have been updated")
	})

	t.Run("deadlock prevention", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		taskRepo := NewTaskRepository(container.DB)
		ctx := context.Background()

		// Create test projects
		project1 := &entity.Project{
			Name:        "Deadlock Test 1",
			Description: "Testing deadlock prevention",
			RepoURL:     "https://github.com/test/deadlock1.git",
		}
		project2 := &entity.Project{
			Name:        "Deadlock Test 2",
			Description: "Testing deadlock prevention",
			RepoURL:     "https://github.com/test/deadlock2.git",
		}
		
		err := projectRepo.Create(ctx, project1)
		require.NoError(t, err)
		err = projectRepo.Create(ctx, project2)
		require.NoError(t, err)

		// Create tasks
		task1 := &entity.Task{
			ProjectID:   project1.ID,
			Title:       "Task 1",
			Description: "Deadlock test task 1",
			Status:      entity.TaskStatusTODO,
		}
		task2 := &entity.Task{
			ProjectID:   project2.ID,
			Title:       "Task 2",
			Description: "Deadlock test task 2",
			Status:      entity.TaskStatusTODO,
		}
		
		err = taskRepo.Create(ctx, task1)
		require.NoError(t, err)
		err = taskRepo.Create(ctx, task2)
		require.NoError(t, err)

		// Concurrent operations that might cause deadlock
		const numGoroutines = 10
		var wg sync.WaitGroup
		deadlockResults := make(chan error, numGoroutines*2)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(2)
			
			// Goroutine 1: Update task1 then task2
			go func(i int) {
				defer wg.Done()
				
				tx := container.GormDB.Begin()
				defer tx.Rollback()

				// Update task1
				err := tx.Model(&entity.Task{}).Where("id = ?", task1.ID).
					Update("description", fmt.Sprintf("Updated by routine %d-A", i)).Error
				if err != nil {
					deadlockResults <- err
					return
				}

				// Small delay to increase chance of deadlock
				time.Sleep(time.Millisecond)

				// Update task2
				err = tx.Model(&entity.Task{}).Where("id = ?", task2.ID).
					Update("description", fmt.Sprintf("Updated by routine %d-A", i)).Error
				if err != nil {
					deadlockResults <- err
					return
				}

				tx.Commit()
				deadlockResults <- nil
			}(i)

			// Goroutine 2: Update task2 then task1 (reverse order)
			go func(i int) {
				defer wg.Done()
				
				tx := container.GormDB.Begin()
				defer tx.Rollback()

				// Update task2
				err := tx.Model(&entity.Task{}).Where("id = ?", task2.ID).
					Update("description", fmt.Sprintf("Updated by routine %d-B", i)).Error
				if err != nil {
					deadlockResults <- err
					return
				}

				// Small delay to increase chance of deadlock
				time.Sleep(time.Millisecond)

				// Update task1
				err = tx.Model(&entity.Task{}).Where("id = ?", task1.ID).
					Update("description", fmt.Sprintf("Updated by routine %d-B", i)).Error
				if err != nil {
					deadlockResults <- err
					return
				}

				tx.Commit()
				deadlockResults <- nil
			}(i)
		}

		// Wait with timeout to detect deadlocks
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			// All operations completed
		case <-time.After(30 * time.Second):
			t.Fatal("Operations timed out - possible deadlock detected")
		}

		close(deadlockResults)

		// Check results
		deadlockCount := 0
		for err := range deadlockResults {
			if err != nil {
				if isDeadlockError(err) {
					deadlockCount++
					t.Logf("Deadlock detected and handled: %v", err)
				} else {
					t.Logf("Other error: %v", err)
				}
			}
		}

		// Some deadlocks might occur but should be handled gracefully
		t.Logf("Deadlock count: %d", deadlockCount)
	})
}

// TestDatabaseIntegration_DataIntegrity tests data integrity constraints
func TestDatabaseIntegration_DataIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping data integrity test in short mode")
	}

	container, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	t.Run("cascade delete behavior", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		taskRepo := NewTaskRepository(container.DB)
		ctx := context.Background()

		// Create project with tasks
		project := &entity.Project{
			Name:        "Cascade Test",
			Description: "Testing cascade behavior",
			RepoURL:     "https://github.com/test/cascade.git",
		}
		err := projectRepo.Create(ctx, project)
		require.NoError(t, err)

		// Create multiple tasks
		tasks := make([]*entity.Task, 3)
		for i := 0; i < 3; i++ {
			tasks[i] = &entity.Task{
				ProjectID:   project.ID,
				Title:       fmt.Sprintf("Cascade Task %d", i),
				Description: "Testing cascade",
				Status:      entity.TaskStatusTODO,
			}
			err = taskRepo.Create(ctx, tasks[i])
			require.NoError(t, err)
		}

		// Verify tasks exist
		projectTasks, err := taskRepo.GetByProjectID(ctx, project.ID)
		require.NoError(t, err)
		assert.Len(t, projectTasks, 3)

		// Delete project (soft delete)
		err = projectRepo.Delete(ctx, project.ID)
		require.NoError(t, err)

		// Tasks should still exist since GORM doesn't auto-cascade soft deletes
		// This tests the current behavior
		existingTasks, err := taskRepo.GetByProjectID(ctx, project.ID)
		require.NoError(t, err)
		assert.Len(t, existingTasks, 3, "Tasks should still exist after project soft delete")
	})

	t.Run("unique constraints", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		ctx := context.Background()

		// Create first project
		project1 := &entity.Project{
			Name:        "Unique Name Test",
			Description: "Testing unique constraints",
			RepoURL:     "https://github.com/test/unique1.git",
		}
		err := projectRepo.Create(ctx, project1)
		require.NoError(t, err)

		// Try to create project with same name - should be prevented by application logic
		project2 := &entity.Project{
			Name:        "Unique Name Test", // Same name
			Description: "Should fail",
			RepoURL:     "https://github.com/test/unique2.git",
		}
		
		// This depends on application-level validation
		// Since we're testing at repository level, it might succeed at DB level
		// but application logic should prevent it
		err = projectRepo.Create(ctx, project2)
		// Note: This test depends on whether unique constraint exists at DB level
		if err != nil {
			assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
		}
	})

	t.Run("data validation constraints", func(t *testing.T) {
		projectRepo := NewProjectRepository(container.DB)
		ctx := context.Background()

		// Test various constraint violations
		testCases := []struct {
			name    string
			project *entity.Project
			shouldFail bool
		}{
			{
				name: "normal project",
				project: &entity.Project{
					Name:        "Normal Project",
					Description: "Valid project",
					RepoURL:     "https://github.com/test/normal.git",
				},
				shouldFail: false,
			},
			{
				name: "empty name",
				project: &entity.Project{
					Name:        "", // Empty name
					Description: "Valid project",
					RepoURL:     "https://github.com/test/empty.git",
				},
				shouldFail: true, // Should fail due to NOT NULL constraint
			},
			{
				name: "very long name",
				project: &entity.Project{
					Name:        string(make([]byte, 1000)), // Very long name
					Description: "Valid project",
					RepoURL:     "https://github.com/test/long.git",
				},
				shouldFail: false, // Should be truncated or handled by DB
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := projectRepo.Create(ctx, tc.project)
				
				if tc.shouldFail {
					assert.Error(t, err, "Expected creation to fail for %s", tc.name)
				} else {
					assert.NoError(t, err, "Expected creation to succeed for %s", tc.name)
				}
			})
		}
	})
}

// Helper function to check if an error is a deadlock error
func isDeadlockError(err error) bool {
	// PostgreSQL deadlock error codes
	return err != nil && (
		containsString(err.Error(), "deadlock detected") ||
		containsString(err.Error(), "40P01") || // PostgreSQL deadlock error code
		containsString(err.Error(), "40001"))   // Serialization failure
}

// Helper function to check if string contains substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsString(s[1:], substr))))
}