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

**Planning Process (Quy trình lập kế hoạch):**
1. Phân tích yêu cầu task
2. Khảo sát codebase hiện tại
3. Xác định scope thay đổi
4. Lập danh sách các bước thực hiện
5. Ước lượng thời gian và độ phức tạp

**Implementation Process (Quy trình thực hiện):**
1. Checkout và setup worktree
2. Thực hiện code changes theo plan
3. Chạy tests và linting
4. Commit changes với message có ý nghĩa
5. Tạo Pull Request với mô tả chi tiết

**Error Handling (Xử lý lỗi):**
- Retry mechanism cho các lỗi tạm thời
- Fallback strategy khi AI không thể hoàn thành
- Logging chi tiết cho debugging
- Notification cho developer khi cần can thiệp

### 2.2 AI Coding Agent - Tác Nhân AI Lập Trình

**Capabilities (Khả năng):**
- Code generation và modification
- Test writing và execution
- Code review và optimization
- Documentation generation
- Bug fixing và refactoring

**Configuration (Cấu hình):**
- Model selection (GPT-4, Claude, v.v.)
- Temperature và creativity settings
- Context window management
- Custom prompts cho từng loại task

**Integration (Tích hợp):**
- IDE/Editor plugins
- CI/CD pipeline hooks
- Code review tools
- Project management systems

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

## 4. Implementation Guidelines - Hướng Dẫn Triển Khai

### 4.1 Database Design
- Sử dụng PostgreSQL cho data persistence
- Redis cho caching và session management
- Migration scripts cho schema changes

### 4.2 API Design
- RESTful endpoints cho CRUD operations
- GraphQL cho complex queries
- WebSocket cho real-time notifications

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
