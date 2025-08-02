# Tổng quan hệ thống

## Mục tiêu và yêu cầu

Hệ thống tự động hóa task cho developer được thiết kế để:

- Tự động hóa quy trình phát triển phần mềm thông qua AI hỗ trợ lập kế hoạch và thực hiện task
- Giảm thiểu công sức thủ công trong việc lập kế hoạch và thực hiện task
- Cải thiện tính nhất quán trong quy trình phát triển
- Duy trì quyền kiểm soát của developer tại các điểm quyết định quan trọng
- Tích hợp mượt mà với các công cụ phát triển hiện có

## Phạm vi

**Trong phạm vi:**

- Quản lý vòng đời task với chuyển đổi trạng thái tự động (TODO → PLANNING → PLAN_REVIEWING → IMPLEMENTING → CODE_REVIEWING → DONE)
- Khả năng lập kế hoạch task bằng AI
- Tích hợp với hệ thống quản lý phiên bản (Git branching và worktree)
- Cấu hình và quản lý dự án
- Quy trình review và phê duyệt task
- Tạo và quản lý Pull Request tự động

**Ngoài phạm vi:**

- Quản lý triển khai code và production
- Tính năng cộng tác nhóm ngoài quản lý task
- Tích hợp với công cụ quản lý dự án bên ngoài (phiên bản đầu)
- Báo cáo và phân tích nâng cao

## Giả định và ràng buộc

**Giả định:**

- Developer có kiến thức cơ bản về Git và quy trình phát triển phần mềm
- Hệ thống được sử dụng trong môi trường phát triển, không phải production
- AI CLI tools (claude-code, etc.) đã được cài đặt và cấu hình sẵn
- Repository có quyền truy cập và authentication phù hợp

**Ràng buộc:**

- Hệ thống chỉ hỗ trợ Git repositories
- MVP chỉ hỗ trợ Claude Code CLI trong giai đoạn đầu
- Tối đa 100 task đồng thời mỗi dự án
- Lập kế hoạch task phải hoàn thành trong vòng 5 phút
- Thời gian phản hồi UI phải dưới 2 giây

# Công nghệ sử dụng

## Backend

- **Language & Framework**: Go với Gin framework
  - Code architecture: Clean Architecture pattern với layers (handler, usecase, repository)
  - Dependency injection với Wire
  - Graceful shutdown và health checks
- **API**: RESTful API với OpenAPI/Swagger documentation
- **Real-time**: WebSocket cho live updates của task status

## Frontend

Clone from https://github.com/satnaing/shadcn-admin

- **Framework**: React.js với TypeScript
- **UI**: ShadcnUI (TailwindCSS + RadixUI)
- **Build Tool**: Vite
- **Routing**: TanStack Router
- **Type Checking**: TypeScript
- **Linting/Formatting**: Eslint & Prettier
- **Icons**: Tabler Icons
- **Build tool**: Vite cho fast development và optimized builds
- **Testing**: Jest + React Testing Library

## Database & Storage

- **Primary Database**: PostgreSQL
  - Database migration: golang-migrate
  - Connection pooling: pgxpool
- **Cache**: Redis cho session management và temporary data
- **File Storage**: Local filesystem cho worktree và temporary files

## DevOps & Infrastructure

- **Logging**: Logrus với structured logging (JSON format)
- **Testing**:
  - Backend: Testify + GoMock
  - Frontend: Jest + React Testing Library
  - Integration tests: Testcontainers
- **CI/CD**: GitHub Actions cho automated testing, building và deployment
- **Monitoring**:
  - Metrics: Prometheus với custom metrics
  - Visualization: Grafana dashboards
  - Health monitoring: Built-in health check endpoints

## External Integrations

- **AI CLI Tools**: Claude Code CLI (MVP), extensible architecture cho future CLIs
- **Git Integration**: Git CLI cho worktree và branch management
- **Process Management**: OS process spawning và monitoring

