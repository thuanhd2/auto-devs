# Core Concepts - Các Khái Niệm Cốt Lõi

Hệ thống tự động hóa task cho developer bao gồm các khái niệm cốt lõi sau:

## 1. Task Management - Quản Lý Task

### 1.1 Project & Task - Dự Án & Nhiệm Vụ

**Project (Dự án):**
- Là container chứa các task liên quan đến một codebase cụ thể
- Mỗi project được cấu hình với:
  - Repository Git URL
  - Branch chính (main/master)
  - Cấu hình AI agent (model, prompts)
  - Quyền truy cập và authentication

**Task (Nhiệm vụ):**
- Đơn vị công việc cụ thể cần được thực hiện
- Có các trạng thái (status) theo workflow:
  - `TODO`: Task mới được tạo, chờ bắt đầu
  - `PLANNING`: AI đang phân tích và lập kế hoạch
  - `PLAN_REVIEWING`: Kế hoạch đã sẵn sàng, chờ developer review
  - `IMPLEMENTING`: Đang thực hiện code
  - `CODE_REVIEWING`: Code đã hoàn thành, chờ review và merge PR
  - `DONE`: Task hoàn thành
  - `CANCELLED`: Task bị hủy

### 1.2 Task Execution - Thực Thi Task

**Task Lifecycle (Vòng đời task):**

1. **Tạo Task**: Developer tạo task mới với mô tả yêu cầu
2. **Start Planning**: Kích hoạt AI agent để phân tích và lập kế hoạch
3. **Plan Review**: Developer xem xét kế hoạch, có thể:
   - Approve: Chấp nhận và chuyển sang implement
   - Reject: Từ chối và yêu cầu plan lại
   - Cancel: Hủy task
4. **Start Implementation**: AI agent bắt đầu code theo kế hoạch
5. **Code Review**: Tạo Pull Request, chờ review và merge
6. **Complete**: Task chuyển sang DONE sau khi PR được merge

**Task Context (Ngữ cảnh task):**
- Thông tin về codebase hiện tại
- Dependencies và requirements
- Test cases cần pass
- Code style và conventions

### 1.3 Git Worktree - Quản Lý Git Worktree

**Mục đích:**
- Mỗi task được thực hiện trên branch riêng biệt
- Tránh conflict giữa các task đang chạy đồng thời
- Đảm bảo isolation và rollback dễ dàng

**Implementation:**
- Sử dụng Git worktree để tạo working directory riêng cho mỗi task
- Branch naming convention: `task-{task_id}-{slug}`
- Tự động cleanup worktree sau khi task hoàn thành

## 2. AI Executor - Bộ Thực Thi AI

### 2.1 Process - Quy Trình Xử Lý

**System Role (Vai trò hệ thống):**
- Hệ thống chỉ orchestrate và observe, không implement AI logic
- Tất cả AI capabilities được delegate cho external CLI tools
- System chịu trách nhiệm về lifecycle management và monitoring

**Task Execution Orchestration (Điều phối thực thi task):**
1. **Setup Phase**: 
   - Tạo worktree và setup environment
   - Chạy pre-run scripts nếu có
   - Chuẩn bị working directory
2. **CLI Spawn Phase**:
   - Spawn AI CLI process với appropriate command
   - Spawn monitor process để track progress
   - Setup logging và communication channels
3. **Monitoring Phase**:
   - Theo dõi process status và health
   - Parse logs để extract progress information
   - Update task status dựa trên CLI output
4. **Completion Phase**:
   - Detect khi CLI đã hoàn thành task
   - Collect results và artifacts
   - Cleanup processes và temporary resources

**Status Observation (Quan sát trạng thái):**
- Monitor CLI process PID và exit codes
- Parse stdout/stderr để determine progress
- Detect error patterns trong CLI output
- Track file system changes để confirm completion

**Error Handling (Xử lý lỗi):**
- Retry mechanism khi CLI process fails unexpectedly
- Graceful shutdown cho hanging processes
- Detailed logging cho debugging CLI issues
- Notification system cho developer intervention

### 2.2 AI Coding Agent - Tác Nhân AI Lập Trình

**Architecture (Kiến trúc):**
- Tất cả AI coding agent đều là CLI tools (command-line interface)
- Các CLI được hỗ trợ: claude-code, google-gemini-cli, qwen-coder, v.v.
- Hệ thống không implement logic AI, chỉ orchestrate và observe các CLI

**Process Execution (Thực thi process):**
- Khi AI Executor cần execute task, spawn 2 processes:
  1. Process khởi động AI CLI agent
  2. Process follow-up để monitor status và progress
- System theo dõi trạng thái CLI để xác định:
  - Task đã hoàn thành hay chưa
  - Có lỗi xảy ra hay không
  - Progress hiện tại của task

**CLI Integration Strategy (Chiến lược tích hợp CLI):**
- Plugin-based architecture để dễ dàng thêm CLI agent mới
- MVP milestone: chỉ hỗ trợ Claude Code CLI
- Extensible design cho future CLI integrations
- Standardized interface để communicate với các CLI khác nhau

**Configuration (Cấu hình):**
- CLI selection per project
- Command-line arguments cho từng CLI
- Environment variables setup
- Working directory management

## 3. System Architecture - Kiến Trúc Hệ Thống

### 3.1 Core Components

**Task Manager:**
- REST API cho CRUD operations
- WebSocket cho real-time updates
- Database cho persistence

**AI Agent Controller:**
- Queue management cho task processing
- Resource allocation và scaling
- Monitoring và metrics

**Git Integration:**
- Worktree management
- Branch operations
- PR creation và tracking

### 3.2 Data Models

**Project Model:**
```
- id: string
- name: string
- repository_url: string
- main_branch: string
- ai_config: object
- created_at: datetime
```

**Task Model:**
```
- id: string
- project_id: string
- title: string
- description: string
- status: enum
- plan: text
- branch_name: string
- pr_url: string
- created_at: datetime
- updated_at: datetime
```

**Execution Model:**
```
- id: string
- task_id: string
- ai_cli_type: string (claude-code, gemini-cli, qwen-coder)
- status: enum (queued, running, completed, failed, cancelled)
- cli_process_pid: integer
- monitor_process_pid: integer
- working_directory: string
- cli_command: string
- cli_args: json
- environment_vars: json
- started_at: datetime
- completed_at: datetime
- exit_code: integer
- logs: text
```

**Process Model:**
```
- id: string
- execution_id: string
- process_type: enum (setup, cli_agent, monitor, cleanup)
- pid: integer
- command: string
- status: enum (starting, running, completed, failed, killed)
- started_at: datetime
- completed_at: datetime
- exit_code: integer
- stdout_log: text
- stderr_log: text
```

**PreRunScript Model:**
```
- id: string
- project_id: string
- name: string
- script_content: text
- script_type: enum (bash, python, node)
- execution_order: integer
- is_active: boolean
- created_at: datetime
```

## 4. Implementation Guidelines - Hướng Dẫn Triển Khai

### 4.1 Database Design
- Sử dụng PostgreSQL cho data persistence
- Redis cho caching và session management
- Migration scripts cho schema changes

### 4.2 API Design
- RESTful endpoints cho CRUD operations
- WebSocket cho real-time notifications
- JSON response format cho tất cả endpoints

### 4.3 Security
- Authentication với JWT tokens
- Authorization dựa trên project ownership
- Input validation và sanitization
- Rate limiting cho AI API calls

### 4.4 Scalability
- Horizontal scaling với load balancers
- Queue system cho background jobs
- Database sharding nếu cần
- CDN cho static assets
