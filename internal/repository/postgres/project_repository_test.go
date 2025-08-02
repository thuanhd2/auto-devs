package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	ctx := context.Background()

	// Create PostgreSQL test container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection details
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create database connection
	config := database.Config{
		Host:     host,
		Port:     port.Port(),
		Username: "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	db, err := database.NewConnection(config)
	require.NoError(t, err)

	// Run migrations
	err = runTestMigrations(db)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func runTestMigrations(db *database.DB) error {
	// Create the same schema as in migrations
	schema := `
		-- Enable UUID extension
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		-- Create projects table
		CREATE TABLE projects (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			repo_url VARCHAR(500) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Create tasks table
		CREATE TABLE tasks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
			project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'TODO',
			branch_name VARCHAR(255),
			pull_request VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT valid_status CHECK (
				status IN (
					'TODO',
					'PLANNING',
					'PLAN_REVIEWING',
					'IMPLEMENTING',
					'CODE_REVIEWING',
					'DONE',
					'CANCELLED'
				)
			)
		);

		-- Create indexes for better performance
		CREATE INDEX idx_tasks_project_id ON tasks (project_id);
		CREATE INDEX idx_tasks_status ON tasks (status);
		CREATE INDEX idx_tasks_created_at ON tasks (created_at);
		CREATE INDEX idx_projects_created_at ON projects (created_at);

		-- Create updated_at trigger function
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		-- Create triggers for auto-updating updated_at columns
		CREATE TRIGGER update_projects_updated_at 
			BEFORE UPDATE ON projects 
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

		CREATE TRIGGER update_tasks_updated_at 
			BEFORE UPDATE ON tasks 
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := db.Exec(schema)
	return err
}

func TestProjectRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}

	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Verify the project was created
	assert.NotEqual(t, uuid.Nil, project.ID)
	assert.NotZero(t, project.CreatedAt)
	assert.NotZero(t, project.UpdatedAt)
}

func TestProjectRepository_CreateWithExistingID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	projectID := uuid.New()
	project := &entity.Project{
		ID:          projectID,
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}

	err := repo.Create(ctx, project)
	require.NoError(t, err)

	assert.Equal(t, projectID, project.ID)
}

func TestProjectRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Get project
	retrieved, err := repo.GetByID(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, project.ID, retrieved.ID)
	assert.Equal(t, project.Name, retrieved.Name)
	assert.Equal(t, project.Description, retrieved.Description)
	assert.Equal(t, project.RepoURL, retrieved.RepoURL)
}

func TestProjectRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestProjectRepository_GetAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create multiple projects
	project1 := &entity.Project{
		Name:        "Project 1",
		Description: "Description 1",
		RepoURL:     "https://github.com/test/repo1.git",
	}
	project2 := &entity.Project{
		Name:        "Project 2",
		Description: "Description 2",
		RepoURL:     "https://github.com/test/repo2.git",
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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Original Name",
		Description: "Original Description",
		RepoURL:     "https://github.com/test/original.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	originalUpdatedAt := project.UpdatedAt

	// Update project
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	project.Name = "Updated Name"
	project.Description = "Updated Description"
	project.RepoURL = "https://github.com/test/updated.git"

	err = repo.Update(ctx, project)
	require.NoError(t, err)

	// Verify updates
	assert.True(t, project.UpdatedAt.After(originalUpdatedAt))

	// Get and verify
	retrieved, err := repo.GetByID(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
	assert.Equal(t, "https://github.com/test/updated.git", retrieved.RepoURL)
}

func TestProjectRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	project := &entity.Project{
		ID:          uuid.New(),
		Name:        "Non-existent",
		Description: "Description",
		RepoURL:     "https://github.com/test/repo.git",
	}

	err := repo.Update(ctx, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestProjectRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := repo.Create(ctx, project)
	require.NoError(t, err)

	// Delete project
	err = repo.Delete(ctx, project.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, project.ID)
	assert.Error(t, err)
}

func TestProjectRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestProjectRepository_GetWithTaskCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create tasks
	task1 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 1",
		Description: "Description 1",
		Status:      entity.TaskStatusTodo,
	}
	task2 := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      entity.TaskStatusDone,
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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Test Project",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	projectRepo := NewProjectRepository(db)
	taskRepo := NewTaskRepository(db)
	ctx := context.Background()

	// Create project
	project := &entity.Project{
		Name:        "Project with Tasks",
		Description: "Test Description",
		RepoURL:     "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create task
	task := &entity.Task{
		ProjectID:   project.ID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTodo,
	}
	err = taskRepo.Create(ctx, task)
	require.NoError(t, err)

	// Try to delete project with tasks (should work due to CASCADE)
	err = projectRepo.Delete(ctx, project.ID)
	require.NoError(t, err)

	// Verify project is deleted
	_, err = projectRepo.GetByID(ctx, project.ID)
	assert.Error(t, err)
	
	// Verify task is also deleted due to CASCADE
	_, err = taskRepo.GetByID(ctx, task.ID)
	assert.Error(t, err)
}