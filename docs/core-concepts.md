# Core Concepts

Vibe Kanban là một hệ thống quản lý task tự động sử dụng AI coding agents để thực hiện các công việc phát triển phần mềm. Hệ thống được thiết kế dựa trên hai khái niệm chính:

## 1. Task Management (Quản lý Task)

### 1.1 Projects (Dự án)
- **Định nghĩa**: Đại diện cho các Git repositories kèm theo các script phát triển
- **Cấu hình bao gồm**:
  - Đường dẫn repository
  - Setup scripts (script thiết lập)
  - Development scripts (script phát triển)
  - Cleanup scripts (script dọn dẹp)

### 1.2 Tasks (Nhiệm vụ)
Tasks có vòng đời rõ ràng với các trạng thái:
- **`Todo`**: Trạng thái ban đầu, sẵn sàng thực thi
- **`InProgress`**: Task đang được thực hiện
- **`InReview`**: Task đã hoàn thành, chờ review
- **`Done`**: Task hoàn thành và được chấp nhận
- **`Cancelled`**: Task bị hủy bỏ

### 1.3 Task Attempts (Lần thử thực hiện Task)
- **Môi trường cô lập**: Mỗi attempt chạy trong git worktree riêng biệt
- **Mỗi attempt bao gồm**:
  - Branch riêng biệt với naming convention: `vk-{short_uuid}-{task_title}`
  - Đường dẫn worktree độc lập
  - Process thực thi riêng
  - GitHub pull request (tùy chọn)

### 1.4 Git Worktree
- **Mục đích**: Cô lập môi trường thực thi để tránh conflict
- **Đặc điểm**:
  - Sử dụng base directory theo platform
  - Hỗ trợ các thao tác git phức tạp (branch, rebase, merge)
  - Có thể resume task sau khi cleanup ("cold tasks")

### 1.5 Workflow Thực thi
1. **Task Creation**: Tạo task mới
2. **Worktree Initialization**: Khởi tạo worktree
3. **Setup Script Execution**: Chạy script thiết lập
4. **AI Coding Agent Execution**: AI thực hiện coding
5. **Changes Review**: Review các thay đổi
6. **Merge/Cleanup**: Merge hoặc dọn dẹp (tùy chọn)

## 2. AI Executor System (Hệ thống AI Executor)

### 2.1 Kiến trúc Executor
- **Core Executor Trait**: Interface chuẩn hóa cho việc quản lý các AI coding agents
- **Tính linh hoạt**: Cho phép chuyển đổi giữa các AI services khác nhau

### 2.2 Các loại Executor được hỗ trợ
1. **Claude Executor**: Sử dụng Anthropic Claude API
2. **Gemini Executor**: Sử dụng Google Gemini API  
3. **AMP Executor**: Executor chuyên biệt cho AMP
4. **OpenCode Executors**:
   - SST (Serverless Toolkit)
   - Charm
   - Claude Code Router (CCR)

### 2.3 Quản lý Executor
- **Executor Factory**: Tạo động các executor theo cấu hình
- **Process Lifecycle**: Quản lý toàn bộ vòng đời (spawn, monitor, cleanup)
- **MCP Integration**: Tích hợp Model Context Protocol cho khả năng mở rộng

### 2.4 Workflow Thực thi AI
1. **Cấu hình Executor**: Xác định AI agent sẽ sử dụng
2. **ProcessService**: Spawn executor phù hợp
3. **Real-time Streaming**: Capture tương tác với AI agent
4. **Background Monitoring**: Theo dõi quá trình hoàn thành

### 2.5 Tính năng đặc biệt
- **Session Continuity**: Duy trì session cho follow-up executions
- **Real-time Logging**: Ghi log và streaming output theo thời gian thực
- **Multi-process Support**: Hỗ trợ nhiều loại process và execution states

## 3. Tích hợp và Mở rộng

### 3.1 MCP Server Integration
- **JSON-RPC Interface**: Giao diện chuẩn cho external tools
- **Advanced Configuration**: Cấu hình nâng cao cho AI coding agents
- **Extensibility**: Dễ dàng thêm AI agents mới

### 3.2 GitHub Integration
- **Pull Request**: Tự động tạo PR cho các thay đổi
- **Branch Management**: Quản lý branch theo naming convention
- **Remote Tracking**: Theo dõi branch trên remote repository

### 3.3 Multi-platform Support
- **Cross-platform**: Hoạt động trên nhiều hệ điều hành
- **Platform-specific**: Tối ưu hóa theo từng platform
- **Isolation**: Đảm bảo cô lập môi trường trên mọi platform

## 4. Nguyên tắc Thiết kế

### 4.1 Isolation (Cô lập)
- Mỗi task attempt hoạt động trong môi trường độc lập
- Tránh conflict giữa các task đang chạy song song
- Dễ dàng cleanup và rollback

### 4.2 Flexibility (Linh hoạt)
- Hỗ trợ nhiều AI coding agents
- Có thể thêm executor mới mà không ảnh hưởng hệ thống
- Cấu hình linh hoạt theo từng project

### 4.3 Monitoring (Giám sát)
- Theo dõi real-time quá trình thực thi
- Log chi tiết cho debugging
- State management đầy đủ

### 4.4 Scalability (Khả năng mở rộng)
- Thiết kế trait-based cho tính nhất quán
- Factory pattern cho việc tạo executor
- MCP protocol cho tích hợp external tools

## 5. Implementation Guidelines

### 5.1 Để implement Task Management:
1. Thiết lập Git worktree cho isolation
2. Implement state machine cho task lifecycle
3. Tạo naming convention cho branches
4. Xây dựng cleanup mechanism

### 5.2 Để implement AI Executor:
1. Định nghĩa Executor trait với interface chuẩn
2. Implement các executor cụ thể cho từng AI service
3. Tạo Executor Factory cho dynamic instantiation
4. Xây dựng ProcessService cho lifecycle management

### 5.3 Để tích hợp MCP:
1. Implement JSON-RPC interface
2. Định nghĩa protocol cho external tool communication
3. Xây dựng configuration management
4. Tạo extension points cho custom tools