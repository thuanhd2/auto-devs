package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/mocks"
	"github.com/auto-devs/auto-devs/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)







// Test cases for enhanced task management features

func TestTaskUsecase_CreateWithEnhancedFeatures(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	projectID := uuid.New()
	parentTaskID := uuid.New()

	// Test creating task with all enhanced features
	req := usecase.CreateTaskRequest{
		ProjectID:      projectID,
		Title:          "Enhanced Task",
		Description:    "Task with all new features",
		Priority:       entity.TaskPriorityHigh,
		EstimatedHours: &[]float64{8.5}[0],
		Tags:           []string{"frontend", "urgent"},
		ParentTaskID:   &parentTaskID,
		AssignedTo:     &[]string{"user123"}[0],
		DueDate:        &[]time.Time{time.Now().AddDate(0, 0, 7)}[0],
		BranchName:     &[]string{"feature/enhanced-task"}[0],
		PullRequest:    &[]string{"PR-123"}[0],
	}

	// Setup mocks
	mockTaskRepo.On("ValidateProjectExists", ctx, projectID).Return(true, nil)
	mockTaskRepo.On("CheckDuplicateTitle", ctx, projectID, req.Title, (*uuid.UUID)(nil)).Return(false, nil)
	mockTaskRepo.On("ValidateTaskExists", ctx, parentTaskID).Return(true, nil)
	mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(nil)
	mockProjectRepo.On("GetByID", ctx, projectID).Return(&entity.Project{ID: projectID, Name: "Test Project"}, nil)
	mockNotificationUsecase.On("SendTaskCreatedNotification", ctx, mock.AnythingOfType("*entity.Task"), mock.AnythingOfType("*entity.Project")).Return(nil)

	// Execute
	task, err := uc.Create(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, req.Title, task.Title)
	assert.Equal(t, req.Priority, task.Priority)
	assert.Equal(t, req.EstimatedHours, task.EstimatedHours)
	assert.Equal(t, req.Tags, task.Tags)
	assert.Equal(t, req.ParentTaskID, task.ParentTaskID)
	assert.Equal(t, req.AssignedTo, task.AssignedTo)
	assert.Equal(t, req.DueDate, task.DueDate)
	assert.Equal(t, req.BranchName, task.BranchName)
	assert.Equal(t, req.PullRequest, task.PullRequest)

	mockTaskRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
}

func TestTaskUsecase_SearchTasks(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	query := "search term"
	projectID := uuid.New()

	expectedResults := []*entity.TaskSearchResult{
		{
			Task:    &entity.Task{ID: uuid.New(), Title: "Task 1"},
			Score:   0.9,
			Matched: "title",
		},
		{
			Task:    &entity.Task{ID: uuid.New(), Title: "Task 2"},
			Score:   0.7,
			Matched: "description",
		},
	}

	mockTaskRepo.On("SearchTasks", ctx, query, &projectID).Return(expectedResults, nil)

	results, err := uc.SearchTasks(ctx, query, &projectID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_BulkOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	taskIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Test bulk archive
	mockTaskRepo.On("BulkArchive", ctx, taskIDs).Return(nil)
	err := uc.BulkArchive(ctx, taskIDs)
	assert.NoError(t, err)

	// Test bulk unarchive
	mockTaskRepo.On("BulkUnarchive", ctx, taskIDs).Return(nil)
	err = usecase.BulkUnarchive(ctx, taskIDs)
	assert.NoError(t, err)

	// Test bulk update priority
	priority := entity.TaskPriorityHigh
	mockTaskRepo.On("BulkUpdatePriority", ctx, taskIDs, priority).Return(nil)
	err = usecase.BulkUpdatePriority(ctx, taskIDs, priority)
	assert.NoError(t, err)

	// Test bulk assign
	assignedTo := "user123"
	mockTaskRepo.On("BulkAssign", ctx, taskIDs, assignedTo).Return(nil)
	err = usecase.BulkAssign(ctx, taskIDs, assignedTo)
	assert.NoError(t, err)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_TemplateOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	projectID := uuid.New()
	createdBy := "user123"

	// Test create template
	req := usecase.CreateTemplateRequest{
		ProjectID:      projectID,
		Name:           "Bug Fix Template",
		Description:    "Template for bug fixes",
		Title:          "Fix {bug_description}",
		Priority:       entity.TaskPriorityHigh,
		EstimatedHours: &[]float64{4.0}[0],
		Tags:           []string{"bug", "fix"},
		IsGlobal:       true,
		CreatedBy:      createdBy,
	}

	mockTaskRepo.On("ValidateProjectExists", ctx, projectID).Return(true, nil)
	mockTaskRepo.On("CreateTemplate", ctx, mock.AnythingOfType("*entity.TaskTemplate")).Return(nil)

	template, err := uc.CreateTemplate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, req.Name, template.Name)
	assert.Equal(t, req.Title, template.Title)
	assert.Equal(t, req.Priority, template.Priority)
	assert.Equal(t, req.IsGlobal, template.IsGlobal)

	mockTaskRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
}

