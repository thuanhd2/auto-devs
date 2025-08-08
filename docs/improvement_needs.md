### PR – Thời điểm tạo và các cải tiến cần làm

#### 1) PR được tạo khi nào (hiện tại vs. kỳ vọng)
- Kỳ vọng: Ngay sau khi hoàn tất IMPLEMENTING (AI thực thi xong, branch của task đã có commit/push), hệ thống tạo PR từ nhánh của task lên nhánh chính (mặc định `main`).
- Thực tế code hiện tại:
  - Đã có `PRCreator` với phương thức `CreatePRFromImplementation(...)` để tạo PR (sinh title/body, gọi GitHub API, chèn link Task vào PR).
  - Chưa “wire” `PRCreator` vào luồng thực thi/hoàn tất IMPLEMENTING để tạo PR tự động.
  - Chưa có nhóm REST API backend cho Pull Requests, trong khi frontend đã gọi các endpoint `/pull-requests`.
  - WebSocket: backend gửi `type: "pr_update"` kèm `data.type` chi tiết; frontend lại lắng theo `message.type === 'pr_created' | 'pr_merged' | 'pr_closed'`, dẫn tới lệch contract.

Các tệp liên quan:
- Tạo PR: `internal/service/github/pr_creator.go`
- GitHub API: `internal/service/github/github_service.go`, `internal/service/github/repository_operations.go`
- Repo PR: `internal/repository/pull_request.go`, `internal/repository/postgres/pull_request_repository.go`
- Frontend API/hooks: `frontend/src/lib/api/pull-requests.ts`, `frontend/src/hooks/use-pull-requests.ts`, `frontend/src/types/pull-request.ts`, `frontend/src/config/api.ts`
- Router backend: `internal/handler/route.go`
- WebSocket type: `internal/websocket/message.go`

#### 2) Flow quản lý PR hiện tại (đã có)
- PR Monitor (`internal/service/github/pr_monitor.go`):
  - Poll GitHub PR theo chu kỳ; đồng bộ trạng thái PR về DB.
  - Map trạng thái PR → trạng thái Task: OPEN → CODE_REVIEWING, MERGED → DONE, CLOSED (không merge) → giữ nguyên hoặc CANCELLED khi đang CODE_REVIEWING.
  - Khi PR merged: cập nhật Task → DONE, trigger cleanup worktree, gửi WebSocket thông báo.
- PR Sync Worker (`internal/service/github/pr_sync_worker.go`):
  - Chạy nền theo interval, batch + concurrency, retry; đồng bộ và bắn thông báo tương tự Monitor (fallback khi thiếu realtime).

#### 3) Vấn đề tồn tại
- Chưa nối `PRCreator` vào luồng hoàn tất IMPLEMENTING để tự động tạo PR và bắt đầu monitor ngay.
- Thiếu REST API cho PR để FE sử dụng (list/filter, get, get by task, create, update, sync, merge, close, reopen).
- Lệch WebSocket contract giữa backend và frontend (event type).
- Thiếu GitHub Webhook để nhận realtime events (opened, synchronize, closed/merged, review, checks); hiện phụ thuộc polling nhiều.
- Cấu hình base branch cứng `main`, chưa theo `project.main_branch`.
- Merge/close flows chưa được expose qua API/usecase (dù `GitHubService.MergePullRequest` đã có).
- Độ tin cậy: thiếu idempotency/retry cho tạo PR; cần metrics/rate-limit observability.
- Cleanup worktree: đã có trong Monitor/Sync Worker khi merge; cần đảm bảo an toàn và logging đầy đủ.
- Bảo mật cấu hình GitHub: token, validate repo/token trước khi chạy.
- UI/UX: FE đã có khung PR list/detail/actions nhưng cần dữ liệu từ API backend thống nhất, cùng checks/reviews/comments.

