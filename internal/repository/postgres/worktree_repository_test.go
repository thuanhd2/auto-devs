package postgres

import (
	"context"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestWorktree(t *testing.T, worktreeRepo repository.WorktreeRepository, taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, ctx context.Context) *entity.Worktree {
	// Create test project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create test task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err = taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Create test worktree
	worktree := &entity.Worktree{
		TaskID:       task.ID,
		ProjectID:    project.ID,
		BranchName:   "test-branch",
		WorktreePath: "/tmp/test-worktree",
		Status:       entity.WorktreeStatusCreating,
	}
	err = worktreeRepo.Create(ctx, worktree)
	require.NoError(t, err)

	return worktree
}

func TestWorktreeRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	// Create test project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create test task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err = taskRepo.Create(ctx, task)
	require.NoError(t, err)

	worktree := &entity.Worktree{
		TaskID:       task.ID,
		ProjectID:    project.ID,
		BranchName:   "test-branch",
		WorktreePath: "/tmp/test-worktree",
		Status:       entity.WorktreeStatusCreating,
	}

	err = worktreeRepo.Create(ctx, worktree)
	require.NoError(t, err)

	// Verify the worktree was created
	assert.NotEqual(t, uuid.Nil, worktree.ID)
	assert.NotZero(t, worktree.CreatedAt)
	assert.NotZero(t, worktree.UpdatedAt)
	assert.Equal(t, entity.WorktreeStatusCreating, worktree.Status)
}

func TestWorktreeRepository_CreateWithDefaultStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	// Create test project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create test task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err = taskRepo.Create(ctx, task)
	require.NoError(t, err)

	worktree := &entity.Worktree{
		TaskID:       task.ID,
		ProjectID:    project.ID,
		BranchName:   "test-branch",
		WorktreePath: "/tmp/test-worktree",
		// Status not set, should default to creating
	}

	err = worktreeRepo.Create(ctx, worktree)
	require.NoError(t, err)

	assert.Equal(t, entity.WorktreeStatusCreating, worktree.Status)
}

func TestWorktreeRepository_CreateWithInvalidTaskID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := &entity.Worktree{
		TaskID:       uuid.New(), // Non-existent task
		ProjectID:    uuid.New(), // Non-existent project
		BranchName:   "test-branch",
		WorktreePath: "/tmp/test-worktree",
		Status:       entity.WorktreeStatusCreating,
	}

	err := worktreeRepo.Create(ctx, worktree)
	assert.Error(t, err)
	// GORM will return a foreign key constraint error
	assert.Contains(t, err.Error(), "failed to create worktree")
}

func TestWorktreeRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Retrieve the worktree
	retrievedWorktree, err := worktreeRepo.GetByID(ctx, worktree.ID)
	require.NoError(t, err)

	assert.Equal(t, worktree.ID, retrievedWorktree.ID)
	assert.Equal(t, worktree.TaskID, retrievedWorktree.TaskID)
	assert.Equal(t, worktree.ProjectID, retrievedWorktree.ProjectID)
	assert.Equal(t, worktree.BranchName, retrievedWorktree.BranchName)
	assert.Equal(t, worktree.WorktreePath, retrievedWorktree.WorktreePath)
	assert.Equal(t, worktree.Status, retrievedWorktree.Status)
}

func TestWorktreeRepository_GetByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	_, err := worktreeRepo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWorktreeRepository_GetByTaskID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Retrieve the worktree by task ID
	retrievedWorktree, err := worktreeRepo.GetByTaskID(ctx, worktree.TaskID)
	require.NoError(t, err)

	assert.Equal(t, worktree.ID, retrievedWorktree.ID)
	assert.Equal(t, worktree.TaskID, retrievedWorktree.TaskID)
}

func TestWorktreeRepository_GetByTaskID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	_, err := worktreeRepo.GetByTaskID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found for task")
}

