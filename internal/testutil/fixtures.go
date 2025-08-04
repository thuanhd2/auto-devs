package testutil

import (
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/google/uuid"
)

// ProjectFactory provides methods to create test project entities
type ProjectFactory struct{}

// NewProjectFactory creates a new ProjectFactory
func NewProjectFactory() *ProjectFactory {
	return &ProjectFactory{}
}

// CreateProject creates a test project with default values
func (f *ProjectFactory) CreateProject(overrides ...func(*entity.Project)) *entity.Project {
	project := &entity.Project{
		ID:          uuid.New(),
		Name:        "Test Project",
		Description: "Test project description",
		RepoURL:     "https://github.com/test/repo.git",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, override := range overrides {
		override(project)
	}

	return project
}

// CreateMinimalProject creates a project with only required fields
func (f *ProjectFactory) CreateMinimalProject() *entity.Project {
	return &entity.Project{
		Name:    "Minimal Project",
		RepoURL: "https://github.com/test/minimal.git",
	}
}

// CreateProjectWithTasks creates a project with associated tasks
func (f *ProjectFactory) CreateProjectWithTasks(taskCount int) (*entity.Project, []*entity.Task) {
	project := f.CreateProject()
	taskFactory := NewTaskFactory()
	
	tasks := make([]*entity.Task, taskCount)
	for i := 0; i < taskCount; i++ {
		tasks[i] = taskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "Task " + string(rune(i+1))
		})
	}
	
	return project, tasks
}

// TaskFactory provides methods to create test task entities
type TaskFactory struct{}

// NewTaskFactory creates a new TaskFactory
func NewTaskFactory() *TaskFactory {
	return &TaskFactory{}
}

// CreateTask creates a test task with default values
func (f *TaskFactory) CreateTask(overrides ...func(*entity.Task)) *entity.Task {
	task := &entity.Task{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Test Task",
		Description: "Test task description",
		Status:      entity.TaskStatusTODO,
		// Priority field not available in entity.Task
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, override := range overrides {
		override(task)
	}

	return task
}

// CreateMinimalTask creates a task with only required fields
func (f *TaskFactory) CreateMinimalTask(projectID uuid.UUID) *entity.Task {
	return &entity.Task{
		ProjectID: projectID,
		Title:     "Minimal Task",
		Status:    entity.TaskStatusTODO,
	}
}

// CreateTasksWithDifferentStatuses creates tasks with different statuses for testing
func (f *TaskFactory) CreateTasksWithDifferentStatuses(projectID uuid.UUID) []*entity.Task {
	statuses := []entity.TaskStatus{
		entity.TaskStatusTODO,
		entity.TaskStatusPLANNING,
		entity.TaskStatusPLANREVIEWING,
		entity.TaskStatusIMPLEMENTING,
		entity.TaskStatusCODEREVIEWING,
		entity.TaskStatusDONE,
		entity.TaskStatusCANCELLED,
	}

	tasks := make([]*entity.Task, len(statuses))
	for i, status := range statuses {
		tasks[i] = f.CreateTask(func(t *entity.Task) {
			t.ProjectID = projectID
			t.Status = status
			t.Title = "Task with status " + string(status)
		})
	}

	return tasks
}

// CreateTasksWithDifferentPriorities creates tasks with different priorities for testing
// Note: Priority field is not available in current entity.Task, so we'll create tasks with different titles
func (f *TaskFactory) CreateTasksWithDifferentPriorities(projectID uuid.UUID) []*entity.Task {
	priorities := []string{
		"Low",
		"Medium", 
		"High",
		"Critical",
	}

	tasks := make([]*entity.Task, len(priorities))
	for i, priority := range priorities {
		tasks[i] = f.CreateTask(func(t *entity.Task) {
			t.ProjectID = projectID
			t.Title = "Task with priority " + priority
		})
	}

	return tasks
}

// AuditLogFactory provides methods to create test audit log entities
type AuditLogFactory struct{}

// NewAuditLogFactory creates a new AuditLogFactory
func NewAuditLogFactory() *AuditLogFactory {
	return &AuditLogFactory{}
}

// CreateAuditLog creates a test audit log with default values
func (f *AuditLogFactory) CreateAuditLog(overrides ...func(*entity.AuditLog)) *entity.AuditLog {
	auditLog := &entity.AuditLog{
		ID:         uuid.New(),
		EntityType: "task",
		EntityID:   uuid.New(),
		Action:     "create",
		UserID:     "test-user",
		Changes: map[string]interface{}{
			"title":  "Test Task",
			"status": "TODO",
		},
		CreatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(auditLog)
	}

	return auditLog
}

// CreateTaskAuditLog creates an audit log for task operations
func (f *AuditLogFactory) CreateTaskAuditLog(taskID uuid.UUID, action string) *entity.AuditLog {
	return f.CreateAuditLog(func(log *entity.AuditLog) {
		log.EntityType = "task"
		log.EntityID = taskID
		log.Action = action
	})
}

// CreateProjectAuditLog creates an audit log for project operations
func (f *AuditLogFactory) CreateProjectAuditLog(projectID uuid.UUID, action string) *entity.AuditLog {
	return f.CreateAuditLog(func(log *entity.AuditLog) {
		log.EntityType = "project"
		log.EntityID = projectID
		log.Action = action
	})
}

// TestDataSeeder provides methods to seed test data
type TestDataSeeder struct {
	ProjectFactory  *ProjectFactory
	TaskFactory     *TaskFactory
	AuditLogFactory *AuditLogFactory
}

// NewTestDataSeeder creates a new TestDataSeeder
func NewTestDataSeeder() *TestDataSeeder {
	return &TestDataSeeder{
		ProjectFactory:  NewProjectFactory(),
		TaskFactory:     NewTaskFactory(),
		AuditLogFactory: NewAuditLogFactory(),
	}
}

// SeedBasicData seeds basic test data including projects and tasks
func (s *TestDataSeeder) SeedBasicData() (*entity.Project, []*entity.Task) {
	project := s.ProjectFactory.CreateProject()
	tasks := []*entity.Task{
		s.TaskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "First Task"
			t.Status = entity.TaskStatusTODO
		}),
		s.TaskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "Second Task"
			t.Status = entity.TaskStatusDONE
		}),
		s.TaskFactory.CreateTask(func(t *entity.Task) {
			t.ProjectID = project.ID
			t.Title = "Third Task"
			t.Status = entity.TaskStatusIMPLEMENTING
		}),
	}

	return project, tasks
}

// SeedComplexData seeds complex test data with multiple projects, tasks, and audit logs
func (s *TestDataSeeder) SeedComplexData() ([]*entity.Project, []*entity.Task, []*entity.AuditLog) {
	// Create projects
	projects := []*entity.Project{
		s.ProjectFactory.CreateProject(func(p *entity.Project) {
			p.Name = "Frontend Project"
			p.Description = "React frontend application"
		}),
		s.ProjectFactory.CreateProject(func(p *entity.Project) {
			p.Name = "Backend Project"
			p.Description = "Go backend API"
		}),
	}

	// Create tasks
	tasks := []*entity.Task{}
	for _, project := range projects {
		projectTasks := s.TaskFactory.CreateTasksWithDifferentStatuses(project.ID)
		tasks = append(tasks, projectTasks...)
	}

	// Create audit logs
	auditLogs := []*entity.AuditLog{}
	for _, task := range tasks {
		auditLogs = append(auditLogs, s.AuditLogFactory.CreateTaskAuditLog(task.ID, "create"))
	}

	return projects, tasks, auditLogs
}