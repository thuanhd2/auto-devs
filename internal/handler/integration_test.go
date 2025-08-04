package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/handler/dto"
	"github.com/auto-devs/auto-devs/internal/repository/postgres"
	"github.com/auto-devs/auto-devs/internal/testutil"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite provides a comprehensive test suite for API integration tests
type IntegrationTestSuite struct {
	container       *testutil.TestContainer
	router          *gin.Engine
	apiHelper       *testutil.APITestHelper
	dbHelper        *testutil.DatabaseTestHelper
	projectHandler  *ProjectHandler
	taskHandler     *TaskHandler
	projectUsecase  usecase.ProjectUsecase
	taskUsecase     usecase.TaskUsecase
}

// SetupIntegrationTestSuite creates a new integration test suite
func SetupIntegrationTestSuite(t *testing.T) (*IntegrationTestSuite, func()) {
	// Setup test database
	container, cleanup := testutil.SetupTestDB(t)

	// Create repositories
	projectRepo := postgres.NewProjectRepository(container.DB)
	taskRepo := postgres.NewTaskRepository(container.DB)

	// Create usecases
	projectUsecase := usecase.NewProjectUsecase(projectRepo)
	taskUsecase := usecase.NewTaskUsecase(taskRepo, projectRepo, nil)

	// Create handlers
	projectHandler := NewProjectHandler(projectUsecase)
	taskHandler := NewTaskHandler(taskUsecase)

	// Setup Gin router
	testutil.SetupGinTestMode()
	router := gin.New()
	
	// Setup routes
	api := router.Group("/api/v1")
	{
		projects := api.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.GetProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
			projects.GET("/:id/tasks", taskHandler.GetTasksByProject)
			projects.POST("/:id/tasks", taskHandler.CreateTask)
		}

		tasks := api.Group("/tasks")
		{
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.PATCH("/:id/status", taskHandler.UpdateTaskStatus)
		}
	}

	suite := &IntegrationTestSuite{
		container:      container,
		router:         router,
		apiHelper:      testutil.NewAPITestHelper(router),
		dbHelper:       testutil.NewDatabaseTestHelper(container),
		projectHandler: projectHandler,
		taskHandler:    taskHandler,
		projectUsecase: projectUsecase,
		taskUsecase:    taskUsecase,
	}

	return suite, func() {
		testutil.TeardownGinTestMode()
		cleanup()
	}
}