# Sơ đồ tổng thể

## High-level Architecture Diagram

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Web Browser   │◄───┤  Load Balancer   │────┤   Web Frontend  │
│  (React SPA)    │    │   (nginx/haproxy)│    │   (React + TS)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Backend Services                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   API       │  │  WebSocket  │  │   AI Agent Controller   │  │
│  │  Gateway    │  │   Server    │  │                         │  │
│  │ (Gin REST)  │  │             │  │  - Process Manager      │  │
│  └─────────────┘  └─────────────┘  │  - CLI Orchestrator     │  │
│                                    │  - Status Monitor       │  │
│  ┌─────────────────────────────┐    └─────────────────────────┘  │
│  │   Core Management Service   │                                 │
│  │                             │    ┌─────────────────────────┐  │
│  │  - Task Manager             │    │   Git Integration       │  │
│  │  - Project Manager          │    │                         │  │
│  │  - State Machine            │    │  - Worktree Manager     │  │
│  └─────────────────────────────┘    │  - Branch Operations    │  │
│                                     │  - PR Management        │  │
│                                     └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
            ┌─────────────┐ ┌─────────┐ ┌─────────────┐
            │ PostgreSQL  │ │  Redis  │ │ File System │
            │  Database   │ │  Cache  │ │ (Worktrees) │
            └─────────────┘ └─────────┘ └─────────────┘

                                │
                                ▼
                    ┌─────────────────────────┐
                    │   External Systems      │
                    ├─────────────────────────┤
                    │  ┌─────────────────┐    │
                    │  │  AI CLI Tools   │    │
                    │  │ (Claude Code)   │    │
                    │  └─────────────────┘    │
                    │                         │
                    │  ┌─────────────────┐    │
                    │  │  Git Repository │    │
                    │  │ (GitHub/GitLab) │    │
                    │  └─────────────────┘    │
                    └─────────────────────────┘
```

## Các thành phần chính

### Frontend Layer

- **Web Browser**: React SPA với TypeScript
- **State Management**: Redux Toolkit cho global state
- **Real-time Updates**: WebSocket connection cho live task updates

### Backend Services Layer

- **API Gateway**: Gin-based REST API server
- **WebSocket Server**: Real-time communication với frontend
- **Core Management Service**: Unified service quản lý cả tasks và projects
  - Task Manager: Quản lý lifecycle và state của tasks
  - Project Manager: Quản lý projects và configurations
  - State Machine: Xử lý task state transitions
- **AI Agent Controller**: Orchestrate AI CLI processes
- **Git Integration**: Quản lý Git operations

### Data Layer

- **PostgreSQL**: Primary data storage
- **Redis**: Caching và session management
- **File System**: Temporary files và Git worktrees

### External Systems

- **AI CLI Tools**: Claude Code CLI và future AI agents
- **Git Repository**: Remote Git repositories (GitHub, GitLab)

## Luồng dữ liệu chính

### Task Execution Flow

1. **Task Creation**: User tạo task qua Web UI → API Gateway → Core Management Service → Database
2. **Planning Phase**: Core Management Service → AI Agent Controller → Spawn Claude CLI → Monitor progress
3. **Implementation Phase**: AI Agent Controller → Git Integration → Create worktree/branch → Execute CLI
4. **Completion**: Git Integration → Create PR → Core Management Service updates task status → Notify user qua WebSocket

### Real-time Updates Flow

1. **Status Change**: Backend service cập nhật task status
2. **Event Publishing**: Publish event qua internal event bus
3. **WebSocket Broadcast**: WebSocket server broadcast update đến connected clients
4. **UI Update**: Frontend receives update và cập nhật UI state

## Giao tiếp giữa các service

### Internal Communication

- **HTTP REST**: Communication giữa frontend và backend
- **WebSocket**: Real-time updates từ backend đến frontend
- **Process IPC**: Communication với spawned CLI processes
- **Database**: Shared data storage giữa các services

### External Communication

- **Git CLI**: Command-line interface với Git operations
- **AI CLI**: Spawned processes với AI coding agents
- **Git Remote**: HTTP/HTTPS với remote repositories

### Message Formats

- **REST API**: JSON với OpenAPI specification
- **WebSocket**: JSON với predefined event types
- **Database**: Structured data với foreign key relationships
- **CLI Integration**: Command-line arguments và environment variables

# Thiết kế chi tiết

## Database Schema

### Projects Table

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    repository_url VARCHAR(500) NOT NULL,
    main_branch VARCHAR(100) DEFAULT 'main',
    ai_config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_projects_name ON projects(name);
CREATE INDEX idx_projects_created_at ON projects(created_at);
```