#### 4) Đề xuất cải tiến (ưu tiên)
1) Tạo PR tự động sau IMPLEMENTING hoàn tất
- Tại thời điểm `ExecutionStatusCompleted` của task:
  - Gọi `PRCreator.CreatePRFromImplementation` để tạo PR (base từ cấu hình project), thêm task links.
  - Lưu PR vào DB (`pull_requests`), cập nhật `task.pr_url`/`branch_name` nếu cần.
  - Cập nhật Task → `CODE_REVIEWING` (nếu chưa), bắn WebSocket “pr_created”.
  - Bắt đầu monitor PR (Monitor service) ngay khi tạo xong.
  - Idempotent: kiểm tra đã có PR OPEN cho `task_id`/`head_branch` trước khi tạo mới.

2) Bổ sung REST API Pull Requests (đồng bộ với FE hiện có)
- Endpoints (gợi ý):
  - `GET /pull-requests?project_id&status&...` (list/filter/paging)
  - `GET /pull-requests/{id}`
  - `GET /tasks/{task_id}/pull-request`
  - `POST /pull-requests` (tuỳ chọn cho manual create)
  - `PUT /pull-requests/{id}` (update trường metadata)
  - `POST /pull-requests/{id}/sync`
  - `POST /pull-requests/{id}/merge` (merge/squash/rebase)
  - `POST /pull-requests/{id}/close` / `POST /pull-requests/{id}/reopen`
- Tầng usecase + repository dùng `internal/repository/pull_request.go` và `postgres/pull_request_repository.go`.

3) Đồng bộ WebSocket contract
- Chọn 1:
  - Backend gửi `message.type` cụ thể theo sự kiện: `pr_created`, `pr_updated`, `pr_merged`, `pr_closed` (thân thiện với FE hiện tại).
  - Hoặc giữ `message.type = 'pr_update'` và FE phân nhánh theo `message.data.type`.

4) Tích hợp GitHub Webhook (realtime, giảm phụ thuộc polling)
- Handle events: PR opened/synchronize/closed/merged, review, check-suite/status.
- Xác thực HMAC, idempotent theo delivery id, cập nhật DB + bắn WebSocket.
- Kết hợp với Sync Worker làm fallback.

5) Cấu hình và metadata PR
- Base branch per project (`project.main_branch`), không cứng `main`.
- Hỗ trợ PR draft, reviewers/labels/assignees, áp template, auto-link Task (đã có `AddTaskLinks`).

6) Thao tác merge/close robust qua API
- Dùng `GitHubService.MergePullRequest` với 3 phương thức; kiểm tra trạng thái PR (draft/conflict) trước merge; cập nhật DB và bắn WS.
- Close/reopen: `UpdatePullRequest state`, đồng bộ DB/WS.

7) Quan sát & ổn định
- Retry/idempotency cho tạo PR.
- Metrics/logging cho PR create/sync/merge, rate-limit; alert khi lỗi tăng.

8) Dọn worktree an toàn
- Giữ hành vi cleanup khi merged; log lỗi không chặn flow, bảo đảm không xoá nhầm.

9) Bảo mật & cấu hình
- Quản lý token GitHub (secret), validate token/repo trước thao tác (`ValidateToken`, `ValidateRepository`).

10) UI/UX
- Cập nhật FE để dùng API backend mới; hiển thị checks/reviews/comments; sửa WebSocket theo quyết định ở mục (3).

#### 5) Tóm tắt ngắn
- Hiện PR “được thiết kế” để tạo ngay sau IMPLEMENTING nhưng chưa được tích hợp vào flow; thiếu API backend và lệch WS.
- Monitor/Sync đã tốt: đồng bộ trạng thái PR → Task, cleanup worktree, WS.
- Ưu tiên: wire tạo PR tự động sau IMPLEMENTING, thêm API PR, đồng bộ WebSocket, bổ sung Webhook GitHub; sau đó hoàn thiện merge/close, cấu hình base branch, và quan sát/bảo mật.