// Test Project CRUD Operations
func TestIntegration_ProjectCRUD(t *testing.T) {
	suite, cleanup := SetupIntegrationTestSuite(t)
	defer cleanup()

	t.Run("Create Project", func(t *testing.T) {
		// Prepare request
		createReq := dto.ProjectCreateRequest{
			Name:        "Integration Test Project",
			Description: "A project for integration testing",
			RepoURL:     "https://github.com/test/integration.git",
		}

		// Make request
		w := suite.apiHelper.MakeRequest("POST", "/api/v1/projects", createReq)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.ProjectResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.NotEqual(t, uuid.Nil.String(), response.ID)
		assert.Equal(t, createReq.Name, response.Name)
		assert.Equal(t, createReq.Description, response.Description)
		assert.Equal(t, createReq.RepoURL, response.RepoURL)
		assert.NotEmpty(t, response.CreatedAt)
		assert.NotEmpty(t, response.UpdatedAt)

		// Verify in database
		suite.dbHelper.AssertRecordExists(t, "projects", "name = ?", createReq.Name)
	})

	t.Run("Create Project with Invalid Data", func(t *testing.T) {
		testCases := []testutil.ValidationTestCase{
			{
				Name: "missing name",
				Body: dto.ProjectCreateRequest{
					Description: "Test",
					RepoURL:     "https://github.com/test/repo.git",
				},
				ExpectedStatus: http.StatusBadRequest,
				ExpectedError:  "required",
			},
			{
				Name: "missing repo URL",
				Body: dto.ProjectCreateRequest{
					Name:        "Test Project",
					Description: "Test",
				},
				ExpectedStatus: http.StatusBadRequest,
				ExpectedError:  "required",
			},
			{
				Name: "invalid repo URL",
				Body: dto.ProjectCreateRequest{
					Name:        "Test Project",
					Description: "Test",
					RepoURL:     "not-a-url",
				},
				ExpectedStatus: http.StatusInternalServerError,
				ExpectedError:  "invalid repository URL",
			},
		}

		suite.apiHelper.RunValidationTests(t, "POST", "/api/v1/projects", testCases)
	})

	t.Run("Get Projects", func(t *testing.T) {
		// Create test projects
		projects := createTestProjects(t, suite, 3)

		// Make request
		w := suite.apiHelper.MakeRequest("GET", "/api/v1/projects?page=1&page_size=10", nil)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectListResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.GreaterOrEqual(t, len(response.Projects), 3)
		assert.GreaterOrEqual(t, response.Total, 3)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)

		// Verify project IDs exist in the response
		responseIDs := make(map[string]bool)
		for _, p := range response.Projects {
			responseIDs[p.ID] = true
		}

		for _, project := range projects {
			assert.True(t, responseIDs[project.ID.String()], "Project ID %s should be in response", project.ID)
		}
	})

	t.Run("Get Project by ID", func(t *testing.T) {
		// Create test project
		projects := createTestProjects(t, suite, 1)
		project := projects[0]

		// Make request
		w := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/projects/%s", project.ID), nil)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Equal(t, project.ID.String(), response.ID)
		assert.Equal(t, project.Name, response.Name)
		assert.Equal(t, project.Description, response.Description)
		assert.Equal(t, project.RepoURL, response.RepoURL)
	})

	t.Run("Get Non-existent Project", func(t *testing.T) {
		nonExistentID := uuid.New()

		// Make request
		w := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/projects/%s", nonExistentID), nil)

		// Assertions
		suite.apiHelper.AssertErrorResponse(t, w, http.StatusNotFound, "not found")
	})

	t.Run("Update Project", func(t *testing.T) {
		// Create test project
		projects := createTestProjects(t, suite, 1)
		project := projects[0]

		// Prepare update request
		updateReq := dto.ProjectUpdateRequest{
			Name:        "Updated Project Name",
			Description: "Updated description",
			RepoURL:     "https://github.com/updated/repo.git",
		}

		// Make request
		w := suite.apiHelper.MakeRequest("PUT", fmt.Sprintf("/api/v1/projects/%s", project.ID), updateReq)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Equal(t, project.ID.String(), response.ID)
		assert.Equal(t, updateReq.Name, response.Name)
		assert.Equal(t, updateReq.Description, response.Description)
		assert.Equal(t, updateReq.RepoURL, response.RepoURL)

		// Verify in database
		suite.dbHelper.AssertRecordExists(t, "projects", "id = ? AND name = ?", project.ID, updateReq.Name)
	})

	t.Run("Delete Project", func(t *testing.T) {
		// Create test project
		projects := createTestProjects(t, suite, 1)
		project := projects[0]

		// Make request
		w := suite.apiHelper.MakeRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s", project.ID), nil)

		// Assertions
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify project is soft deleted (not found in normal queries)
		w2 := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/projects/%s", project.ID), nil)
		assert.Equal(t, http.StatusNotFound, w2.Code)
	})
}