### Tasks Table

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'TODO',
    plan TEXT,
    branch_name VARCHAR(255),
    pr_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_status CHECK (status IN ('TODO', 'PLANNING', 'PLAN_REVIEWING', 'IMPLEMENTING', 'CODE_REVIEWING', 'DONE', 'CANCELLED'))
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
```

### Executions Table

```sql
CREATE TABLE executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    ai_cli_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    cli_process_pid INTEGER,
    monitor_process_pid INTEGER,
    working_directory VARCHAR(1000),
    cli_command TEXT NOT NULL,
    cli_args JSONB DEFAULT '[]',
    environment_vars JSONB DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    exit_code INTEGER,
    logs TEXT,

    CONSTRAINT valid_execution_status CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled'))
);

CREATE INDEX idx_executions_task_id ON executions(task_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_started_at ON executions(started_at);
```

### Processes Table

```sql
CREATE TABLE processes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    process_type VARCHAR(50) NOT NULL,
    pid INTEGER,
    command TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'starting',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    exit_code INTEGER,
    stdout_log TEXT,
    stderr_log TEXT,

    CONSTRAINT valid_process_type CHECK (process_type IN ('setup', 'cli_agent', 'monitor', 'cleanup')),
    CONSTRAINT valid_process_status CHECK (status IN ('starting', 'running', 'completed', 'failed', 'killed'))
);

