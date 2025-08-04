package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of TaskRepository
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TaskStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateStatusWithHistory(ctx context.Context, id uuid.UUID, status entity.TaskStatus, changedBy *string, reason *string) error {
	args := m.Called(ctx, id, status, changedBy, reason)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByStatus(ctx context.Context, status entity.TaskStatus) ([]*entity.Task, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByStatuses(ctx context.Context, statuses []entity.TaskStatus) ([]*entity.Task, error) {
	args := m.Called(ctx, statuses)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.TaskStatus, changedBy *string) error {
	args := m.Called(ctx, ids, status, changedBy)
	return args.Error(0)
}

func (m *MockTaskRepository) GetStatusHistory(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskStatusHistory, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]*entity.TaskStatusHistory), args.Error(1)
}

func (m *MockTaskRepository) GetStatusAnalytics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatusAnalytics, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(*entity.TaskStatusAnalytics), args.Error(1)
}

func (m *MockTaskRepository) GetTasksWithFilters(ctx context.Context, filters entity.TaskFilters) ([]*entity.Task, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) SearchTasks(ctx context.Context, query string, projectID *uuid.UUID) ([]*entity.TaskSearchResult, error) {
	args := m.Called(ctx, query, projectID)
	return args.Get(0).([]*entity.TaskSearchResult), args.Error(1)
}

func (m *MockTaskRepository) GetTasksByPriority(ctx context.Context, priority entity.TaskPriority) ([]*entity.Task, error) {
	args := m.Called(ctx, priority)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTasksByTags(ctx context.Context, tags []string) ([]*entity.Task, error) {
	args := m.Called(ctx, tags)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetArchivedTasks(ctx context.Context, projectID *uuid.UUID) ([]*entity.Task, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTasksWithSubtasks(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetSubtasks(ctx context.Context, parentTaskID uuid.UUID) ([]*entity.Task, error) {
	args := m.Called(ctx, parentTaskID)
	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetParentTask(ctx context.Context, taskID uuid.UUID) (*entity.Task, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error {
	args := m.Called(ctx, taskID, parentTaskID)
	return args.Error(0)
}

func (m *MockTaskRepository) BulkDelete(ctx context.Context, taskIDs []uuid.UUID) error {
	args := m.Called(ctx, taskIDs)
	return args.Error(0)
}

func (m *MockTaskRepository) BulkArchive(ctx context.Context, taskIDs []uuid.UUID) error {
	args := m.Called(ctx, taskIDs)
	return args.Error(0)
}

func (m *MockTaskRepository) BulkUnarchive(ctx context.Context, taskIDs []uuid.UUID) error {
	args := m.Called(ctx, taskIDs)
	return args.Error(0)
}

func (m *MockTaskRepository) BulkUpdatePriority(ctx context.Context, taskIDs []uuid.UUID, priority entity.TaskPriority) error {
	args := m.Called(ctx, taskIDs, priority)
	return args.Error(0)
}

func (m *MockTaskRepository) BulkAssign(ctx context.Context, taskIDs []uuid.UUID, assignedTo string) error {
	args := m.Called(ctx, taskIDs, assignedTo)
	return args.Error(0)
}

func (m *MockTaskRepository) CreateTemplate(ctx context.Context, template *entity.TaskTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTemplates(ctx context.Context, projectID uuid.UUID, includeGlobal bool) ([]*entity.TaskTemplate, error) {
	args := m.Called(ctx, projectID, includeGlobal)
	return args.Get(0).([]*entity.TaskTemplate), args.Error(1)
}

func (m *MockTaskRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*entity.TaskTemplate, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.TaskTemplate), args.Error(1)
}

func (m *MockTaskRepository) UpdateTemplate(ctx context.Context, template *entity.TaskTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, projectID uuid.UUID, createdBy string) (*entity.Task, error) {
	args := m.Called(ctx, templateID, projectID, createdBy)
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) GetAuditLogs(ctx context.Context, taskID uuid.UUID, limit *int) ([]*entity.TaskAuditLog, error) {
	args := m.Called(ctx, taskID, limit)
	return args.Get(0).([]*entity.TaskAuditLog), args.Error(1)
}

func (m *MockTaskRepository) CreateAuditLog(ctx context.Context, auditLog *entity.TaskAuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (*entity.TaskStatistics, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(*entity.TaskStatistics), args.Error(1)
}

func (m *MockTaskRepository) AddDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, dependencyType string) error {
	args := m.Called(ctx, taskID, dependsOnTaskID, dependencyType)
	return args.Error(0)
}

func (m *MockTaskRepository) RemoveDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) error {
	args := m.Called(ctx, taskID, dependsOnTaskID)
	return args.Error(0)
}

func (m *MockTaskRepository) GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]*entity.TaskDependency), args.Error(1)
}

func (m *MockTaskRepository) GetDependents(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskDependency, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]*entity.TaskDependency), args.Error(1)
}

func (m *MockTaskRepository) AddComment(ctx context.Context, comment *entity.TaskComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockTaskRepository) GetComments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskComment, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]*entity.TaskComment), args.Error(1)
}

func (m *MockTaskRepository) UpdateComment(ctx context.Context, comment *entity.TaskComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	args := m.Called(ctx, commentID)
	return args.Error(0)
}

func (m *MockTaskRepository) AddAttachment(ctx context.Context, attachment *entity.TaskAttachment) error {
	args := m.Called(ctx, attachment)
	return args.Error(0)
}

