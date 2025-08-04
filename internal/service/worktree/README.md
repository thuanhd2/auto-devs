# Worktree Management Service

Service này cung cấp chức năng quản lý thư mục worktree cho các task trong hệ thống. Worktree là thư mục riêng biệt cho mỗi task, cho phép phát triển song song và độc lập.

## Cấu trúc thư mục

```
/worktrees/
├── project-{id}/
│   └── task-{id}/     # Task-specific worktrees
```

## Tính năng chính

### 1. Quản lý cấu hình worktree

- Cấu hình thư mục cơ sở
- Kiểm tra quyền truy cập và validation
- Giám sát dung lượng ổ đĩa cơ bản

### 2. Quản lý đường dẫn worktree

- Tạo đường dẫn duy nhất cho mỗi task
- Xử lý xung đột đường dẫn cơ bản
- Validation độ dài và ký tự đường dẫn

### 3. Validation thư mục worktree

- Kiểm tra sự tồn tại và quyền của thư mục
- Validation dung lượng ổ đĩa khả dụng
- Đảm bảo thư mục không đang được sử dụng

### 4. Cleanup worktree

- Xóa worktree khi task hoàn thành/hủy bỏ
- Cleanup branch từ local repository
- Xóa thư mục worktree và files

## Sử dụng

### Khởi tạo WorktreeManager

```go
import "github.com/auto-devs/auto-devs/internal/service/worktree"

// Sử dụng cấu hình mặc định
manager, err := worktree.NewWorktreeManager(nil)
if err != nil {
    log.Fatal(err)
}

// Hoặc sử dụng cấu hình tùy chỉnh
config := &worktree.WorktreeConfig{
    BaseDirectory:   "/custom/worktrees",
    MaxPathLength:   2048,
    MinDiskSpace:    200 * 1024 * 1024, // 200MB
    CleanupInterval: 12 * time.Hour,
    EnableLogging:   true,
}

manager, err := worktree.NewWorktreeManager(config)
```

### Tạo worktree cho task

```go
ctx := context.Background()
worktreePath, err := manager.CreateWorktree(ctx, "project-123", "task-456")
if err != nil {
    log.Printf("Failed to create worktree: %v", err)
    return
}

fmt.Printf("Created worktree at: %s\n", worktreePath)
```

### Kiểm tra worktree tồn tại

```go
if manager.WorktreeExists(worktreePath) {
    fmt.Println("Worktree exists")
} else {
    fmt.Println("Worktree does not exist")
}
```

### Lấy thông tin worktree

```go
info, err := manager.GetWorktreeInfo(worktreePath)
if err != nil {
    log.Printf("Failed to get worktree info: %v", err)
    return
}

fmt.Printf("Worktree path: %s\n", info.Path)
fmt.Printf("Created at: %s\n", info.CreatedAt)
fmt.Printf("File count: %d\n", info.FileCount)
fmt.Printf("Size: %d bytes\n", info.Size)
```

### Liệt kê worktrees của project

```go
worktrees, err := manager.ListWorktrees("project-123")
if err != nil {
    log.Printf("Failed to list worktrees: %v", err)
    return
}

for _, wt := range worktrees {
    fmt.Printf("Worktree: %s\n", wt)
}
```

### Cleanup worktree

```go
err = manager.CleanupWorktree(ctx, worktreePath)
if err != nil {
    log.Printf("Failed to cleanup worktree: %v", err)
    return
}

fmt.Println("Worktree cleaned up successfully")
```

## Cấu hình

### WorktreeConfig

```go
type WorktreeConfig struct {
    BaseDirectory    string        // Thư mục cơ sở cho tất cả worktrees
    MaxPathLength    int           // Độ dài tối đa đường dẫn cho phép
    MinDiskSpace     int64         // Dung lượng ổ đĩa tối thiểu cần thiết (bytes)
    CleanupInterval  time.Duration // Khoảng thời gian cleanup operations
    EnableLogging    bool          // Bật logging chi tiết
    LogLevel         slog.Level    // Log level cho worktree operations
}
```

### Cấu hình mặc định

- **BaseDirectory**: `/worktrees`
- **MaxPathLength**: 4096
- **MinDiskSpace**: 100MB
- **CleanupInterval**: 24 giờ
- **EnableLogging**: true
- **LogLevel**: Info

## Validation

### Validation đường dẫn

- Loại bỏ ký tự không hợp lệ (`/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`)
- Thay thế bằng dấu gạch dưới (`_`)
- Loại bỏ khoảng trắng đầu/cuối
- Giới hạn độ dài tối đa

### Validation quyền truy cập

- Kiểm tra thư mục tồn tại
- Kiểm tra quyền đọc
- Kiểm tra quyền ghi

### Validation dung lượng ổ đĩa

- Kiểm tra dung lượng khả dụng
- Đảm bảo đủ không gian cho worktree mới

## Error Handling

Service này trả về các lỗi có ý nghĩa cho các trường hợp:

- Thư mục không tồn tại
- Không đủ quyền truy cập
- Không đủ dung lượng ổ đĩa
- Worktree đã tồn tại
- Đường dẫn không hợp lệ
- Lỗi hệ thống file

## Testing

Chạy tests:

```bash
go test ./internal/service/worktree -v
```

Tests bao gồm:

- Tạo worktree manager với cấu hình khác nhau
- Tạo và xóa worktree
- Validation đường dẫn
- Kiểm tra tồn tại worktree
- Liệt kê worktrees
- Lấy thông tin worktree
- Cleanup operations

## Tích hợp với Git

Worktree service được thiết kế để tích hợp với Git service hiện có:

```go
// Tạo worktree cho task
worktreePath, err := worktreeManager.CreateWorktree(ctx, projectID, taskID)
if err != nil {
    return err
}

// Tạo branch cho task
branchName, err := gitManager.GenerateBranchName(taskID, taskTitle)
if err != nil {
    return err
}

// Tạo worktree Git
err = gitManager.CreateBranchFromMain(ctx, worktreePath, branchName)
if err != nil {
    return err
}
```

## Bảo mật

- Tất cả đường dẫn được validate để tránh path traversal attacks
- Kiểm tra quyền truy cập trước khi thực hiện operations
- Cleanup tự động để tránh rò rỉ tài nguyên
- Logging chi tiết cho audit trail
