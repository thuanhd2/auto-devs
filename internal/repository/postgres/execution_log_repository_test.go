package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test execution
func CreateTestExecution(t *testing.T, ctx context.Context, db *database.GormDB) *entity.Execution {
	// Create test project first
	projectRepo := NewProjectRepository(db)
	project := CreateTestProject(t, projectRepo, ctx)
	
	// Create test task
	taskRepo := NewTaskRepository(db)
	task := CreateTestTask(t, taskRepo, project.ID, ctx)
	
	// Create execution
	execution := &entity.Execution{
		TaskID:    task.ID,
		Status:    entity.ExecutionStatusRunning,
		StartedAt: time.Now(),
		Progress:  0.0,
	}
	
	err := db.WithContext(ctx).Create(execution).Error
	require.NoError(t, err)
	
	return execution
}

// Helper function to create test execution logs
func CreateTestExecutionLogs(executionID uuid.UUID, count int) []*entity.ExecutionLog {
	logs := make([]*entity.ExecutionLog, count)
	for i := 0; i < count; i++ {
		logs[i] = &entity.ExecutionLog{
			ID:          uuid.New(),
			ExecutionID: executionID,
			Level:       entity.LogLevelInfo,
			Message:     fmt.Sprintf("Test log message %d", i+1),
			Timestamp:   time.Now().Add(time.Duration(i) * time.Second),
			Source:      "stdout",
			Line:        i + 1,
		}
	}
	return logs
}

func TestExecutionLogRepository_BatchInsertOrUpdate_Insert(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create test logs
	logs := CreateTestExecutionLogs(execution.ID, 3)

	// Test batch insert
	err := repo.BatchInsertOrUpdate(ctx, logs)
	require.NoError(t, err)

	// Verify logs were inserted
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 3)

	// Verify log contents
	for i, log := range retrievedLogs {
		assert.Equal(t, logs[i].ExecutionID, log.ExecutionID)
		assert.Equal(t, logs[i].Message, log.Message)
		assert.Equal(t, logs[i].Level, log.Level)
		assert.Equal(t, logs[i].Line, log.Line)
	}
}

func TestExecutionLogRepository_BatchInsertOrUpdate_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create and insert initial logs
	initialLogs := CreateTestExecutionLogs(execution.ID, 2)
	err := repo.BatchInsertOrUpdate(ctx, initialLogs)
	require.NoError(t, err)

	// Verify initial insertion
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 2)

	// Create updated logs with same execution_id and line but different messages
	updatedLogs := []*entity.ExecutionLog{
		{
			ID:          uuid.New(), // Different ID but same execution_id and line
			ExecutionID: execution.ID,
			Level:       entity.LogLevelWarn,
			Message:     "Updated message for line 1",
			Timestamp:   time.Now(),
			Source:      "stderr",
			Line:        1, // Same line as first initial log
		},
		{
			ID:          uuid.New(), // Different ID but same execution_id and line
			ExecutionID: execution.ID,
			Level:       entity.LogLevelError,
			Message:     "Updated message for line 2",
			Timestamp:   time.Now(),
			Source:      "stderr",
			Line:        2, // Same line as second initial log
		},
	}

	// Test batch update
	err = repo.BatchInsertOrUpdate(ctx, updatedLogs)
	require.NoError(t, err)

	// Verify logs were updated, not duplicated
	retrievedLogs, err = repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 2) // Still only 2 logs

	// Verify log contents were updated
	for _, log := range retrievedLogs {
		switch log.Line {
		case 1:
			assert.Equal(t, "Updated message for line 1", log.Message)
			assert.Equal(t, entity.LogLevelWarn, log.Level)
			assert.Equal(t, "stderr", log.Source)
		case 2:
			assert.Equal(t, "Updated message for line 2", log.Message)
			assert.Equal(t, entity.LogLevelError, log.Level)
			assert.Equal(t, "stderr", log.Source)
		}
	}
}

func TestExecutionLogRepository_BatchInsertOrUpdate_Mixed(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create and insert initial log
	initialLogs := CreateTestExecutionLogs(execution.ID, 1)
	err := repo.BatchInsertOrUpdate(ctx, initialLogs)
	require.NoError(t, err)

	// Create mixed batch: one update (same line) and one insert (new line)
	mixedLogs := []*entity.ExecutionLog{
		{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			Level:       entity.LogLevelError,
			Message:     "Updated message for existing line",
			Timestamp:   time.Now(),
			Source:      "stderr",
			Line:        1, // Update existing line 1
		},
		{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			Level:       entity.LogLevelInfo,
			Message:     "New message for line 2",
			Timestamp:   time.Now(),
			Source:      "stdout",
			Line:        2, // New line
		},
	}

	// Test batch insert/update
	err = repo.BatchInsertOrUpdate(ctx, mixedLogs)
	require.NoError(t, err)

	// Verify we have 2 logs total
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 2)

	// Verify contents
	logsByLine := make(map[int]*entity.ExecutionLog)
	for _, log := range retrievedLogs {
		logsByLine[log.Line] = log
	}

	// Line 1 should be updated
	assert.Equal(t, "Updated message for existing line", logsByLine[1].Message)
	assert.Equal(t, entity.LogLevelError, logsByLine[1].Level)

	// Line 2 should be new
	assert.Equal(t, "New message for line 2", logsByLine[2].Message)
	assert.Equal(t, entity.LogLevelInfo, logsByLine[2].Level)
}

func TestExecutionLogRepository_BatchInsertOrUpdate_EmptySlice(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Test with empty slice
	err := repo.BatchInsertOrUpdate(ctx, []*entity.ExecutionLog{})
	assert.NoError(t, err) // Should not error with empty slice
}

