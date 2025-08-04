# Hướng dẫn viết Unit Test cho Usecase

## Tổng quan

Khi viết unit test cho usecase, chúng ta cần mock tất cả các dependency (repository, service, usecase khác) để đảm bảo test chỉ tập trung vào logic của usecase đó.

## Cấu trúc test

### 1. Setup Mock Objects

```go
// Mock repository
type MockTaskRepository struct {
    mock.Mock
}

// Implement tất cả methods của interface
func (m *MockTaskRepository) Create(ctx context.Context, task *entity.Task) error {
    args := m.Called(ctx, task)
    return args.Error(0)
}

// Mock service/usecase khác
type MockNotificationUsecase struct {
    mock.Mock
}

func (m *MockNotificationUsecase) CreateNotification(ctx context.Context, req CreateNotificationRequest) (*entity.Notification, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*entity.Notification), args.Error(1)
}
```

### 2. Setup Test Function

```go
func setupTaskUsecaseTest() (*taskUsecase, *MockTaskRepository, *MockProjectRepository, *MockNotificationUsecase) {
    mockTaskRepo := &MockTaskRepository{}
    mockProjectRepo := &MockProjectRepository{}
    mockNotificationUsecase := &MockNotificationUsecase{}

    usecase := &taskUsecase{
        taskRepo:            mockTaskRepo,
        projectRepo:         mockProjectRepo,
        notificationUsecase: mockNotificationUsecase,
    }

    return usecase, mockTaskRepo, mockProjectRepo, mockNotificationUsecase
}
```

### 3. Viết Test Cases

```go
func TestTaskUsecase_Create(t *testing.T) {
    usecase, mockTaskRepo, mockProjectRepo, _ := setupTaskUsecaseTest()
    ctx := context.Background()

    t.Run("successful task creation", func(t *testing.T) {
        // Arrange
        req := CreateTaskRequest{
            ProjectID:   uuid.New(),
            Title:       "Test Task",
            Description: "Test Description",
            Priority:    entity.TaskPriorityMedium,
        }

        // Mock expectations
        mockProjectRepo.On("ValidateProjectExists", ctx, req.ProjectID).Return(true, nil)
        mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(nil)

        // Act
        result, err := usecase.Create(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, req.Title, result.Title)
        assert.Equal(t, req.Description, result.Description)
        assert.Equal(t, req.Priority, result.Priority)
        assert.Equal(t, entity.TaskStatusTodo, result.Status)

        // Verify mocks
        mockProjectRepo.AssertExpectations(t)
        mockTaskRepo.AssertExpectations(t)
    })

    t.Run("project not found", func(t *testing.T) {
        // Arrange
        req := CreateTaskRequest{
            ProjectID: uuid.New(),
            Title:     "Test Task",
        }

        // Mock expectations
        mockProjectRepo.On("ValidateProjectExists", ctx, req.ProjectID).Return(false, nil)

        // Act
        result, err := usecase.Create(ctx, req)

        // Assert
        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "project not found")

        // Verify mocks
        mockProjectRepo.AssertExpectations(t)
    })

    t.Run("repository error", func(t *testing.T) {
        // Arrange
        req := CreateTaskRequest{
            ProjectID: uuid.New(),
            Title:     "Test Task",
        }

        // Mock expectations
        mockProjectRepo.On("ValidateProjectExists", ctx, req.ProjectID).Return(true, nil)
        mockTaskRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(errors.New("database error"))

        // Act
        result, err := usecase.Create(ctx, req)

        // Assert
        assert.Error(t, err)
        assert.Nil(t, result)
        assert.Contains(t, err.Error(), "database error")

        // Verify mocks
        mockProjectRepo.AssertExpectations(t)
        mockTaskRepo.AssertExpectations(t)
    })
}
```

## Best Practices

### 1. Test Structure (AAA Pattern)

- **Arrange**: Chuẩn bị dữ liệu và mock expectations
- **Act**: Thực thi method cần test
- **Assert**: Kiểm tra kết quả và verify mocks

### 2. Test Cases Coverage

- **Happy Path**: Test case thành công
- **Error Cases**: Test các trường hợp lỗi
- **Edge Cases**: Test các trường hợp đặc biệt
- **Validation**: Test validation logic

### 3. Mock Expectations

