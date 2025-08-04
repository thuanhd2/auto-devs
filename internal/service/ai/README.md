# AI Service - CLI Manager

## Tổng quan

CLI Manager là một service quản lý việc tương tác với Claude Code CLI để thực hiện các task AI. Service này sử dụng cơ chế command line với stdin để truyền prompt, cho phép tích hợp mượt mà với hệ thống quản lý task.

## Cài đặt và Cấu hình

### 1. Yêu cầu hệ thống

- Go 1.21+
- Node.js 18+ (để chạy Claude Code CLI)
- Anthropic API Key

### 2. Cài đặt Claude Code CLI

```bash
# Cài đặt Claude Code CLI globally
npm install -g @anthropic-ai/claude-code

# Hoặc sử dụng npx (khuyến nghị)
npx -y @anthropic-ai/claude-code@latest --help
```

### 3. Thiết lập API Key

```bash
# Thiết lập environment variable
export ANTHROPIC_API_KEY="your-api-key-here"

# Hoặc trong file .env
echo "ANTHROPIC_API_KEY=your-api-key-here" >> .env
```

## Cấu trúc và Sử dụng

### CLIConfig

Cấu hình chính cho CLI Manager:

```go
type CLIConfig struct {
    CLICommand   string        // Command line hoàn chỉnh
    APIKey       string        // Anthropic API Key
    Timeout      time.Duration // Timeout cho execution
    WorkingDir   string        // Working directory
    RetryCount   int           // Số lần retry
    RetryDelay   time.Duration // Delay giữa các lần retry
}
```

### Khởi tạo CLIManager

```go
package main

import (
    "context"
    "log"
    "time"

    "your-project/internal/service/ai"
    "your-project/internal/entity"
)

func main() {
    // Cấu hình mặc định
    config := ai.DefaultCLIConfig()

    // Tùy chỉnh cấu hình
    config.APIKey = "your-api-key"
    config.CLICommand = "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose --output-format=stream-json"
    config.Timeout = 30 * time.Minute
    config.WorkingDir = "/path/to/your/project"
    config.RetryCount = 3
    config.RetryDelay = 5 * time.Second

    // Khởi tạo manager
    manager, err := ai.NewCLIManager(config)
    if err != nil {
        log.Fatal("Failed to create CLI manager:", err)
    }

    defer manager.Close()
}
```

## Hướng dẫn sử dụng

### 1. Thực thi Task cơ bản

```go
func executeBasicTask(manager *ai.CLIManager) {
    ctx := context.Background()

    // Tạo task
    task := entity.Task{
        ID:          uuid.New(),
        Title:       "Implement user authentication",
        Description: "Add JWT-based authentication to the API with proper error handling",
        Status:      entity.TaskStatusPLANNING,
        Priority:    entity.TaskPriorityHigh,
    }

    // Tạo plan
    plan := &ai.Plan{
        ID:     "plan-auth-001",
        TaskID: task.ID.String(),
        Steps: []ai.PlanStep{
            {
                ID:          "step-1",
                Description: "Create user model with validation",
                Action:      "create",
                Order:       1,
            },
            {
                ID:          "step-2",
                Description: "Implement JWT token generation",
                Action:      "implement",
                Order:       2,
            },
            {
                ID:          "step-3",
                Description: "Add authentication middleware",
                Action:      "implement",
                Order:       3,
            },
        },
    }

    // Thực thi task
    result, err := manager.ExecuteTask(ctx, task, plan)
    if err != nil {
        log.Printf("Task execution failed: %v", err)
        return
    }

    log.Printf("Task completed successfully!")
    log.Printf("Output: %s", result.Output)
    log.Printf("Execution time: %v", result.ExecutionTime)
}
```

### 2. Thực thi với Prompt tùy chỉnh

```go
func executeCustomPrompt(manager *ai.CLIManager) {
    ctx := context.Background()

    task := entity.Task{
        ID:          uuid.New(),
        Title:       "Refactor database queries",
        Description: "Optimize database queries for better performance",
        Status:      entity.TaskStatusIMPLEMENTING,
        Priority:    entity.TaskPriorityMedium,
    }

    // Tạo prompt tùy chỉnh
    customPrompt := `
    Task: Refactor database queries for better performance

    Current codebase uses GORM with PostgreSQL. Please:
    1. Identify slow queries
    2. Add proper indexes
    3. Optimize N+1 queries
    4. Add query logging for monitoring

    Focus on the user and project repositories.
    `

    // Thực thi với prompt tùy chỉnh
    result, err := manager.ExecuteCommand(ctx, customPrompt)
    if err != nil {
        log.Printf("Custom execution failed: %v", err)
        return
    }

    log.Printf("Custom execution completed: %s", result.Output)
}
```