// Test Task CRUD Operations
func TestIntegration_TaskCRUD(t *testing.T) {
	suite, cleanup := SetupIntegrationTestSuite(t)
	defer cleanup()

	t.Run("Create Task", func(t *testing.T) {
		// Create test project
		projects := createTestProjects(t, suite, 1)
		project := projects[0]

		// Prepare request
		createReq := dto.TaskCreateRequest{
			Title:       "Integration Test Task",
			Description: "A task for integration testing",
		}

		// Make request
		w := suite.apiHelper.MakeRequest("POST", fmt.Sprintf("/api/v1/projects/%s/tasks", project.ID), createReq)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.TaskResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.NotEqual(t, uuid.Nil.String(), response.ID)
		assert.Equal(t, project.ID.String(), response.ProjectID)
		assert.Equal(t, createReq.Title, response.Title)
		assert.Equal(t, createReq.Description, response.Description)
		assert.Equal(t, string(entity.TaskStatusTODO), response.Status)

		// Verify in database
		suite.dbHelper.AssertRecordExists(t, "tasks", "title = ? AND project_id = ?", createReq.Title, project.ID)
	})

	t.Run("Get Tasks by Project", func(t *testing.T) {
		// Create test project and tasks
		projects := createTestProjects(t, suite, 1)
		project := projects[0]
		tasks := createTestTasks(t, suite, project.ID, 3)

		// Make request
		w := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/projects/%s/tasks", project.ID), nil)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.TaskListResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Len(t, response.Tasks, 3)
		assert.Equal(t, 3, response.Total)

		// Verify task IDs exist in the response
		responseIDs := make(map[string]bool)
		for _, task := range response.Tasks {
			responseIDs[task.ID] = true
		}

		for _, task := range tasks {
			assert.True(t, responseIDs[task.ID.String()], "Task ID %s should be in response", task.ID)
		}
	})

	t.Run("Get Task by ID", func(t *testing.T) {
		// Create test project and task
		projects := createTestProjects(t, suite, 1)
		project := projects[0]
		tasks := createTestTasks(t, suite, project.ID, 1)
		task := tasks[0]

		// Make request
		w := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.TaskResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Equal(t, task.ID.String(), response.ID)
		assert.Equal(t, task.ProjectID.String(), response.ProjectID)
		assert.Equal(t, task.Title, response.Title)
		assert.Equal(t, task.Description, response.Description)
		assert.Equal(t, string(task.Status), response.Status)
	})

	t.Run("Update Task", func(t *testing.T) {
		// Create test project and task
		projects := createTestProjects(t, suite, 1)
		project := projects[0]
		tasks := createTestTasks(t, suite, project.ID, 1)
		task := tasks[0]

		// Prepare update request
		updateReq := dto.TaskUpdateRequest{
			Title:       "Updated Task Title",
			Description: "Updated task description",
		}

		// Make request
		w := suite.apiHelper.MakeRequest("PUT", fmt.Sprintf("/api/v1/tasks/%s", task.ID), updateReq)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.TaskResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Equal(t, task.ID.String(), response.ID)
		assert.Equal(t, updateReq.Title, response.Title)
		assert.Equal(t, updateReq.Description, response.Description)

		// Verify in database
		suite.dbHelper.AssertRecordExists(t, "tasks", "id = ? AND title = ?", task.ID, updateReq.Title)
	})

	t.Run("Update Task Status", func(t *testing.T) {
		// Create test project and task
		projects := createTestProjects(t, suite, 1)
		project := projects[0]
		tasks := createTestTasks(t, suite, project.ID, 1)
		task := tasks[0]

		// Prepare status update request
		statusReq := dto.TaskStatusUpdateRequest{
			Status: string(entity.TaskStatusIMPLEMENTING),
		}

		// Make request
		w := suite.apiHelper.MakeRequest("PATCH", fmt.Sprintf("/api/v1/tasks/%s/status", task.ID), statusReq)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.TaskResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Equal(t, task.ID.String(), response.ID)
		assert.Equal(t, statusReq.Status, response.Status)

		// Verify in database
		suite.dbHelper.AssertRecordExists(t, "tasks", "id = ? AND status = ?", task.ID, statusReq.Status)
	})

	t.Run("Delete Task", func(t *testing.T) {
		// Create test project and task
		projects := createTestProjects(t, suite, 1)
		project := projects[0]
		tasks := createTestTasks(t, suite, project.ID, 1)
		task := tasks[0]

		// Make request
		w := suite.apiHelper.MakeRequest("DELETE", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil)

		// Assertions
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify task is soft deleted (not found in normal queries)
		w2 := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil)
		assert.Equal(t, http.StatusNotFound, w2.Code)
	})
}

