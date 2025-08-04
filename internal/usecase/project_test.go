package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProjectUsecase_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		// Setup mocks
		projectRepo := &testutil.MockProjectRepository{}

		// Mock expectations
		projectRepo.On("CheckNameExists", ctx, "Test Project", (*uuid.UUID)(nil)).Return(false, nil)
		projectRepo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).
			Run(func(args mock.Arguments) {
				project := args.Get(1).(*entity.Project)
				// Verify the project fields are set correctly
				assert.Equal(t, "Test Project", project.Name)
				assert.Equal(t, "Test Description", project.Description)
				assert.Equal(t, "https://github.com/test/repo.git", project.RepoURL)
				assert.NotEqual(t, uuid.Nil, project.ID)
			}).Return(nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			RepoURL:     "https://github.com/test/repo.git",
		}

		result, err := usecase.Create(ctx, req)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Description, result.Description)
		assert.Equal(t, req.RepoURL, result.RepoURL)
		assert.NotEqual(t, uuid.Nil, result.ID)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		// Setup mocks
		projectRepo := &testutil.MockProjectRepository{}

		// Mock expectations
		projectRepo.On("CheckNameExists", ctx, "Existing Project", (*uuid.UUID)(nil)).Return(true, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := CreateProjectRequest{
			Name:        "Existing Project",
			Description: "Test Description",
			RepoURL:     "https://github.com/test/repo.git",
		}

		result, err := usecase.Create(ctx, req)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project name already exists")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("invalid repository URL", func(t *testing.T) {
		// Setup mocks
		projectRepo := &testutil.MockProjectRepository{}

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			RepoURL:     "not-a-valid-url",
		}

		result, err := usecase.Create(ctx, req)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid repository URL")
	})

	t.Run("repository creation fails", func(t *testing.T) {
		// Setup mocks
		projectRepo := &testutil.MockProjectRepository{}

		// Mock expectations
		projectRepo.On("CheckNameExists", ctx, "Test Project", (*uuid.UUID)(nil)).Return(false, nil)
		projectRepo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).Return(errors.New("database error"))

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			RepoURL:     "https://github.com/test/repo.git",
		}

		result, err := usecase.Create(ctx, req)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		project := projectFactory.CreateProject()

		// Mock expectations
		projectRepo.On("GetByID", ctx, project.ID).Return(project, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		result, err := usecase.GetByID(ctx, project.ID)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, project, result)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("GetByID", ctx, projectID).Return(nil, errors.New("project not found"))

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		result, err := usecase.GetByID(ctx, projectID)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project not found")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_GetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval with pagination", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		projects := []*entity.Project{
			projectFactory.CreateProject(),
			projectFactory.CreateProject(),
		}

		// Mock expectations
		expectedParams := repository.GetProjectsParams{
			Search:    "test",
			SortBy:    "name",
			SortOrder: "asc",
			Page:      1,
			PageSize:  10,
		}

		projectRepo.On("GetAllWithParams", ctx, expectedParams).Return(projects, 2, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		params := GetProjectsParams{
			Search:    "test",
			SortBy:    "name",
			SortOrder: "asc",
			Page:      1,
			PageSize:  10,
		}

		result, err := usecase.GetAll(ctx, params)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Projects, 2)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		// Mock expectations
		projectRepo.On("GetAllWithParams", ctx, mock.AnythingOfType("repository.GetProjectsParams")).
			Return(nil, 0, errors.New("database error"))

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		params := GetProjectsParams{
			Page:     1,
			PageSize: 10,
		}

		result, err := usecase.GetAll(ctx, params)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update all fields", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		originalProject := projectFactory.CreateProject()

		// Mock expectations
		projectRepo.On("GetByID", ctx, originalProject.ID).Return(originalProject, nil)
		projectRepo.On("CheckNameExists", ctx, "Updated Project", &originalProject.ID).Return(false, nil)
		projectRepo.On("Update", ctx, mock.AnythingOfType("*entity.Project")).
			Run(func(args mock.Arguments) {
				project := args.Get(1).(*entity.Project)
				assert.Equal(t, "Updated Project", project.Name)
				assert.Equal(t, "Updated Description", project.Description)
				assert.Equal(t, "https://github.com/test/updated.git", project.RepoURL)
			}).Return(nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := UpdateProjectRequest{
			Name:        "Updated Project",
			Description: "Updated Description",
			RepoURL:     "https://github.com/test/updated.git",
		}

		result, err := usecase.Update(ctx, originalProject.ID, req)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("partial update", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		originalProject := projectFactory.CreateProject()

		// Mock expectations
		projectRepo.On("GetByID", ctx, originalProject.ID).Return(originalProject, nil)
		projectRepo.On("Update", ctx, mock.AnythingOfType("*entity.Project")).
			Run(func(args mock.Arguments) {
				project := args.Get(1).(*entity.Project)
				assert.Equal(t, "Updated Name", project.Name)
				// Description and RepoURL should remain unchanged
				assert.Equal(t, originalProject.Description, project.Description)
				assert.Equal(t, originalProject.RepoURL, project.RepoURL)
			}).Return(nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute - only update name
		req := UpdateProjectRequest{
			Name: "Updated Name",
			// Other fields empty, should not be updated
		}

		result, err := usecase.Update(ctx, originalProject.ID, req)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		project := projectFactory.CreateProject()

		// Mock expectations
		projectRepo.On("GetByID", ctx, project.ID).Return(project, nil)
		projectRepo.On("CheckNameExists", ctx, "Existing Name", &project.ID).Return(true, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := UpdateProjectRequest{
			Name: "Existing Name",
		}

		result, err := usecase.Update(ctx, project.ID, req)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project name already exists")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("GetByID", ctx, projectID).Return(nil, errors.New("project not found"))

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := UpdateProjectRequest{
			Name: "Updated Name",
		}

		result, err := usecase.Update(ctx, projectID, req)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project not found")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("invalid repository URL", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		project := projectFactory.CreateProject()

		// Mock expectations
		projectRepo.On("GetByID", ctx, project.ID).Return(project, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		req := UpdateProjectRequest{
			RepoURL: "not-a-valid-url",
		}

		result, err := usecase.Update(ctx, project.ID, req)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid repository URL")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("Delete", ctx, projectID).Return(nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		err := usecase.Delete(ctx, projectID)

		// Assertions
		require.NoError(t, err)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("deletion fails", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("Delete", ctx, projectID).Return(errors.New("deletion failed"))

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		err := usecase.Delete(ctx, projectID)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deletion failed")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_GetStatistics(t *testing.T) {
	ctx := context.Background()

	t.Run("successful statistics retrieval", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectFactory := testutil.NewProjectFactory()
		project := projectFactory.CreateProject()

		taskStats := map[entity.TaskStatus]int{
			entity.TaskStatusTODO:         5,
			entity.TaskStatusIMPLEMENTING: 3,
			entity.TaskStatusDONE:         2,
		}

		// Mock expectations
		projectRepo.On("GetByID", ctx, project.ID).Return(project, nil)
		projectRepo.On("GetTaskStatistics", ctx, project.ID).Return(taskStats, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		result, err := usecase.GetStatistics(ctx, project.ID)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, project, result.Project)
		assert.Equal(t, 10, result.TotalTasks)
		assert.Equal(t, taskStats, result.TasksByStatus)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("GetByID", ctx, projectID).Return(nil, errors.New("project not found"))

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		result, err := usecase.GetStatistics(ctx, projectID)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project not found")

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_Archive_Restore(t *testing.T) {
	ctx := context.Background()

	t.Run("successful archive", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("Archive", ctx, projectID).Return(nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		err := usecase.Archive(ctx, projectID)

		// Assertions
		require.NoError(t, err)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("successful restore", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("Restore", ctx, projectID).Return(nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		err := usecase.Restore(ctx, projectID)

		// Assertions
		require.NoError(t, err)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_CheckNameExists(t *testing.T) {
	ctx := context.Background()

	t.Run("name exists", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		// Mock expectations
		projectRepo.On("CheckNameExists", ctx, "Existing Project", (*uuid.UUID)(nil)).Return(true, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		exists, err := usecase.CheckNameExists(ctx, "Existing Project", nil)

		// Assertions
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("name does not exist", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		// Mock expectations
		projectRepo.On("CheckNameExists", ctx, "New Project", (*uuid.UUID)(nil)).Return(false, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		exists, err := usecase.CheckNameExists(ctx, "New Project", nil)

		// Assertions
		require.NoError(t, err)
		assert.False(t, exists)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})

	t.Run("exclude current project", func(t *testing.T) {
		// Setup
		projectRepo := &testutil.MockProjectRepository{}

		projectID := uuid.New()

		// Mock expectations
		projectRepo.On("CheckNameExists", ctx, "Project Name", &projectID).Return(false, nil)

		// Create usecase
		usecase := NewProjectUsecase(projectRepo)

		// Execute
		exists, err := usecase.CheckNameExists(ctx, "Project Name", &projectID)

		// Assertions
		require.NoError(t, err)
		assert.False(t, exists)

		// Verify mocks
		projectRepo.AssertExpectations(t)
	})
}

func TestProjectUsecase_ValidationLogic(t *testing.T) {
	ctx := context.Background()

	t.Run("validate repository URL formats", func(t *testing.T) {
		projectRepo := &testutil.MockProjectRepository{}
		usecase := NewProjectUsecase(projectRepo)

		testCases := []struct {
			name    string
			url     string
			isValid bool
		}{
			{"https git URL", "https://github.com/user/repo.git", true},
			{"https git URL without .git", "https://github.com/user/repo", true},
			{"ssh git URL", "git@github.com:user/repo.git", true},
			{"invalid URL", "not-a-url", false},
			{"empty URL", "", false},
			{"ftp URL", "ftp://example.com/repo", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.isValid {
					projectRepo.On("CheckNameExists", ctx, "Test", (*uuid.UUID)(nil)).Return(false, nil).Maybe()
					projectRepo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Maybe()
				}

				req := CreateProjectRequest{
					Name:        "Test",
					Description: "Test",
					RepoURL:     tc.url,
				}

				_, err := usecase.Create(ctx, req)

				if tc.isValid {
					assert.NoError(t, err, "URL %s should be valid", tc.url)
				} else {
					assert.Error(t, err, "URL %s should be invalid", tc.url)
					if err != nil {
						assert.Contains(t, err.Error(), "invalid repository URL")
					}
				}
			})
		}
	})

	t.Run("validate project name", func(t *testing.T) {
		projectRepo := &testutil.MockProjectRepository{}
		usecase := NewProjectUsecase(projectRepo)

		testCases := []struct {
			name      string
			projName  string
			shouldErr bool
		}{
			{"normal name", "My Project", false},
			{"empty name", "", true},
			{"whitespace only", "   ", true},
			{"very long name", strings.Repeat("A", 300), false}, // Should be handled by DB constraints
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if !tc.shouldErr {
					projectRepo.On("CheckNameExists", ctx, mock.AnythingOfType("string"), (*uuid.UUID)(nil)).Return(false, nil).Maybe()
					projectRepo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Maybe()
				}

				req := CreateProjectRequest{
					Name:        tc.projName,
					Description: "Test",
					RepoURL:     "https://github.com/test/repo.git",
				}

				_, err := usecase.Create(ctx, req)

				if tc.shouldErr {
					assert.Error(t, err, "Project name '%s' should cause an error", tc.projName)
				} else {
					assert.NoError(t, err, "Project name '%s' should be valid", tc.projName)
				}
			})
		}
	})
}

// Helper to create a mock NewProjectUsecase for testing
func NewProjectUsecase(projectRepo repository.ProjectRepository) ProjectUsecase {
	// This would be the actual implementation
	// For now, return a mock that satisfies the interface
	return &mockProjectUsecase{
		projectRepo: projectRepo,
	}
}

// mockProjectUsecase implements the ProjectUsecase interface for testing
type mockProjectUsecase struct {
	projectRepo repository.ProjectRepository
}

func (u *mockProjectUsecase) Create(ctx context.Context, req CreateProjectRequest) (*entity.Project, error) {
	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("project name cannot be empty")
	}

	// Validate repository URL
	if !isValidRepoURL(req.RepoURL) {
		return nil, errors.New("invalid repository URL")
	}

	// Check if name exists
	exists, err := u.projectRepo.CheckNameExists(ctx, req.Name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("project name already exists")
	}

	// Create project
	project := &entity.Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		RepoURL:     req.RepoURL,
	}

	if err := u.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (u *mockProjectUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	return u.projectRepo.GetByID(ctx, id)
}

func (u *mockProjectUsecase) GetAll(ctx context.Context, params GetProjectsParams) (*GetProjectsResult, error) {
	repoParams := repository.GetProjectsParams{
		Search:    params.Search,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
		Page:      params.Page,
		PageSize:  params.PageSize,
	}

	projects, total, err := u.projectRepo.GetAllWithParams(ctx, repoParams)
	if err != nil {
		return nil, err
	}

	return &GetProjectsResult{
		Projects: projects,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (u *mockProjectUsecase) Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest) (*entity.Project, error) {
	// Get existing project
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate repository URL if provided
	if req.RepoURL != "" && !isValidRepoURL(req.RepoURL) {
		return nil, errors.New("invalid repository URL")
	}

	// Check name uniqueness if provided
	if req.Name != "" {
		exists, err := u.projectRepo.CheckNameExists(ctx, req.Name, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("project name already exists")
		}
		project.Name = req.Name
	}

	// Update other fields
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.RepoURL != "" {
		project.RepoURL = req.RepoURL
	}

	if err := u.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (u *mockProjectUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.projectRepo.Delete(ctx, id)
}

func (u *mockProjectUsecase) GetWithTasks(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	return u.projectRepo.GetByID(ctx, id)
}

func (u *mockProjectUsecase) GetStatistics(ctx context.Context, id uuid.UUID) (*ProjectStatistics, error) {
	project, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	taskStats, err := u.projectRepo.GetTaskStatistics(ctx, id)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, count := range taskStats {
		total += count
	}

	return &ProjectStatistics{
		Project:       project,
		TotalTasks:    total,
		TasksByStatus: taskStats,
	}, nil
}

func (u *mockProjectUsecase) Archive(ctx context.Context, id uuid.UUID) error {
	return u.projectRepo.Archive(ctx, id)
}

func (u *mockProjectUsecase) Restore(ctx context.Context, id uuid.UUID) error {
	return u.projectRepo.Restore(ctx, id)
}

func (u *mockProjectUsecase) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return u.projectRepo.CheckNameExists(ctx, name, excludeID)
}

func (u *mockProjectUsecase) GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error) {
	// Mock implementation
	return nil, nil
}

func (u *mockProjectUsecase) UpdateSettings(ctx context.Context, projectID uuid.UUID, settings *entity.ProjectSettings) (*entity.ProjectSettings, error) {
	// Mock implementation
	return nil, nil
}

// Helper function to validate repository URLs
func isValidRepoURL(urlStr string) bool {
	if strings.TrimSpace(urlStr) == "" {
		return false
	}

	// Check for SSH format
	if matched, _ := regexp.MatchString(`^git@[\w\.-]+:[\w\.-]+/[\w\.-]+(?:\.git)?$`, urlStr); matched {
		return true
	}

	// Check for HTTPS format
	if u, err := url.Parse(urlStr); err == nil && (u.Scheme == "https" || u.Scheme == "http") {
		return true
	}

	return false
}

// Types for the mock implementation
type GetProjectsResult struct {
	Projects []*entity.Project
	Total    int
	Page     int
	PageSize int
}

type ProjectStatistics struct {
	Project       *entity.Project
	TotalTasks    int
	TasksByStatus map[entity.TaskStatus]int
}