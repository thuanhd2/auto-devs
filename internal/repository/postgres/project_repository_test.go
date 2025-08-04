package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	project := &entity.Project{
		Name:          "Test Project",
		Description:   "Test Description",
		RepositoryURL: "https://github.com/test/repo.git",
	}

	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Verify the project was created
	assert.NotEqual(t, uuid.Nil, project.ID)
	assert.NotZero(t, project.CreatedAt)
	assert.NotZero(t, project.UpdatedAt)
}

func TestProjectRepository_CreateWithExistingID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	projectID := uuid.New()
	project := &entity.Project{
		ID:            projectID,
		Name:          "Test Project",
		Description:   "Test Description",
		RepositoryURL: "https://github.com/test/repo.git",
	}

	err := repo.Create(ctx, project)
	require.NoError(t, err)

	assert.Equal(t, projectID, project.ID)
}

func TestProjectRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Get project
	retrieved, err := repo.GetByID(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, project.ID, retrieved.ID)
	assert.Equal(t, project.Name, retrieved.Name)
	assert.Equal(t, project.Description, retrieved.Description)
	assert.Equal(t, project.RepositoryURL, retrieved.RepositoryURL)
}

func TestProjectRepository_GetByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestProjectRepository_GetAll(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create multiple projects
	project1 := &entity.Project{
		Name:        "Project 1",
		Description: "Description 1",
		RepositoryURL:     "https://github.com/test/repo1.git",
	}
	project2 := &entity.Project{
		Name:        "Project 2",
		Description: "Description 2",
		RepositoryURL:     "https://github.com/test/repo2.git",
	}

	err := repo.Create(ctx, project1)
	require.NoError(t, err)
	err = repo.Create(ctx, project2)
	require.NoError(t, err)

	// Get all projects
	projects, err := repo.GetAll(ctx)
	require.NoError(t, err)

	assert.Len(t, projects, 2)
	// Projects should be ordered by created_at DESC (newest first)
	assert.Equal(t, project2.ID, projects[0].ID)
	assert.Equal(t, project1.ID, projects[1].ID)
}

func TestProjectRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Original Name",
		Description: "Original Description",
		RepositoryURL:     "https://github.com/test/original.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	originalUpdatedAt := project.UpdatedAt

	// Update project
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	project.Name = "Updated Name"
	project.Description = "Updated Description"
	project.RepositoryURL = "https://github.com/test/updated.git"

	err = repo.Update(ctx, project)
	require.NoError(t, err)

	// Verify updates
	assert.True(t, project.UpdatedAt.After(originalUpdatedAt))

	// Get and verify
	retrieved, err := repo.GetByID(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
	assert.Equal(t, "https://github.com/test/updated.git", retrieved.RepositoryURL)
}

func TestProjectRepository_Update_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	project := &entity.Project{
		ID:          uuid.New(),
		Name:        "Non-existent",
		Description: "Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}

	err := repo.Update(ctx, project)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "project not found")
	}
}

func TestProjectRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Delete project (soft delete)
	err = repo.Delete(ctx, project.ID)
	require.NoError(t, err)

	// Verify deletion (should not be found due to soft delete)
	_, err = repo.GetByID(ctx, project.ID)
	assert.Error(t, err)
}

func TestProjectRepository_Delete_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestProjectRepository_GetWithTaskCount(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create tasks
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTODO,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusDONE,
	}

	err = taskRepo.Create(ctx, task1)
	require.NoError(t, err)
	err = taskRepo.Create(ctx, task2)
	require.NoError(t, err)

	// Get project with task count
	result, err := projectRepo.GetWithTaskCount(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, project.ID, result.Project.ID)
	assert.Equal(t, project.Name, result.Project.Name)
	assert.Equal(t, 2, result.TaskCount)
}

func TestProjectRepository_GetWithTaskCount_NoTasks(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Get project with task count
	result, err := repo.GetWithTaskCount(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, project.ID, result.Project.ID)
	assert.Equal(t, 0, result.TaskCount)
}

func TestProjectRepository_Delete_WithTasks(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Project with Tasks",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}
	err = taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Delete project (soft delete)
	err = projectRepo.Delete(ctx, project.ID)
	require.NoError(t, err)

	// Verify project is soft deleted (should not be found)
	_, err = projectRepo.GetByID(ctx, project.ID)
	assert.Error(t, err)

	// Note: GORM soft delete doesn't automatically cascade to related records
	// The task should still exist but the project should be soft deleted
	// This is different from raw SQL where CASCADE would delete related records
	_, err = taskRepo.GetByID(ctx, task.ID)
	// Task should still exist since GORM doesn't auto-cascade soft deletes
	assert.NoError(t, err)
}