### 3. Xử lý lỗi và Retry

```go
func executeWithErrorHandling(manager *ai.CLIManager) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
    defer cancel()

    task := entity.Task{
        ID:          uuid.New(),
        Title:       "Complex refactoring task",
        Description: "Large scale refactoring with multiple dependencies",
        Status:      entity.TaskStatusPLANNING,
        Priority:    entity.TaskPriorityHigh,
    }

    plan := &ai.Plan{
        ID:     "plan-refactor-001",
        TaskID: task.ID.String(),
        Steps: []ai.PlanStep{
            {
                ID:          "step-1",
                Description: "Analyze current architecture",
                Action:      "analyze",
                Order:       1,
            },
        },
    }

    // Thực thi với error handling
    result, err := manager.ExecuteTask(ctx, task, plan)
    if err != nil {
        // Kiểm tra loại lỗi
        switch {
        case errors.Is(err, context.DeadlineExceeded):
            log.Printf("Task timed out after %v", ctx.Err())
        case errors.Is(err, ai.ErrInvalidConfig):
            log.Printf("Invalid configuration: %v", err)
        case errors.Is(err, ai.ErrExecutionFailed):
            log.Printf("Execution failed after retries: %v", err)
        default:
            log.Printf("Unexpected error: %v", err)
        }
        return
    }

    log.Printf("Task completed successfully in %v", result.ExecutionTime)
}
```

## Cấu hình nâng cao

### 1. Cấu hình cho môi trường khác nhau

```go
// Development environment
func getDevConfig() *ai.CLIConfig {
    return &ai.CLIConfig{
        CLICommand:   "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose",
        APIKey:       os.Getenv("ANTHROPIC_API_KEY"),
        Timeout:      10 * time.Minute,
        WorkingDir:   "./",
        RetryCount:   2,
        RetryDelay:   3 * time.Second,
    }
}

// Production environment
func getProdConfig() *ai.CLIConfig {
    return &ai.CLIConfig{
        CLICommand:   "/usr/local/bin/claude-code -p --model claude-3.5-sonnet --max-tokens 4000",
        APIKey:       os.Getenv("ANTHROPIC_API_KEY"),
        Timeout:      30 * time.Minute,
        WorkingDir:   "/app",
        RetryCount:   5,
        RetryDelay:   10 * time.Second,
    }
}
```

### 2. Cấu hình với custom CLI

```go
// Sử dụng custom CLI binary
config := &ai.CLIConfig{
    CLICommand:   "/path/to/custom/claude-cli --model claude-3.5-sonnet --max-tokens 4000 --temperature 0.1",
    APIKey:       "your-api-key",
    Timeout:      20 * time.Minute,
    WorkingDir:   "/path/to/project",
    RetryCount:   3,
    RetryDelay:   5 * time.Second,
}
```

## Environment Variables

CLI Manager tự động thiết lập các environment variables:

```bash
# Bắt buộc
ANTHROPIC_API_KEY=your-api-key

# Tùy chọn
CLAUDE_WORKING_DIR=/path/to/working/directory
CLAUDE_LOG_LEVEL=info  # info, error, debug
CLAUDE_MODEL=claude-3.5-sonnet
CLAUDE_MAX_TOKENS=4000
```

## Xử lý lỗi

### Các loại lỗi thường gặp

```go
// 1. Lỗi cấu hình
if errors.Is(err, ai.ErrInvalidConfig) {
    log.Printf("Kiểm tra lại cấu hình CLI Manager")
}

// 2. Lỗi timeout
if errors.Is(err, context.DeadlineExceeded) {
    log.Printf("Task mất quá nhiều thời gian, tăng timeout hoặc chia nhỏ task")
}

// 3. Lỗi API
if errors.Is(err, ai.ErrAPIError) {
    log.Printf("Lỗi API, kiểm tra API key và quota")
}

// 4. Lỗi execution
if errors.Is(err, ai.ErrExecutionFailed) {
    log.Printf("Lỗi thực thi, kiểm tra CLI command và working directory")
}
```

### Retry Strategy

```go
// Cấu hình retry tùy chỉnh
config := ai.DefaultCLIConfig()
config.RetryCount = 5
config.RetryDelay = 10 * time.Second

// Exponential backoff
config.RetryDelay = time.Duration(1<<config.RetryCount) * time.Second
```