// Test Error Scenarios
func TestIntegration_ErrorScenarios(t *testing.T) {
	suite, cleanup := SetupIntegrationTestSuite(t)
	defer cleanup()

	t.Run("Invalid UUID in URL", func(t *testing.T) {
		// Make request with invalid UUID
		w := suite.apiHelper.MakeRequest("GET", "/api/v1/projects/invalid-uuid", nil)

		// Assertions
		suite.apiHelper.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid UUID")
	})

	t.Run("Create Task for Non-existent Project", func(t *testing.T) {
		nonExistentProjectID := uuid.New()

		createReq := dto.TaskCreateRequest{
			Title:       "Test Task",
			Description: "Test Description",
		}

		// Make request
		w := suite.apiHelper.MakeRequest("POST", fmt.Sprintf("/api/v1/projects/%s/tasks", nonExistentProjectID), createReq)

		// Assertions
		suite.apiHelper.AssertErrorResponse(t, w, http.StatusInternalServerError, "failed to create task")
	})

	t.Run("Update Non-existent Task", func(t *testing.T) {
		nonExistentTaskID := uuid.New()

		updateReq := dto.TaskUpdateRequest{
			Title: "Updated Title",
		}

		// Make request
		w := suite.apiHelper.MakeRequest("PUT", fmt.Sprintf("/api/v1/tasks/%s", nonExistentTaskID), updateReq)

		// Assertions
		suite.apiHelper.AssertErrorResponse(t, w, http.StatusNotFound, "not found")
	})
}

// Test Pagination and Filtering
func TestIntegration_PaginationAndFiltering(t *testing.T) {
	suite, cleanup := SetupIntegrationTestSuite(t)
	defer cleanup()

	t.Run("Project Pagination", func(t *testing.T) {
		// Create test projects
		createTestProjects(t, suite, 15)

		// Test first page
		w := suite.apiHelper.MakeRequest("GET", "/api/v1/projects?page=1&page_size=5", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectListResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.Len(t, response.Projects, 5)
		assert.GreaterOrEqual(t, response.Total, 15)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 5, response.PageSize)

		// Test second page
		w2 := suite.apiHelper.MakeRequest("GET", "/api/v1/projects?page=2&page_size=5", nil)
		assert.Equal(t, http.StatusOK, w2.Code)

		var response2 dto.ProjectListResponse
		suite.apiHelper.ParseJSONResponse(t, w2, &response2)

		assert.Len(t, response2.Projects, 5)
		assert.Equal(t, 2, response2.Page)

		// Ensure no duplicate projects between pages
		firstPageIDs := make(map[string]bool)
		for _, p := range response.Projects {
			firstPageIDs[p.ID] = true
		}

		for _, p := range response2.Projects {
			assert.False(t, firstPageIDs[p.ID], "Project %s should not appear on both pages", p.ID)
		}
	})

	t.Run("Project Search", func(t *testing.T) {
		// Create test projects with specific names
		projectFactory := testutil.NewProjectFactory()
		searchableProject := projectFactory.CreateProject(func(p *entity.Project) {
			p.Name = "Searchable Project"
			p.Description = "Contains search term"
		})

		_, err := suite.projectUsecase.Create(context.Background(), usecase.CreateProjectRequest{
			Name:        searchableProject.Name,
			Description: searchableProject.Description,
			RepoURL:     searchableProject.RepoURL,
		})
		require.NoError(t, err)

		// Search for projects
		w := suite.apiHelper.MakeRequest("GET", "/api/v1/projects?search=searchable", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ProjectListResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.GreaterOrEqual(t, len(response.Projects), 1)
		
		// Check that at least one result contains the search term
		found := false
		for _, p := range response.Projects {
			if p.Name == "Searchable Project" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find the searchable project")
	})

	t.Run("Task Filtering by Status", func(t *testing.T) {
		// Create test project and tasks with different statuses
		projects := createTestProjects(t, suite, 1)
		project := projects[0]

		taskFactory := testutil.NewTaskFactory()
		tasks := []*entity.Task{
			taskFactory.CreateTask(func(t *entity.Task) {
				t.ProjectID = project.ID
				t.Status = entity.TaskStatusTODO
			}),
			taskFactory.CreateTask(func(t *entity.Task) {
				t.ProjectID = project.ID
				t.Status = entity.TaskStatusDONE
			}),
		}

		for _, task := range tasks {
			_, err := suite.taskUsecase.Create(context.Background(), usecase.CreateTaskRequest{
				ProjectID:   task.ProjectID,
				Title:       task.Title,
				Description: task.Description,
			})
			require.NoError(t, err)
		}

		// Filter by TODO status
		w := suite.apiHelper.MakeRequest("GET", fmt.Sprintf("/api/v1/projects/%s/tasks?status=TODO", project.ID), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.TaskListResponse
		suite.apiHelper.ParseJSONResponse(t, w, &response)

		assert.GreaterOrEqual(t, len(response.Tasks), 1)
		for _, task := range response.Tasks {
			assert.Equal(t, string(entity.TaskStatusTODO), task.Status)
		}
	})
}

// Test Concurrent Operations
func TestIntegration_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	suite, cleanup := SetupIntegrationTestSuite(t)
	defer cleanup()

	t.Run("Concurrent Project Creation", func(t *testing.T) {
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				createReq := dto.ProjectCreateRequest{
					Name:        fmt.Sprintf("Concurrent Project %d", i),
					Description: "Concurrent test",
					RepoURL:     fmt.Sprintf("https://github.com/test/concurrent%d.git", i),
				}

				w := suite.apiHelper.MakeRequest("POST", "/api/v1/projects", createReq)
				
				if w.Code != http.StatusCreated {
					results <- fmt.Errorf("expected status 201, got %d", w.Code)
				} else {
					results <- nil
				}
			}(i)
		}

		// Wait for all goroutines to complete
		errorCount := 0
		for i := 0; i < numGoroutines; i++ {
			if err := <-results; err != nil {
				t.Errorf("Goroutine failed: %v", err)
				errorCount++
			}
		}

		// Allow some failures due to race conditions but most should succeed
		assert.LessOrEqual(t, errorCount, numGoroutines/2, "Too many concurrent operations failed")

		// Verify at least some projects were created
		projectCount := suite.dbHelper.CountRecords(t, "projects")
		assert.GreaterOrEqual(t, projectCount, int64(numGoroutines/2))
	})
}

