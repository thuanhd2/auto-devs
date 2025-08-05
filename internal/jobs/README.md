# Job Client Integration

Package này cung cấp integration với asynq để xử lý background jobs trong hệ thống.

## Cấu trúc

### Client

- `Client`: Wrapper cho asynq.Client để enqueue jobs
- `ClientInterface`: Interface cho job client operations
- `JobClientAdapter`: Adapter để kết nối job client với usecase layer

### Job Types

- `TaskPlanningJob`: Job để xử lý planning cho tasks

## Cách sử dụng

### 1. Cấu hình Redis

Thêm cấu hình Redis vào file config:

```go
type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}
```

### 2. Dependency Injection

Job client được inject thông qua Wire:

```go
// Trong wire.go
func ProvideJobClient(cfg *config.Config) *jobs.Client {
    redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
    return jobs.NewClient(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
}

func ProvideJobClientAdapter(client *jobs.Client) usecase.JobClientInterface {
    return jobs.NewJobClientAdapter(client)
}
```

### 3. Sử dụng trong TaskUsecase

```go
func (u *taskUsecase) StartPlanning(ctx context.Context, taskID uuid.UUID, branchName string) (string, error) {
    // Validate task
    task, err := u.taskRepo.GetByID(ctx, taskID)
    if err != nil {
        return "", fmt.Errorf("failed to get task: %w", err)
    }

    if task.Status != entity.TaskStatusTODO {
        return "", fmt.Errorf("task must be in TODO status to start planning")
    }

    // Enqueue planning job
    payload := &TaskPlanningPayload{
        TaskID:     taskID,
        BranchName: branchName,
        ProjectID:  task.ProjectID,
    }

    jobID, err := u.jobClient.EnqueueTaskPlanning(payload, 0)
    if err != nil {
        return "", fmt.Errorf("failed to enqueue planning job: %w", err)
    }

    return jobID, nil
}
```

### 4. Job Processing

Jobs được xử lý bởi `Processor` trong `internal/jobs/processor.go`:

```go
func (p *Processor) ProcessTaskPlanning(ctx context.Context, task *asynq.Task) error {
    payload, err := ParseTaskPlanningPayload(task)
    if err != nil {
        return fmt.Errorf("failed to parse task planning payload: %w", err)
    }

    // Process the planning job
    // 1. Update task status to PLANNING
    // 2. Create git worktree
    // 3. Run AI executor
    // 4. Update task with results

    return nil
}
```

## Testing

### Unit Tests

```go
func TestStartPlanning_WithJobClient(t *testing.T) {
    // Setup mocks
    mockJobClient := &MockJobClient{}
    mockJobClient.On("EnqueueTaskPlanning", mock.Anything, time.Duration(0)).Return("job-123", nil)

    // Test usecase
    usecase := &taskUsecase{jobClient: mockJobClient}
    jobID, err := usecase.StartPlanning(ctx, taskID, branchName)

    assert.NoError(t, err)
    assert.Equal(t, "job-123", jobID)
}
```

### Integration Tests

Để test với Redis thực tế:

```go
func TestJobClientIntegration(t *testing.T) {
    // Setup Redis connection
    client := jobs.NewClient("localhost:6379", "", 0)
    defer client.Close()

    // Test enqueue
    payload := &jobs.TaskPlanningPayload{
        TaskID:     uuid.New(),
        BranchName: "test-branch",
        ProjectID:  uuid.New(),
    }

    jobID, err := client.EnqueueTaskPlanningString(payload, 0)
    assert.NoError(t, err)
    assert.NotEmpty(t, jobID)
}
```

## Environment Variables

Cấu hình Redis thông qua environment variables:

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Job Queue Configuration

Jobs được xử lý với các cấu hình sau:

- **Queue**: `planning` (dedicated queue cho planning jobs)
- **Max Retry**: 3 lần
- **Timeout**: 30 phút
- **Concurrency**: 4 workers

## Error Handling

- Jobs sẽ được retry tối đa 3 lần nếu fail
- Timeout được set 30 phút cho planning jobs
- Errors được log thông qua asynq error handler

## Monitoring

Job status có thể được monitor thông qua:

- asynq.Inspector để xem job queue status
- Redis commands để xem queue metrics
- Application logs cho job processing events
