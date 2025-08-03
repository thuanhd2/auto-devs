package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProjectUsecase is a mock implementation of ProjectUsecase
type MockProjectUsecase struct {
	mock.Mock
}

func (m *MockProjectUsecase) Create(ctx context.Context, req usecase.CreateProjectRequest) (*entity.Project, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectUsecase) GetAll(ctx context.Context, params usecase.GetProjectsParams) (*usecase.GetProjectsResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.GetProjectsResult), args.Error(1)
}

func (m *MockProjectUsecase) Update(ctx context.Context, id uuid.UUID, req usecase.UpdateProjectRequest) (*entity.Project, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectUsecase) GetWithTasks(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectUsecase) GetStatistics(ctx context.Context, id uuid.UUID) (*usecase.ProjectStatistics, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ProjectStatistics), args.Error(1)
}

func (m *MockProjectUsecase) Archive(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectUsecase) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectUsecase) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectUsecase) GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(*entity.ProjectSettings), args.Error(1)
}

func (m *MockProjectUsecase) UpdateSettings(ctx context.Context, projectID uuid.UUID, settings *entity.ProjectSettings) (*entity.ProjectSettings, error) {
	args := m.Called(ctx, projectID, settings)
	return args.Get(0).(*entity.ProjectSettings), args.Error(1)
}

func setupProjectHandler() (*ProjectHandler, *MockProjectUsecase) {
	mockUsecase := new(MockProjectUsecase)
	handler := NewProjectHandler(mockUsecase)
	return handler, mockUsecase
}

func setupGinRouter(handler *ProjectHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	projects := v1.Group("/projects")
	{
		projects.POST("", handler.CreateProject)
		projects.GET("", handler.ListProjects)
		projects.GET("/:id", handler.GetProject)
		projects.PUT("/:id", handler.UpdateProject)
		projects.DELETE("/:id", handler.DeleteProject)
		projects.GET("/:id/tasks", handler.GetProjectWithTasks)
		projects.GET("/:id/statistics", handler.GetProjectStatistics)
		projects.POST("/:id/archive", handler.ArchiveProject)
		projects.POST("/:id/restore", handler.RestoreProject)
	}

	return router
}

func TestProjectHandler_CreateProject(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	t.Run("successful creation", func(t *testing.T) {
		project := &entity.Project{
			ID:          uuid.New(),
			Name:        "Test Project",
			Description: "Test Description",
			RepoURL:     "https://github.com/test/repo.git",
		}

		mockUsecase.On("Create", mock.Anything, mock.MatchedBy(func(req usecase.CreateProjectRequest) bool {
			return req.Name == "Test Project" && req.RepoURL == "https://github.com/test/repo.git"
		})).Return(project, nil)

		reqBody := dto.ProjectCreateRequest{
			Name:        "Test Project",
			Description: "Test Description",
			RepoURL:     "https://github.com/test/repo.git",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.ProjectResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, project.ID, response.ID)
		assert.Equal(t, project.Name, response.Name)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		reqBody := dto.ProjectCreateRequest{
			Name: "", // Invalid empty name
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProjectHandler_ListProjects(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	t.Run("successful list with pagination", func(t *testing.T) {
		projects := []*entity.Project{
			{ID: uuid.New(), Name: "Project 1", RepoURL: "https://github.com/test/repo1.git"},
			{ID: uuid.New(), Name: "Project 2", RepoURL: "https://github.com/test/repo2.git"},
		}

		result := &usecase.GetProjectsResult{
			Projects: projects,
			Total:    2,
			Page:     1,
			PageSize: 10,
		}

		mockUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(params usecase.GetProjectsParams) bool {
			return params.Page == 1 && params.PageSize == 10
		})).Return(result, nil)

		req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Projects, 2)
		assert.Equal(t, 2, response.Total)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("with search and sorting", func(t *testing.T) {
		result := &usecase.GetProjectsResult{
			Projects: []*entity.Project{},
			Total:    0,
			Page:     1,
			PageSize: 10,
		}

		mockUsecase.On("GetAll", mock.Anything, mock.MatchedBy(func(params usecase.GetProjectsParams) bool {
			return params.Search == "test" && params.SortBy == "name" && params.SortOrder == "asc"
		})).Return(result, nil)

		req, _ := http.NewRequest("GET", "/api/v1/projects?search=test&sort_by=name&sort_order=asc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUsecase.AssertExpectations(t)
	})
}

func TestProjectHandler_GetProject(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	t.Run("successful get", func(t *testing.T) {
		projectID := uuid.New()
		project := &entity.Project{
			ID:      projectID,
			Name:    "Test Project",
			RepoURL: "https://github.com/test/repo.git",
		}

		mockUsecase.On("GetByID", mock.Anything, projectID).Return(project, nil)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%s", projectID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, projectID, response.ID)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/projects/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProjectHandler_GetProjectStatistics(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	projectID := uuid.New()
	stats := &usecase.ProjectStatistics{
		TaskCounts: map[entity.TaskStatus]int{
			entity.TaskStatusTodo: 3,
			entity.TaskStatusDone: 2,
		},
		TotalTasks:        5,
		CompletionPercent: 40.0,
	}

	mockUsecase.On("GetStatistics", mock.Anything, projectID).Return(stats, nil)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%s/statistics", projectID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ProjectStatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 5, response.TotalTasks)
	assert.Equal(t, 40.0, response.CompletionPercent)
	assert.Equal(t, 3, response.TaskCounts[entity.TaskStatusTodo])

	mockUsecase.AssertExpectations(t)
}

func TestProjectHandler_ArchiveProject(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	projectID := uuid.New()
	mockUsecase.On("Archive", mock.Anything, projectID).Return(nil)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%s/archive", projectID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestProjectHandler_RestoreProject(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	projectID := uuid.New()
	mockUsecase.On("Restore", mock.Anything, projectID).Return(nil)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%s/restore", projectID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestProjectHandler_UpdateProject(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	projectID := uuid.New()
	updatedProject := &entity.Project{
		ID:      projectID,
		Name:    "Updated Project",
		RepoURL: "https://github.com/test/updated.git",
	}

	mockUsecase.On("Update", mock.Anything, projectID, mock.MatchedBy(func(req usecase.UpdateProjectRequest) bool {
		return req.Name == "Updated Project"
	})).Return(updatedProject, nil)

	reqBody := dto.ProjectUpdateRequest{
		Name:    stringPtr("Updated Project"),
		RepoURL: stringPtr("https://github.com/test/updated.git"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s", projectID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ProjectResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Project", response.Name)

	mockUsecase.AssertExpectations(t)
}

func TestProjectHandler_DeleteProject(t *testing.T) {
	handler, mockUsecase := setupProjectHandler()
	router := setupGinRouter(handler)

	projectID := uuid.New()
	mockUsecase.On("Delete", mock.Anything, projectID).Return(nil)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s", projectID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUsecase.AssertExpectations(t)
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