func TestProjectRepository_GetAllWithParams(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create test projects
	projects := []*entity.Project{
		{Name: "Alpha Project", Description: "First project", RepositoryURL: "https://github.com/test/alpha.git"},
		{Name: "Beta Search", Description: "Second project", RepositoryURL: "https://github.com/test/beta.git"},
		{Name: "Gamma Project", Description: "Third search term", RepositoryURL: "https://github.com/test/gamma.git"},
	}

	for _, p := range projects {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	t.Run("search functionality", func(t *testing.T) {
		params := repository.GetProjectsParams{
			Search:   "search",
			Page:     1,
			PageSize: 10,
		}

		results, total, err := repo.GetAllWithParams(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)
	})

	t.Run("pagination", func(t *testing.T) {
		params := repository.GetProjectsParams{
			Page:     1,
			PageSize: 2,
		}

		results, total, err := repo.GetAllWithParams(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 2)

		params.Page = 2
		results, total, err = repo.GetAllWithParams(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 1)
	})

	t.Run("sorting by name", func(t *testing.T) {
		params := repository.GetProjectsParams{
			SortBy:    "name",
			SortOrder: "asc",
			Page:      1,
			PageSize:  10,
		}

		results, _, err := repo.GetAllWithParams(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, "Alpha Project", results[0].Name)
		assert.Equal(t, "Beta Search", results[1].Name)
		assert.Equal(t, "Gamma Project", results[2].Name)
	})
}

func TestProjectRepository_CheckNameExists(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Unique Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	t.Run("existing name", func(t *testing.T) {
		exists, err := repo.CheckNameExists(ctx, "Unique Project", nil)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing name", func(t *testing.T) {
		exists, err := repo.CheckNameExists(ctx, "Non-existent Project", nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("exclude current project", func(t *testing.T) {
		exists, err := repo.CheckNameExists(ctx, "Unique Project", &project.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestProjectRepository_GetTaskStatistics(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Stats Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create tasks with different statuses
	tasks := []*entity.Task{
		{ProjectID: project.ID, Title: "Task 1", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 2", Status: entity.TaskStatusTODO},
		{ProjectID: project.ID, Title: "Task 3", Status: entity.TaskStatusDONE},
		{ProjectID: project.ID, Title: "Task 4", Status: entity.TaskStatusIMPLEMENTING},
	}

	for _, task := range tasks {
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Get statistics
	stats, err := projectRepo.GetTaskStatistics(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, 2, stats[entity.TaskStatusTODO])
	assert.Equal(t, 1, stats[entity.TaskStatusDONE])
	assert.Equal(t, 1, stats[entity.TaskStatusIMPLEMENTING])
}

func TestProjectRepository_Archive_Restore(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Archive Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Archive project
	err = repo.Archive(ctx, project.ID)
	require.NoError(t, err)

	// Verify it's archived (not found in normal queries)
	_, err = repo.GetByID(ctx, project.ID)
	assert.Error(t, err)

	// Restore project
	err = repo.Restore(ctx, project.ID)
	require.NoError(t, err)

	// Verify it's restored
	restored, err := repo.GetByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, project.ID, restored.ID)
}

func TestProjectRepository_GetLastActivityAt(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Activity Project",
		Description: "Test Description",
		RepositoryURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	t.Run("no tasks - returns project updated_at", func(t *testing.T) {
		lastActivity, err := projectRepo.GetLastActivityAt(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, lastActivity)
		// Should be approximately the project's updated_at
		timeDiff := lastActivity.Sub(project.UpdatedAt)
		assert.Less(t, timeDiff.Milliseconds(), int64(1000)) // Within 1 second
	})

	t.Run("with tasks - returns latest task updated_at", func(t *testing.T) {
		// Create task
		task := &entity.Task{
			ProjectID: project.ID,
			Title:     "Recent Task",
			Status:    entity.TaskStatusTODO,
		}
		err := taskRepo.Create(ctx, task)
		require.NoError(t, err)

		// Update task to have a newer timestamp
		time.Sleep(10 * time.Millisecond)
		task.Title = "Updated Task"
		err = taskRepo.Update(ctx, task)
		require.NoError(t, err)

		lastActivity, err := projectRepo.GetLastActivityAt(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, lastActivity)
		// Should be the task's updated_at, which is newer than project's
		assert.True(t, lastActivity.After(project.UpdatedAt))
	})
}