func TestExecutionLogRepository_BatchInsertOrUpdate_InvalidExecution(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create logs with non-existent execution ID
	logs := []*entity.ExecutionLog{
		{
			ID:          uuid.New(),
			ExecutionID: uuid.New(), // Non-existent execution
			Level:       entity.LogLevelInfo,
			Message:     "Test message",
			Timestamp:   time.Now(),
			Source:      "stdout",
			Line:        1,
		},
	}

	// Should fail due to foreign key constraint
	err := repo.BatchInsertOrUpdate(ctx, logs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert/update log")
}

func TestExecutionLogRepository_BatchInsertOrUpdateAsync_Insert(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create test logs
	logs := CreateTestExecutionLogs(execution.ID, 3)

	// Test async batch insert
	err := repo.BatchInsertOrUpdateAsync(ctx, logs)
	require.NoError(t, err)

	// Wait for async processing to complete
	time.Sleep(100 * time.Millisecond)

	// Verify logs were inserted
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 3)

	// Verify log contents
	for i, log := range retrievedLogs {
		assert.Equal(t, logs[i].ExecutionID, log.ExecutionID)
		assert.Equal(t, logs[i].Message, log.Message)
		assert.Equal(t, logs[i].Level, log.Level)
		assert.Equal(t, logs[i].Line, log.Line)
	}
}

func TestExecutionLogRepository_BatchInsertOrUpdateAsync_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create and insert initial logs
	initialLogs := CreateTestExecutionLogs(execution.ID, 2)
	err := repo.BatchInsertOrUpdateAsync(ctx, initialLogs)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify initial insertion
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 2)

	// Create updated logs with same execution_id and line
	updatedLogs := []*entity.ExecutionLog{
		{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			Level:       entity.LogLevelWarn,
			Message:     "Updated async message for line 1",
			Timestamp:   time.Now(),
			Source:      "stderr",
			Line:        1,
		},
	}

	// Test async batch update
	err = repo.BatchInsertOrUpdateAsync(ctx, updatedLogs)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify logs were updated
	retrievedLogs, err = repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 2) // Still only 2 logs

	// Find the updated log
	var updatedLog *entity.ExecutionLog
	for _, log := range retrievedLogs {
		if log.Line == 1 {
			updatedLog = log
			break
		}
	}

	require.NotNil(t, updatedLog)
	assert.Equal(t, "Updated async message for line 1", updatedLog.Message)
	assert.Equal(t, entity.LogLevelWarn, updatedLog.Level)
}

func TestExecutionLogRepository_BatchInsertOrUpdateAsync_EmptySlice(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Test with empty slice
	err := repo.BatchInsertOrUpdateAsync(ctx, []*entity.ExecutionLog{})
	assert.NoError(t, err) // Should not error with empty slice
}

func TestExecutionLogRepository_BatchInsertOrUpdateAsync_QueueFullFallback(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create many logs to potentially fill the async buffer
	logs := CreateTestExecutionLogs(execution.ID, 50)

	// Test async batch insert - should either queue or fallback to synchronous
	err := repo.BatchInsertOrUpdateAsync(ctx, logs)
	assert.NoError(t, err) // Should not error even if queue is full (fallback to sync)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify logs were processed (either async or sync fallback)
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 50)
}

func TestExecutionLogRepository_BatchInsertOrUpdateAsync_ContextCancellation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)

	// Create test execution
	execution := CreateTestExecution(t, context.Background(), db)

	// Create test logs
	logs := CreateTestExecutionLogs(execution.ID, 3)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test with cancelled context
	err := repo.BatchInsertOrUpdateAsync(ctx, logs)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestExecutionLogRepository_BatchInsertOrUpdate_DefaultValues(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create logs without ID and timestamp
	logs := []*entity.ExecutionLog{
		{
			// ID not set - should be generated
			ExecutionID: execution.ID,
			Level:       entity.LogLevelInfo,
			Message:     "Test message without ID",
			// Timestamp not set - should be set to current time
			Source: "stdout",
			Line:   1,
		},
	}

	// Test batch insert
	err := repo.BatchInsertOrUpdate(ctx, logs)
	require.NoError(t, err)

	// Verify log was inserted with default values
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 1)

	log := retrievedLogs[0]
	assert.NotEqual(t, uuid.Nil, log.ID) // ID should be generated
	assert.False(t, log.Timestamp.IsZero()) // Timestamp should be set
}

func TestExecutionLogRepository_BatchInsertOrUpdateAsync_DefaultValues(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewExecutionLogRepository(db)
	ctx := context.Background()

	// Create test execution
	execution := CreateTestExecution(t, ctx, db)

	// Create logs without ID and timestamp
	logs := []*entity.ExecutionLog{
		{
			// ID not set - should be generated
			ExecutionID: execution.ID,
			Level:       entity.LogLevelInfo,
			Message:     "Test async message without ID",
			// Timestamp not set - should be set to current time
			Source: "stdout",
			Line:   1,
		},
	}

	// Test async batch insert
	err := repo.BatchInsertOrUpdateAsync(ctx, logs)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify log was inserted with default values
	retrievedLogs, err := repo.GetByExecutionID(ctx, execution.ID)
	require.NoError(t, err)
	assert.Len(t, retrievedLogs, 1)

	log := retrievedLogs[0]
	assert.NotEqual(t, uuid.Nil, log.ID) // ID should be generated
	assert.False(t, log.Timestamp.IsZero()) // Timestamp should be set
}