```go
// Mock với specific arguments
mockRepo.On("GetByID", ctx, expectedID).Return(expectedResult, nil)

// Mock với any arguments
mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Task")).Return(nil)

// Mock với matcher
mockRepo.On("Update", ctx, mock.MatchedBy(func(task *entity.Task) bool {
    return task.Title == "Expected Title"
})).Return(nil)
```

### 4. Verify Mocks

```go
// Verify tất cả expectations được gọi
mockRepo.AssertExpectations(t)

// Verify specific method được gọi đúng số lần
mockRepo.AssertNumberOfCalls(t, "Create", 1)

// Verify method không được gọi
mockRepo.AssertNotCalled(t, "Delete")
```

## Ví dụ Test cho ProjectUsecase

```go
func TestProjectUsecase_Create(t *testing.T) {
    usecase, mockProjectRepo, mockAuditUsecase := setupProjectUsecaseTest()
    ctx := context.Background()

    t.Run("successful project creation", func(t *testing.T) {
        // Arrange
        req := CreateProjectRequest{
            Name:        "Test Project",
            Description: "Test Description",
            RepoURL:     "https://github.com/test/project",
        }

        // Mock expectations
        mockProjectRepo.On("CheckNameExists", ctx, req.Name, (*uuid.UUID)(nil)).Return(false, nil)
        mockProjectRepo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).Return(nil)
        mockAuditUsecase.On("CreateAuditLog", ctx, mock.AnythingOfType("CreateAuditLogRequest")).Return(&entity.AuditLog{}, nil)

        // Act
        result, err := usecase.Create(ctx, req)

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, req.Name, result.Name)
        assert.Equal(t, req.Description, result.Description)
        assert.Equal(t, req.RepoURL, result.RepoURL)

        // Verify mocks
        mockProjectRepo.AssertExpectations(t)
        mockAuditUsecase.AssertExpectations(t)
    })
}
```

## Chạy Tests

```bash
# Chạy tất cả tests
go test ./...

# Chạy tests với verbose output
go test ./... -v

# Chạy tests với coverage
go test ./... -cover

# Chạy tests trong package cụ thể
go test ./internal/usecase -v

# Chạy test function cụ thể
go test ./internal/usecase -v -run TestTaskUsecase_Create
```

## Generate Mocks

Để generate mock files, sử dụng lệnh:

```bash
make mocks
```

Lệnh này sẽ:

1. Download mockery tool nếu chưa có
2. Generate mock files cho tất cả interfaces được định nghĩa trong `.mockery.yaml`
3. Lưu mock files vào thư mục tương ứng

## Lưu ý quan trọng

1. **Import Cycle**: Tránh import cycle bằng cách không import package chính trong test
2. **Mock Interface**: Chỉ mock interface, không mock struct
3. **Test Isolation**: Mỗi test case phải độc lập, không phụ thuộc vào test case khác
4. **Cleanup**: Sử dụng `t.Cleanup()` để cleanup resources nếu cần
5. **Table Driven Tests**: Sử dụng table driven tests cho các test cases tương tự

## Ví dụ Table Driven Test

```go
func TestTaskUsecase_ValidateStatusTransition(t *testing.T) {
    usecase, mockTaskRepo, _, _ := setupTaskUsecaseTest()
    ctx := context.Background()

    tests := []struct {
        name        string
        currentStatus entity.TaskStatus
        newStatus    entity.TaskStatus
        shouldError  bool
        errorMsg     string
    }{
        {
            name:         "valid transition from todo to in progress",
            currentStatus: entity.TaskStatusTodo,
            newStatus:     entity.TaskStatusInProgress,
            shouldError:   false,
        },
        {
            name:         "invalid transition from done to todo",
            currentStatus: entity.TaskStatusDone,
            newStatus:     entity.TaskStatusTodo,
            shouldError:   true,
            errorMsg:      "invalid status transition",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            taskID := uuid.New()
            task := &entity.Task{
                ID:     taskID,
                Status: tt.currentStatus,
            }

            mockTaskRepo.On("GetByID", ctx, taskID).Return(task, nil)

            err := usecase.ValidateStatusTransition(ctx, taskID, tt.newStatus)

            if tt.shouldError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }

            mockTaskRepo.AssertExpectations(t)
        })
    }
}
```