func TestTaskUsecase_DependencyOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	taskID := uuid.New()
	dependsOnTaskID := uuid.New()

	// Test add dependency
	mockTaskRepo.On("ValidateTaskExists", ctx, taskID).Return(true, nil)
	mockTaskRepo.On("ValidateTaskExists", ctx, dependsOnTaskID).Return(true, nil)
	mockTaskRepo.On("AddDependency", ctx, taskID, dependsOnTaskID, "blocks").Return(nil)

	err := uc.AddDependency(ctx, taskID, dependsOnTaskID, "blocks")
	assert.NoError(t, err)

	// Test get dependencies
	expectedDependencies := []*entity.TaskDependency{
		{
			ID:              uuid.New(),
			TaskID:          taskID,
			DependsOnTaskID: dependsOnTaskID,
			DependencyType:  "blocks",
		},
	}

	mockTaskRepo.On("GetDependencies", ctx, taskID).Return(expectedDependencies, nil)

	dependencies, err := uc.GetDependencies(ctx, taskID)
	assert.NoError(t, err)
	assert.Equal(t, expectedDependencies, dependencies)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_CommentOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	taskID := uuid.New()
	commentText := "This is a test comment"
	createdBy := "user123"

	// Test add comment
	req := usecase.AddCommentRequest{
		TaskID:    taskID,
		Comment:   commentText,
		CreatedBy: createdBy,
	}

	mockTaskRepo.On("ValidateTaskExists", ctx, taskID).Return(true, nil)
	mockTaskRepo.On("AddComment", ctx, mock.AnythingOfType("*entity.TaskComment")).Return(nil)

	comment, err := uc.AddComment(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, taskID, comment.TaskID)
	assert.Equal(t, commentText, comment.Comment)
	assert.Equal(t, createdBy, comment.CreatedBy)

	// Test get comments
	expectedComments := []*entity.TaskComment{
		{
			ID:        uuid.New(),
			TaskID:    taskID,
			Comment:   commentText,
			CreatedBy: createdBy,
		},
	}

	mockTaskRepo.On("GetComments", ctx, taskID).Return(expectedComments, nil)

	comments, err := uc.GetComments(ctx, taskID)
	assert.NoError(t, err)
	assert.Equal(t, expectedComments, comments)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(mocks.MockTaskRepository)
	mockProjectRepo := new(mocks.MockProjectRepository)
	mockNotificationUsecase := new(mocks.MockNotificationUsecase)

	uc := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	// Test empty search query
	_, err := uc.SearchTasks(ctx, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search query cannot be empty")

	// Test invalid priority
	_, err = usecase.GetTasksByPriority(ctx, "INVALID_PRIORITY")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid priority")

	// Test empty tags
	_, err = usecase.GetTasksByTags(ctx, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one tag must be provided")

	// Test empty task IDs for bulk operations
	err = usecase.BulkDelete(ctx, []uuid.UUID{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no task IDs provided")

	// Test empty assigned_to for bulk assign
	err = usecase.BulkAssign(ctx, []uuid.UUID{uuid.New()}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assigned_to cannot be empty")
}