func TestWorktreeRepository_GetByProjectID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Retrieve worktrees by project ID
	worktrees, err := worktreeRepo.GetByProjectID(ctx, worktree.ProjectID)
	require.NoError(t, err)

	assert.Len(t, worktrees, 1)
	assert.Equal(t, worktree.ID, worktrees[0].ID)
}

func TestWorktreeRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Update the worktree
	worktree.Status = entity.WorktreeStatusActive
	worktree.BranchName = "updated-branch"

	err := worktreeRepo.Update(ctx, worktree)
	require.NoError(t, err)

	// Retrieve and verify the update
	retrievedWorktree, err := worktreeRepo.GetByID(ctx, worktree.ID)
	require.NoError(t, err)

	assert.Equal(t, entity.WorktreeStatusActive, retrievedWorktree.Status)
	assert.Equal(t, "updated-branch", retrievedWorktree.BranchName)
}

func TestWorktreeRepository_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := &entity.Worktree{
		ID:           uuid.New(),
		TaskID:       uuid.New(),
		ProjectID:    uuid.New(),
		BranchName:   "test-branch",
		WorktreePath: "/tmp/test-worktree",
		Status:       entity.WorktreeStatusActive,
	}

	err := worktreeRepo.Update(ctx, worktree)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWorktreeRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Delete the worktree
	err := worktreeRepo.Delete(ctx, worktree.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = worktreeRepo.GetByID(ctx, worktree.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWorktreeRepository_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	err := worktreeRepo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWorktreeRepository_UpdateStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Update status
	err := worktreeRepo.UpdateStatus(ctx, worktree.ID, entity.WorktreeStatusActive)
	require.NoError(t, err)

	// Verify the status was updated
	retrievedWorktree, err := worktreeRepo.GetByID(ctx, worktree.ID)
	require.NoError(t, err)

	assert.Equal(t, entity.WorktreeStatusActive, retrievedWorktree.Status)
}

func TestWorktreeRepository_UpdateStatus_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	err := worktreeRepo.UpdateStatus(ctx, uuid.New(), entity.WorktreeStatusActive)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWorktreeRepository_GetByStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Get worktrees by status
	worktrees, err := worktreeRepo.GetByStatus(ctx, entity.WorktreeStatusCreating)
	require.NoError(t, err)

	assert.Len(t, worktrees, 1)
	assert.Equal(t, worktree.ID, worktrees[0].ID)
}

func TestWorktreeRepository_GetByStatuses(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Get worktrees by multiple statuses
	worktrees, err := worktreeRepo.GetByStatuses(ctx, []entity.WorktreeStatus{
		entity.WorktreeStatusCreating,
		entity.WorktreeStatusActive,
	})
	require.NoError(t, err)

	assert.Len(t, worktrees, 1)
	assert.Equal(t, worktree.ID, worktrees[0].ID)
}

func TestWorktreeRepository_GetByBranchName(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Get worktrees by branch name
	worktrees, err := worktreeRepo.GetByBranchName(ctx, "test-branch")
	require.NoError(t, err)

	assert.Len(t, worktrees, 1)
	assert.Equal(t, worktree.ID, worktrees[0].ID)
}

func TestWorktreeRepository_GetByWorktreePath(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Get worktree by path
	retrievedWorktree, err := worktreeRepo.GetByWorktreePath(ctx, "/tmp/test-worktree")
	require.NoError(t, err)

	assert.Equal(t, worktree.ID, retrievedWorktree.ID)
}

func TestWorktreeRepository_GetByWorktreePath_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	_, err := worktreeRepo.GetByWorktreePath(ctx, "/non-existent-path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found with path")
}

