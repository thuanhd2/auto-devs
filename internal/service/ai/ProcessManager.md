## ProcessManager

`ProcessManager` là service chính để spawn, monitor và control các AI execution processes.

### Tính năng chính

- **Process Spawning**: Tạo và khởi chạy các AI processes
- **Process Monitoring**: Theo dõi trạng thái và resource usage của processes
- **Process Control**: Terminate hoặc kill processes
- **Output Collection**: Thu thập stdout/stderr streams
- **Resource Tracking**: Theo dõi CPU và memory usage

### Cách sử dụng

```go
package main

import (
    "fmt"
    "time"
    "your-project/internal/service/ai"
)

func main() {
    // Tạo ProcessManager instance
    pm := ai.NewProcessManager()

    // Spawn một process
    process, err := pm.SpawnProcess("python script.py", "/path/to/workdir")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Process started with ID: %s, PID: %d\n", process.ID, process.PID)

    // Monitor process status
    go func() {
        for process.IsRunning() {
            status := process.GetStatus()
            duration := process.GetDuration()
            fmt.Printf("Status: %s, Duration: %v\n", status, duration)
            time.Sleep(1 * time.Second)
        }
    }()

    // Wait for process to complete
    time.Sleep(5 * time.Second)

    // Get output
    stdout, stderr := process.GetOutput()
    fmt.Printf("Stdout: %s\n", string(stdout))
    fmt.Printf("Stderr: %s\n", string(stderr))

    // Check final status
    if process.ExitCode != nil {
        fmt.Printf("Exit code: %d\n", *process.ExitCode)
    }
}
```

### Process Control

```go
// Graceful termination với SIGTERM
err := pm.TerminateProcess(process)
if err != nil {
    fmt.Printf("Failed to terminate process: %v\n", err)
}

// Force kill với SIGKILL
err = pm.KillProcess(process)
if err != nil {
    fmt.Printf("Failed to kill process: %v\n", err)
}
```

### Process Monitoring

```go
// Lấy tất cả active processes
processes := pm.ListProcesses()
for _, p := range processes {
    fmt.Printf("Process %s: %s (PID: %d)\n", p.ID, p.GetStatus(), p.PID)
}

// Lấy process theo ID
if process, exists := pm.GetProcess("process_id"); exists {
    fmt.Printf("Found process: %s\n", process.ID)
}
```

### Environment Variables

ProcessManager tự động set các environment variables sau cho mỗi process:

- `AI_PROCESS_ID`: Unique ID của process
- `AI_WORK_DIR`: Working directory của process

### Process States

Process có thể ở các trạng thái sau:

- `starting`: Process đang được khởi tạo
- `running`: Process đang chạy
- `stopped`: Process đã kết thúc bình thường
- `killed`: Process bị force kill
- `error`: Process gặp lỗi

### Thread Safety

ProcessManager được thiết kế để thread-safe và có thể được sử dụng từ multiple goroutines.

### Error Handling

ProcessManager cung cấp comprehensive error handling:

- Process creation errors
- Process termination errors
- Resource cleanup errors
- Monitoring errors

### Resource Management

ProcessManager tự động cleanup resources khi process kết thúc:

- Remove process từ internal map
- Close stdout/stderr pipes
- Cancel context
- Cleanup temporary files (nếu có)

### Testing

Package có comprehensive test suite:

```bash
go test ./internal/service/ai -v
```

Tests bao gồm:

- Process spawning
- Process monitoring
- Process control (terminate/kill)
- Output collection
- Environment variables
- Working directory setup
- Error handling
