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
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

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
	task, err := taskUsecase.Create(ctx, req)

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
}

func TestTaskUsecase_SearchTasks(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

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

	results, err := taskUsecase.SearchTasks(ctx, query, &projectID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
}

func TestTaskUsecase_BulkOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

	taskIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Test bulk archive
	mockTaskRepo.On("BulkArchive", ctx, taskIDs).Return(nil)
	err := taskUsecase.BulkArchive(ctx, taskIDs)
	assert.NoError(t, err)

	// Test bulk unarchive
	mockTaskRepo.On("BulkUnarchive", ctx, taskIDs).Return(nil)
	err = taskUsecase.BulkUnarchive(ctx, taskIDs)
	assert.NoError(t, err)

	// Test bulk delete
	mockTaskRepo.On("BulkDelete", ctx, taskIDs).Return(nil)
	err = taskUsecase.BulkDelete(ctx, taskIDs)
	assert.NoError(t, err)

	// Test bulk update priority
	mockTaskRepo.On("BulkUpdatePriority", ctx, taskIDs, entity.TaskPriorityHigh).Return(nil)
	err = taskUsecase.BulkUpdatePriority(ctx, taskIDs, entity.TaskPriorityHigh)
	assert.NoError(t, err)

	// Test bulk assign
	assignedTo := "user123"
	mockTaskRepo.On("BulkAssign", ctx, taskIDs, assignedTo).Return(nil)
	err = taskUsecase.BulkAssign(ctx, taskIDs, assignedTo)
	assert.NoError(t, err)
}

func TestTaskUsecase_TemplateOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

	projectID := uuid.New()

	// Test create template
	req := usecase.CreateTemplateRequest{
		ProjectID:      projectID,
		Name:           "Bug Template",
		Description:    "Template for bug reports",
		Title:          "Bug: {title}",
		Priority:       entity.TaskPriorityMedium,
		EstimatedHours: &[]float64{2.0}[0],
		Tags:           []string{"bug", "template"},
		IsGlobal:       false,
		CreatedBy:      "user123",
	}

	mockTaskRepo.On("ValidateProjectExists", ctx, req.ProjectID).Return(true, nil)
	mockTaskRepo.On("CreateTemplate", ctx, mock.AnythingOfType("*entity.TaskTemplate")).Return(nil)

	template, err := taskUsecase.CreateTemplate(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, req.Name, template.Name)
	assert.Equal(t, req.Priority, template.Priority)
	assert.Equal(t, req.IsGlobal, template.IsGlobal)
}

func TestTaskUsecase_DependencyOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

	taskID := uuid.New()
	dependsOnTaskID := uuid.New()

	// Test add dependency
	mockTaskRepo.On("ValidateTaskExists", ctx, taskID).Return(true, nil)
	mockTaskRepo.On("ValidateTaskExists", ctx, dependsOnTaskID).Return(true, nil)
	mockTaskRepo.On("AddDependency", ctx, taskID, dependsOnTaskID, "blocks").Return(nil)
	err := taskUsecase.AddDependency(ctx, taskID, dependsOnTaskID, "blocks")
	assert.NoError(t, err)

	// Test get dependencies
	expectedDependencies := []*entity.TaskDependency{
		{TaskID: taskID, DependsOnTaskID: dependsOnTaskID, DependencyType: "blocks"},
	}
	mockTaskRepo.On("GetDependencies", ctx, taskID).Return(expectedDependencies, nil)

	dependencies, err := taskUsecase.GetDependencies(ctx, taskID)
	assert.NoError(t, err)
	assert.Equal(t, expectedDependencies, dependencies)
}

func TestTaskUsecase_CommentOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

	taskID := uuid.New()
	commentID := uuid.New()

	// Test add comment
	req := usecase.AddCommentRequest{
		TaskID:    taskID,
		Comment:   "This is a test comment",
		CreatedBy: "user123",
	}

	mockTaskRepo.On("ValidateTaskExists", ctx, taskID).Return(true, nil)
	mockTaskRepo.On("AddComment", ctx, mock.AnythingOfType("*entity.TaskComment")).Return(nil)

	comment, err := taskUsecase.AddComment(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, req.Comment, comment.Comment)
	assert.Equal(t, req.CreatedBy, comment.CreatedBy)

	// Test get comments
	expectedComments := []*entity.TaskComment{
		{ID: commentID, TaskID: taskID, Comment: req.Comment, CreatedBy: req.CreatedBy},
	}
	mockTaskRepo.On("GetComments", ctx, taskID).Return(expectedComments, nil)

	comments, err := taskUsecase.GetComments(ctx, taskID)
	assert.NoError(t, err)
	assert.Equal(t, expectedComments, comments)
}

func TestTaskUsecase_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := mocks.NewMockTaskRepository(t)
	mockProjectRepo := mocks.NewMockProjectRepository(t)
	mockNotificationUsecase := mocks.NewMockNotificationUsecase(t)
	mockWorktreeUsecase := mocks.NewMockWorktreeUsecase(t)

	taskUsecase := usecase.NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase, mockWorktreeUsecase)

	projectID := uuid.New()

	// Test duplicate title validation
	req := usecase.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "Duplicate Task",
	}

	mockTaskRepo.On("ValidateProjectExists", ctx, projectID).Return(true, nil)
	mockTaskRepo.On("CheckDuplicateTitle", ctx, projectID, req.Title, (*uuid.UUID)(nil)).Return(true, nil)

	_, err := taskUsecase.Create(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists in this project")
}