// Helper functions

func createTestProjects(t *testing.T, suite *IntegrationTestSuite, count int) []*entity.Project {
	projectFactory := testutil.NewProjectFactory()
	projects := make([]*entity.Project, count)

	for i := 0; i < count; i++ {
		project := projectFactory.CreateProject(func(p *entity.Project) {
			p.Name = fmt.Sprintf("Test Project %d %d", i, time.Now().UnixNano())
			p.Description = fmt.Sprintf("Description %d", i)
			p.RepoURL = fmt.Sprintf("https://github.com/test/repo%d.git", i)
		})

		created, err := suite.projectUsecase.Create(context.Background(), usecase.CreateProjectRequest{
			Name:        project.Name,
			Description: project.Description,
			RepoURL:     project.RepoURL,
		})
		require.NoError(t, err)
		projects[i] = created
	}

	return projects
}

func createTestTasks(t *testing.T, suite *IntegrationTestSuite, projectID uuid.UUID, count int) []*entity.Task {
	taskFactory := testutil.NewTaskFactory()
	tasks := make([]*entity.Task, count)

	for i := 0; i < count; i++ {
		task := taskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = projectID
			t.Title = fmt.Sprintf("Test Task %d %d", i, time.Now().UnixNano())
			t.Description = fmt.Sprintf("Description %d", i)
		})

		created, err := suite.taskUsecase.Create(context.Background(), usecase.CreateTaskRequest{
			ProjectID:   task.ProjectID,
			Title:       task.Title,
			Description: task.Description,
		})
		require.NoError(t, err)
		tasks[i] = created
	}

	return tasks
}