CREATE INDEX idx_processes_execution_id ON processes(execution_id);
CREATE INDEX idx_processes_pid ON processes(pid);
CREATE INDEX idx_processes_status ON processes(status);
```

## API Specifications

### Projects API

#### GET /api/v1/projects

**Response:**

```json
{
  "projects": [
    {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "repository_url": "string",
      "main_branch": "string",
      "ai_config": {},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /api/v1/projects

**Request:**

```json
{
  "name": "string",
  "description": "string",
  "repository_url": "string",
  "main_branch": "string",
  "ai_config": {
    "cli_type": "claude-code",
    "model": "claude-3-sonnet",
    "max_tokens": 4096
  }
}
```

### Tasks API

#### GET /api/v1/projects/{project_id}/tasks

**Query Parameters:**

- status: filter by task status
- limit: number of results (default: 50)
- offset: pagination offset

**Response:**

```json
{
  "tasks": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "title": "string",
      "description": "string",
      "status": "TODO",
      "plan": "string",
      "branch_name": "string",
      "pr_url": "string",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "has_more": true
}
```

#### POST /api/v1/projects/{project_id}/tasks

**Request:**

```json
{
  "title": "string",
  "description": "string"
}
```

#### POST /api/v1/tasks/{task_id}/start-planning

**Response:**

```json
{
  "message": "Planning started",
  "execution_id": "uuid"
}
```

#### POST /api/v1/tasks/{task_id}/approve-plan

**Request:**

```json
{
  "approved": true,
  "feedback": "string"
}
```

#### POST /api/v1/tasks/{task_id}/start-implementation

**Response:**

```json
{
  "message": "Implementation started",
  "execution_id": "uuid"
}
```

#### POST /api/v1/tasks/{task_id}/stop-execution

**Description**: Dừng execution đang chạy của task
**Response:**

```json
{
  "message": "Execution stopped",
  "execution_id": "uuid",
  "stopped_at": "2024-01-01T00:00:00Z"
}
```

#### POST /api/v1/executions/{execution_id}/kill

**Description**: Kill process của execution cụ thể
**Response:**

```json
{
  "message": "Execution killed",
  "execution_id": "uuid",
  "killed_processes": [
    {
      "process_id": "uuid",
      "pid": 12345,
      "process_type": "cli_agent"
    }
  ]
}
```

### WebSocket Events

#### Task Status Updates

```json
{
  "type": "task_status_updated",
  "data": {
    "task_id": "uuid",
    "old_status": "PLANNING",
    "new_status": "PLAN_REVIEWING",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

#### Execution Progress

```json
{
  "type": "execution_progress",
  "data": {
    "execution_id": "uuid",
    "task_id": "uuid",
    "progress": 0.6,
    "current_step": "Implementing feature X",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

#### Execution Stopped

```json
{
  "type": "execution_stopped",
  "data": {
    "execution_id": "uuid",
    "task_id": "uuid",
    "reason": "user_requested",
    "stopped_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Process Killed

```json
{
  "type": "process_killed",
  "data": {
    "execution_id": "uuid",
    "task_id": "uuid",
    "process_id": "uuid",
    "pid": 12345,
    "process_type": "cli_agent",
    "killed_at": "2024-01-01T00:00:00Z"
  }
}
```

## Algorithms và Logic phức tạp

### Task State Machine

```go
type TaskStatus string

const (
    StatusTODO           TaskStatus = "TODO"
    StatusPlanning       TaskStatus = "PLANNING"
    StatusPlanReviewing  TaskStatus = "PLAN_REVIEWING"
    StatusImplementing   TaskStatus = "IMPLEMENTING"
    StatusCodeReviewing  TaskStatus = "CODE_REVIEWING"
    StatusDone           TaskStatus = "DONE"
    StatusCancelled      TaskStatus = "CANCELLED"
)

func (t *Task) CanTransitionTo(newStatus TaskStatus) bool {
    transitions := map[TaskStatus][]TaskStatus{
        StatusTODO:          {StatusPlanning, StatusCancelled},
        StatusPlanning:      {StatusPlanReviewing, StatusCancelled},
        StatusPlanReviewing: {StatusImplementing, StatusPlanning, StatusCancelled},
        StatusImplementing:  {StatusCodeReviewing, StatusCancelled},
        StatusCodeReviewing: {StatusDone, StatusCancelled},
        StatusDone:          {},
        StatusCancelled:     {},
    }

    validTransitions := transitions[t.Status]
    for _, validStatus := range validTransitions {
        if validStatus == newStatus {
            return true
        }
    }
    return false
}
```

### AI CLI Process Management

```go
type ProcessManager struct {
    executions map[string]*Execution
    mutex      sync.RWMutex
}

func (pm *ProcessManager) SpawnCLIAgent(exec *Execution) error {
    // 1. Setup working directory
    workDir := pm.createWorktree(exec.TaskID)

    // 2. Prepare CLI command
    cmd := exec.BuildCommand(workDir)

    // 3. Start process
    process := exec.Command("claude-code", cmd.Args...)
    process.Dir = workDir
    process.Env = cmd.Environment

    if err := process.Start(); err != nil {
        return err
    }

    // 4. Start monitoring goroutine
    go pm.monitorProcess(exec, process)

    // 5. Update execution record
    exec.CliProcessPID = process.Process.Pid
    exec.Status = "running"
    exec.StartedAt = time.Now()

    return pm.repository.UpdateExecution(exec)
}

func (pm *ProcessManager) monitorProcess(exec *Execution, process *exec.Cmd) {
    // Monitor stdout/stderr
    stdout, _ := process.StdoutPipe()
    stderr, _ := process.StderrPipe()

    go pm.logOutput(exec.ID, "stdout", stdout)
    go pm.logOutput(exec.ID, "stderr", stderr)

    // Wait for completion
    err := process.Wait()

    // Update execution status
    exec.CompletedAt = time.Now()
    exec.ExitCode = process.ProcessState.ExitCode()

    if err != nil {
        exec.Status = "failed"
    } else {
        exec.Status = "completed"
    }

    pm.repository.UpdateExecution(exec)
    pm.publishExecutionEvent(exec)
}

func (pm *ProcessManager) KillExecution(executionID string) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    exec, exists := pm.executions[executionID]
    if !exists {
        return fmt.Errorf("execution not found: %s", executionID)
    }

    // Get all processes for this execution
    processes, err := pm.repository.GetProcessesByExecution(executionID)
    if err != nil {
        return fmt.Errorf("failed to get processes: %w", err)
    }

    killedProcesses := []ProcessInfo{}

    // Kill all running processes
    for _, proc := range processes {
        if proc.Status == "running" && proc.PID != 0 {
            if err := pm.killProcess(proc.PID); err != nil {
                log.Errorf("Failed to kill process %d: %v", proc.PID, err)
                continue
            }

            // Update process status to killed
            proc.Status = "killed"
            proc.CompletedAt = time.Now()
            proc.ExitCode = -1 // Killed signal

            if err := pm.repository.UpdateProcess(&proc); err != nil {
                log.Errorf("Failed to update process status: %v", err)
            }

            killedProcesses = append(killedProcesses, ProcessInfo{
                ProcessID:   proc.ID,
                PID:        proc.PID,
                ProcessType: proc.ProcessType,
            })

            // Publish process killed event
            pm.publishProcessKilledEvent(exec.TaskID, executionID, &proc)
        }
    }

    // Update execution status
    exec.Status = "cancelled"
    exec.CompletedAt = time.Now()

    if err := pm.repository.UpdateExecution(exec); err != nil {
        return fmt.Errorf("failed to update execution: %w", err)
    }

    // Publish execution stopped event
    pm.publishExecutionStoppedEvent(exec)

    // Remove from active executions
    delete(pm.executions, executionID)

    return nil
}

func (pm *ProcessManager) killProcess(pid int) error {
    process, err := os.FindProcess(pid)
    if err != nil {
        return fmt.Errorf("process not found: %w", err)
    }

    // Try graceful termination first (SIGTERM)
    if err := process.Signal(syscall.SIGTERM); err != nil {
        // If graceful termination fails, force kill (SIGKILL)
        return process.Kill()
    }

    // Wait a bit for graceful shutdown
    timer := time.NewTimer(5 * time.Second)
    defer timer.Stop()

    done := make(chan error, 1)
    go func() {
        _, err := process.Wait()
        done <- err
    }()

    select {
    case <-timer.C:
        // Timeout, force kill
        return process.Kill()
    case err := <-done:
        // Process terminated gracefully
        return err
    }
}

func (pm *ProcessManager) publishExecutionStoppedEvent(exec *Execution) {
    event := ExecutionStoppedEvent{
        Type: "execution_stopped",
        Data: ExecutionStoppedData{
            ExecutionID: exec.ID,
            TaskID:     exec.TaskID,
            Reason:     "user_requested",
            StoppedAt:  time.Now(),
        },
    }
    pm.eventPublisher.Publish("task.execution.stopped", event)
}

func (pm *ProcessManager) publishProcessKilledEvent(taskID, executionID string, proc *Process) {
    event := ProcessKilledEvent{
        Type: "process_killed",
        Data: ProcessKilledData{
            ExecutionID: executionID,
            TaskID:     taskID,
            ProcessID:  proc.ID,
            PID:        proc.PID,
            ProcessType: proc.ProcessType,
            KilledAt:   time.Now(),
        },
    }
    pm.eventPublisher.Publish("task.process.killed", event)
}
```

### Git Worktree Management

```go
func (gm *GitManager) CreateWorktree(taskID, projectRepo, branchName string) (string, error) {
    workDir := filepath.Join(gm.baseDir, "worktrees", taskID)

    // Create worktree
    cmd := exec.Command("git", "worktree", "add", workDir, "-b", branchName)
    cmd.Dir = projectRepo

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("failed to create worktree: %w", err)
    }

    return workDir, nil
}

func (gm *GitManager) CleanupWorktree(taskID string) error {
    workDir := filepath.Join(gm.baseDir, "worktrees", taskID)

    // Remove worktree
    cmd := exec.Command("git", "worktree", "remove", workDir, "--force")

    return cmd.Run()
}
```

## Data Models

### Core Entities

```go
type Project struct {
    ID            string                 `json:"id" db:"id"`
    Name          string                 `json:"name" db:"name"`
    Description   string                 `json:"description" db:"description"`
    RepositoryURL string                 `json:"repository_url" db:"repository_url"`
    MainBranch    string                 `json:"main_branch" db:"main_branch"`
    AIConfig      map[string]interface{} `json:"ai_config" db:"ai_config"`
    CreatedAt     time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

type Task struct {
    ID          string     `json:"id" db:"id"`
    ProjectID   string     `json:"project_id" db:"project_id"`
    Title       string     `json:"title" db:"title"`
    Description string     `json:"description" db:"description"`
    Status      TaskStatus `json:"status" db:"status"`
    Plan        string     `json:"plan" db:"plan"`
    BranchName  string     `json:"branch_name" db:"branch_name"`
    PRURL       string     `json:"pr_url" db:"pr_url"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
    CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

type Execution struct {
    ID               string                 `json:"id" db:"id"`
    TaskID           string                 `json:"task_id" db:"task_id"`
    AICLIType        string                 `json:"ai_cli_type" db:"ai_cli_type"`
    Status           ExecutionStatus        `json:"status" db:"status"`
    CliProcessPID    int                    `json:"cli_process_pid" db:"cli_process_pid"`
    MonitorProcessPID int                   `json:"monitor_process_pid" db:"monitor_process_pid"`
    WorkingDirectory string                 `json:"working_directory" db:"working_directory"`
    CliCommand       string                 `json:"cli_command" db:"cli_command"`
    CliArgs          []string               `json:"cli_args" db:"cli_args"`
    EnvironmentVars  map[string]string      `json:"environment_vars" db:"environment_vars"`
    StartedAt        *time.Time             `json:"started_at" db:"started_at"`
    CompletedAt      *time.Time             `json:"completed_at" db:"completed_at"`
    ExitCode         int                    `json:"exit_code" db:"exit_code"`
    Logs             string                 `json:"logs" db:"logs"`
}

type ProcessInfo struct {
    ProcessID   string `json:"process_id"`
    PID         int    `json:"pid"`
    ProcessType string `json:"process_type"`
}

type ExecutionStoppedEvent struct {
    Type string                `json:"type"`
    Data ExecutionStoppedData  `json:"data"`
}

type ExecutionStoppedData struct {
    ExecutionID string    `json:"execution_id"`
    TaskID      string    `json:"task_id"`
    Reason      string    `json:"reason"`
    StoppedAt   time.Time `json:"stopped_at"`
}

type ProcessKilledEvent struct {
    Type string             `json:"type"`
    Data ProcessKilledData  `json:"data"`
}

type ProcessKilledData struct {
    ExecutionID string    `json:"execution_id"`
    TaskID      string    `json:"task_id"`
    ProcessID   string    `json:"process_id"`
    PID         int       `json:"pid"`
    ProcessType string    `json:"process_type"`
    KilledAt    time.Time `json:"killed_at"`
}
```
