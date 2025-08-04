# Execution Service

## Tổng quan

`ExecutionService` là service chính để orchestrate toàn bộ AI execution workflow. Service này quản lý việc thực thi các AI tasks từ planning đến completion, cung cấp real-time updates và execution control.

## Các tính năng chính

### 1. Execution Workflow Management

- **StartExecution**: Bắt đầu một AI execution mới
- **GetExecution**: Lấy thông tin execution theo ID
- **ListExecutions**: Liệt kê tất cả active executions
- **CleanupCompletedExecutions**: Dọn dẹp completed executions

### 2. Execution Control

- **CancelExecution**: Hủy execution đang chạy
- **PauseExecution**: Tạm dừng execution (placeholder - chưa implement đầy đủ)
- **ResumeExecution**: Tiếp tục execution (placeholder - chưa implement đầy đủ)

### 3. Real-time Updates

- **SetUpdateCallback**: Thiết lập callback cho real-time updates
- **Progress tracking**: Theo dõi tiến độ execution (0.0 - 1.0)
- **Live logging**: Log streaming real-time
- **Status updates**: Cập nhật trạng thái execution

## Cấu trúc dữ liệu

### Execution Status

```go
type ExecutionStatus string

const (
    ExecutionStatusPending   ExecutionStatus = "pending"
    ExecutionStatusRunning   ExecutionStatus = "running"
    ExecutionStatusPaused    ExecutionStatus = "paused"
    ExecutionStatusCompleted ExecutionStatus = "completed"
    ExecutionStatusFailed    ExecutionStatus = "failed"
    ExecutionStatusCancelled ExecutionStatus = "cancelled"
)
```

### Execution

```go
type Execution struct {
    ID          string          `json:"id"`
    TaskID      string          `json:"task_id"`
    Plan        Plan            `json:"plan"`
    Status      ExecutionStatus `json:"status"`
    StartedAt   time.Time       `json:"started_at"`
    CompletedAt *time.Time      `json:"completed_at,omitempty"`
    Error       string          `json:"error,omitempty"`
    Progress    float64         `json:"progress"` // 0.0 to 1.0
    Logs        []string        `json:"logs"`
    Result      *ExecutionResult `json:"result,omitempty"`
}
```

### ExecutionUpdate

```go
type ExecutionUpdate struct {
    ExecutionID string          `json:"execution_id"`
    Status      ExecutionStatus `json:"status"`
    Progress    float64         `json:"progress"`
    Log         string          `json:"log,omitempty"`
    Error       string          `json:"error,omitempty"`
    Timestamp   time.Time       `json:"timestamp"`
}
```

## Cách sử dụng

### Khởi tạo

```go
cliManager, err := NewCLIManager(DefaultCLIConfig())
if err != nil {
    log.Fatal(err)
}

processManager := NewProcessManager()
executionService := NewExecutionService(cliManager, processManager)
```

### Thiết lập real-time updates

```go
executionService.SetUpdateCallback(func(update ExecutionUpdate) {
    // Gửi update qua WebSocket hoặc xử lý theo nhu cầu
    fmt.Printf("Execution %s: %s (%.1f%%)\n",
        update.ExecutionID,
        update.Status,
        update.Progress*100)
})
```

### Bắt đầu execution

```go
plan := Plan{
    ID:          "plan-1",
    TaskID:      "task-1",
    Description: "Sample AI execution plan",
    Steps:       []PlanStep{},
    Context:     map[string]string{},
    CreatedAt:   time.Now(),
}

execution, err := executionService.StartExecution("task-1", plan)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Execution started: %s\n", execution.ID)
```

### Kiểm soát execution

```go
// Hủy execution
err = executionService.CancelExecution(execution.ID)

// Tạm dừng execution
err = executionService.PauseExecution(execution.ID)

// Tiếp tục execution
err = executionService.ResumeExecution(execution.ID)
```

### Theo dõi execution

```go
// Lấy thông tin execution
execution, err := executionService.GetExecution(executionID)
if err != nil {
    log.Printf("Execution not found: %v", err)
    return
}

fmt.Printf("Status: %s, Progress: %.1f%%\n",
    execution.Status,
    execution.Progress*100)

// Liệt kê tất cả active executions
executions := executionService.ListExecutions()
for _, exec := range executions {
    fmt.Printf("Execution %s: %s\n", exec.ID, exec.Status)
}
```

## Workflow Execution

### 1. Execution Lifecycle

```
Pending → Running → (Paused) → Completed/Failed/Cancelled
```

### 2. Process Management

- **Step 1**: Build CLI command từ Plan
- **Step 2**: Spawn process với ProcessManager
- **Step 3**: Monitor process output và progress
- **Step 4**: Handle completion/failure
- **Step 5**: Cleanup resources

### 3. Progress Estimation

Service tự động estimate progress dựa trên output patterns:

- "completed", "done" → 100%
- "processing", "running" → 50%
- "starting", "initializing" → 20%
- Default → 0%

## Integration với WebSocket

ExecutionService có thể tích hợp với WebSocket để gửi real-time updates đến frontend:

```go
// Thiết lập WebSocket integration
executionService.SetUpdateCallback(func(update ExecutionUpdate) {
    message := map[string]interface{}{
        "type": "execution_update",
        "data": update,
    }

    // Gửi đến tất cả connected clients
    websocketHub.Broadcast(message)
})
```

## Error Handling

### Common Errors

- **Execution not found**: Khi execution ID không tồn tại
- **Invalid status transition**: Khi thực hiện action không hợp lệ với status hiện tại
- **Process start failure**: Khi không thể start AI process
- **Command build failure**: Khi không thể build CLI command

### Recovery Mechanisms

- **Automatic cleanup**: Completed executions được tự động cleanup
- **Context cancellation**: Sử dụng context để cancel execution gracefully
- **Process termination**: Force kill process nếu cần thiết

## Testing

Service có comprehensive test suite trong `execution_service_test.go`:

```bash
go test ./internal/service/ai -v -run TestExecutionService
```

## Future Enhancements

### Planned Features

1. **Database persistence**: Lưu execution history vào database
2. **Retry mechanism**: Tự động retry failed executions
3. **Resource monitoring**: Monitor CPU/memory usage
4. **Execution queuing**: Queue management cho multiple executions
5. **Advanced progress tracking**: ML-based progress estimation
6. **Execution templates**: Reusable execution templates

### Pause/Resume Implementation

Hiện tại pause/resume chỉ là placeholder. Cần implement:

- Process suspension/resumption
- State persistence
- Resource management

## Dependencies

- **CLIManager**: Quản lý CLI commands
- **ProcessManager**: Quản lý AI processes
- **uuid**: Generate unique execution IDs
- **context**: Cancellation support
- **sync**: Thread-safe operations