## Testing

### Unit Tests

```bash
# Chạy tất cả tests
go test ./internal/service/ai/ -v

# Chạy test cụ thể
go test ./internal/service/ai/ -run TestCLIManager_ExecuteTask

# Chạy test với coverage
go test ./internal/service/ai/ -cover -coverprofile=coverage.out
```

### Integration Tests

```go
func TestCLIManagerIntegration(t *testing.T) {
    // Skip nếu không có API key
    if os.Getenv("ANTHROPIC_API_KEY") == "" {
        t.Skip("Skipping integration test: no API key")
    }

    config := ai.DefaultCLIConfig()
    config.APIKey = os.Getenv("ANTHROPIC_API_KEY")

    manager, err := ai.NewCLIManager(config)
    require.NoError(t, err)
    defer manager.Close()

    // Test với task thực tế
    task := entity.Task{
        ID:          uuid.New(),
        Title:       "Test task",
        Description: "This is a test task for integration testing",
        Status:      entity.TaskStatusPLANNING,
        Priority:    entity.TaskPriorityLow,
    }

    result, err := manager.ExecuteTask(context.Background(), task, nil)
    require.NoError(t, err)
    assert.NotEmpty(t, result.Output)
}
```

## Best Practices

### 1. Resource Management

```go
// Luôn đóng manager sau khi sử dụng
manager, err := ai.NewCLIManager(config)
if err != nil {
    return err
}
defer manager.Close()

// Sử dụng context với timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()
```

### 2. Error Handling

```go
// Luôn kiểm tra lỗi
result, err := manager.ExecuteTask(ctx, task, plan)
if err != nil {
    // Log lỗi chi tiết
    log.Printf("Task execution failed: %+v", err)

    // Xử lý theo loại lỗi
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        // Retry với timeout lớn hơn
    case errors.Is(err, ai.ErrAPIError):
        // Kiểm tra API key và quota
    default:
        // Lỗi không xác định
    }
    return err
}
```

### 3. Performance Optimization

```go
// Sử dụng connection pooling cho manager
var manager *ai.CLIManager

func init() {
    config := ai.DefaultCLIConfig()
    config.APIKey = os.Getenv("ANTHROPIC_API_KEY")

    var err error
    manager, err = ai.NewCLIManager(config)
    if err != nil {
        log.Fatal(err)
    }
}

func cleanup() {
    if manager != nil {
        manager.Close()
    }
}
```

## Troubleshooting

### Vấn đề thường gặp

1. **"Command not found"**

   - Kiểm tra CLI command có đúng không
   - Đảm bảo Node.js và npm đã được cài đặt

2. **"API key invalid"**

   - Kiểm tra ANTHROPIC_API_KEY environment variable
   - Đảm bảo API key có quyền truy cập Claude

3. **"Timeout exceeded"**

   - Tăng timeout trong config
   - Chia nhỏ task thành các bước nhỏ hơn

4. **"Working directory not found"**
   - Kiểm tra đường dẫn working directory
   - Đảm bảo có quyền truy cập thư mục

### Debug Mode

```go
// Bật debug mode
config := ai.DefaultCLIConfig()
config.Debug = true

// Hoặc set environment variable
os.Setenv("CLAUDE_LOG_LEVEL", "debug")
```

## Migration từ phiên bản cũ

### Từ cấu trúc cũ sang mới

```go
// Cấu trúc cũ
oldConfig := &ai.CLIConfig{
    CLIPath:   "claude",
    Model:     "claude-3.5-sonnet",
    MaxTokens: 4000,
    APIKey:    "your-api-key",
}

// Cấu trúc mới
newConfig := &ai.CLIConfig{
    CLICommand: "npx -y @anthropic-ai/claude-code@latest -p --dangerously-skip-permissions --verbose --output-format=stream-json",
    APIKey:     "your-api-key",
    Timeout:    30 * time.Minute,
    WorkingDir: "/path/to/project",
}
```

## Lưu ý quan trọng

1. **API Key Security**: Không bao giờ commit API key vào source code
2. **Rate Limiting**: Claude API có giới hạn rate, cần xử lý gracefully
3. **Cost Management**: Monitor usage để tránh chi phí cao
4. **Error Recovery**: Implement proper retry logic cho production
5. **Logging**: Log đầy đủ để debug và monitor

## Liên hệ và Hỗ trợ

- **Documentation**: Xem thêm docs trong thư mục `docs/`
- **Issues**: Tạo issue trên GitHub repository
- **Examples**: Xem thêm examples trong thư mục `examples/`