func (m *MockTaskRepository) GetAttachments(ctx context.Context, taskID uuid.UUID) ([]*entity.TaskAttachment, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]*entity.TaskAttachment), args.Error(1)
}

func (m *MockTaskRepository) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	args := m.Called(ctx, attachmentID)
	return args.Error(0)
}

func (m *MockTaskRepository) ExportTasks(ctx context.Context, filters entity.TaskFilters, format entity.TaskExportFormat) ([]byte, error) {
	args := m.Called(ctx, filters, format)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockTaskRepository) CheckDuplicateTitle(ctx context.Context, projectID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, projectID, title, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTaskRepository) ValidateTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	args := m.Called(ctx, taskID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTaskRepository) ValidateProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
	args := m.Called(ctx, projectID)
	return args.Bool(0), args.Error(1)
}

// MockProjectRepository is a mock implementation of ProjectRepository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectRepository) GetAll(ctx context.Context) ([]*entity.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Project), args.Error(1)
}

func (m *MockProjectRepository) GetAllWithParams(ctx context.Context, params repository.GetProjectsParams) ([]*entity.Project, int, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*entity.Project), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) GetWithTaskCount(ctx context.Context, id uuid.UUID) (*repository.ProjectWithTaskCount, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*repository.ProjectWithTaskCount), args.Error(1)
}

func (m *MockProjectRepository) GetTaskStatistics(ctx context.Context, projectID uuid.UUID) (map[entity.TaskStatus]int, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(map[entity.TaskStatus]int), args.Error(1)
}

func (m *MockProjectRepository) GetLastActivityAt(ctx context.Context, projectID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockProjectRepository) Archive(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) CheckNameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepository) GetSettings(ctx context.Context, projectID uuid.UUID) (*entity.ProjectSettings, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(*entity.ProjectSettings), args.Error(1)
}

func (m *MockProjectRepository) CreateSettings(ctx context.Context, settings *entity.ProjectSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockProjectRepository) UpdateSettings(ctx context.Context, settings *entity.ProjectSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

// MockNotificationUsecase is a mock implementation of NotificationUsecase
type MockNotificationUsecase struct {
	mock.Mock
}

func (m *MockNotificationUsecase) SendTaskStatusChangeNotification(ctx context.Context, data entity.TaskStatusChangeNotificationData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockNotificationUsecase) SendTaskCreatedNotification(ctx context.Context, task *entity.Task, project *entity.Project) error {
	args := m.Called(ctx, task, project)
	return args.Error(0)
}

func (m *MockNotificationUsecase) RegisterHandler(notificationType entity.NotificationType, handler entity.NotificationHandler) error {
	args := m.Called(notificationType, handler)
	return args.Error(0)
}

func (m *MockNotificationUsecase) UnregisterHandler(notificationType entity.NotificationType) error {
	args := m.Called(notificationType)
	return args.Error(0)
}

// Test cases for enhanced task management features

func TestTaskUsecase_CreateWithEnhancedFeatures(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	projectID := uuid.New()
	parentTaskID := uuid.New()

	// Test creating task with all enhanced features
	req := CreateTaskRequest{
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
	task, err := usecase.Create(ctx, req)

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
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

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

	results, err := usecase.SearchTasks(ctx, query, &projectID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_BulkOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	taskIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Test bulk archive
	mockTaskRepo.On("BulkArchive", ctx, taskIDs).Return(nil)
	err := usecase.BulkArchive(ctx, taskIDs)
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
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	projectID := uuid.New()
	createdBy := "user123"

	// Test create template
	req := CreateTemplateRequest{
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

	template, err := usecase.CreateTemplate(ctx, req)

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
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	taskID := uuid.New()
	dependsOnTaskID := uuid.New()

	// Test add dependency
	mockTaskRepo.On("ValidateTaskExists", ctx, taskID).Return(true, nil)
	mockTaskRepo.On("ValidateTaskExists", ctx, dependsOnTaskID).Return(true, nil)
	mockTaskRepo.On("AddDependency", ctx, taskID, dependsOnTaskID, "blocks").Return(nil)

	err := usecase.AddDependency(ctx, taskID, dependsOnTaskID, "blocks")
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

	dependencies, err := usecase.GetDependencies(ctx, taskID)
	assert.NoError(t, err)
	assert.Equal(t, expectedDependencies, dependencies)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_CommentOperations(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	taskID := uuid.New()
	commentText := "This is a test comment"
	createdBy := "user123"

	// Test add comment
	req := AddCommentRequest{
		TaskID:    taskID,
		Comment:   commentText,
		CreatedBy: createdBy,
	}

	mockTaskRepo.On("ValidateTaskExists", ctx, taskID).Return(true, nil)
	mockTaskRepo.On("AddComment", ctx, mock.AnythingOfType("*entity.TaskComment")).Return(nil)

	comment, err := usecase.AddComment(ctx, req)

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

	comments, err := usecase.GetComments(ctx, taskID)
	assert.NoError(t, err)
	assert.Equal(t, expectedComments, comments)

	mockTaskRepo.AssertExpectations(t)
}

func TestTaskUsecase_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockNotificationUsecase := new(MockNotificationUsecase)

	usecase := NewTaskUsecase(mockTaskRepo, mockProjectRepo, mockNotificationUsecase)

	// Test empty search query
	_, err := usecase.SearchTasks(ctx, "", nil)
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