func TestWorktreeRepository_CheckDuplicateWorktreePath(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Check for duplicate path
	exists, err := worktreeRepo.CheckDuplicateWorktreePath(ctx, "/tmp/test-worktree", nil)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check with exclude ID
	exists, err = worktreeRepo.CheckDuplicateWorktreePath(ctx, "/tmp/test-worktree", &worktree.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Check for non-existent path
	exists, err = worktreeRepo.CheckDuplicateWorktreePath(ctx, "/non-existent-path", nil)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestWorktreeRepository_CheckDuplicateBranchName(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Check for duplicate branch name in same project
	exists, err := worktreeRepo.CheckDuplicateBranchName(ctx, worktree.ProjectID, "test-branch", nil)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check with exclude ID
	exists, err = worktreeRepo.CheckDuplicateBranchName(ctx, worktree.ProjectID, "test-branch", &worktree.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Check for non-existent branch name
	exists, err = worktreeRepo.CheckDuplicateBranchName(ctx, worktree.ProjectID, "non-existent-branch", nil)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestWorktreeRepository_GetWorktreeStatistics(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Get statistics
	stats, err := worktreeRepo.GetWorktreeStatistics(ctx, worktree.ProjectID)
	require.NoError(t, err)

	assert.Equal(t, worktree.ProjectID, stats.ProjectID)
	assert.Equal(t, 1, stats.TotalWorktrees)
	assert.Equal(t, 0, stats.ActiveWorktrees)
	assert.Equal(t, 0, stats.CompletedWorktrees)
	assert.Equal(t, 0, stats.ErrorWorktrees)
	assert.Equal(t, 1, stats.WorktreesByStatus[entity.WorktreeStatusCreating])
}

func TestWorktreeRepository_GetActiveWorktreesCount(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Update to active status
	err := worktreeRepo.UpdateStatus(ctx, worktree.ID, entity.WorktreeStatusActive)
	require.NoError(t, err)

	// Get active count
	count, err := worktreeRepo.GetActiveWorktreesCount(ctx, worktree.ProjectID)
	require.NoError(t, err)

	assert.Equal(t, 1, count)
}

func TestWorktreeRepository_GetWorktreesByStatusCount(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Get status counts
	statusCounts, err := worktreeRepo.GetWorktreesByStatusCount(ctx, worktree.ProjectID)
	require.NoError(t, err)

	assert.Equal(t, 1, statusCounts[entity.WorktreeStatusCreating])
}

func TestWorktreeRepository_GetWorktreesWithFilters(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Test filters
	filters := entity.WorktreeFilters{
		ProjectID: &worktree.ProjectID,
		Statuses:  []entity.WorktreeStatus{entity.WorktreeStatusCreating},
	}

	worktrees, err := worktreeRepo.GetWorktreesWithFilters(ctx, filters)
	require.NoError(t, err)

	assert.Len(t, worktrees, 1)
	assert.Equal(t, worktree.ID, worktrees[0].ID)
}

func TestWorktreeRepository_BulkOperations(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Test bulk update status
	err := worktreeRepo.BulkUpdateStatus(ctx, []uuid.UUID{worktree.ID}, entity.WorktreeStatusActive)
	require.NoError(t, err)

	// Verify update
	retrievedWorktree, err := worktreeRepo.GetByID(ctx, worktree.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.WorktreeStatusActive, retrievedWorktree.Status)

	// Test bulk delete
	err = worktreeRepo.BulkDelete(ctx, []uuid.UUID{worktree.ID})
	require.NoError(t, err)

	// Verify deletion
	_, err = worktreeRepo.GetByID(ctx, worktree.ID)
	assert.Error(t, err)
}

func TestWorktreeRepository_ValidationMethods(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	worktreeRepo := NewWorktreeRepository(db)
	ctx := context.Background()

	worktree := createTestWorktree(t, worktreeRepo, taskRepo, projectRepo, ctx)

	// Test validation methods
	exists, err := worktreeRepo.ValidateWorktreeExists(ctx, worktree.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = worktreeRepo.ValidateTaskExists(ctx, worktree.TaskID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = worktreeRepo.ValidateProjectExists(ctx, worktree.ProjectID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test with non-existent IDs
	exists, err = worktreeRepo.ValidateWorktreeExists(ctx, uuid.New())
	require.NoError(t, err)
	assert.False(t, exists)
}
