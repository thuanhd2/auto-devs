# Implement Plan: Auto-Devs → Hermes Kanban Callback

## Bối cảnh & mục tiêu

Hermes agent giao task cho Auto-Devs nhưng không thể chờ task chạy dài trong session; khi task xong thì session cũ đã chết, không resume được. Giải pháp: **Kanban bridge** — mọi trạng thái chờ nằm trong Hermes Kanban board (durable, SQLite), Auto-Devs callback vào Kanban HTTP API khi task đổi trạng thái, dispatcher của Hermes tự spawn worker xử lý tiếp. Không phụ thuộc session nào.

**Flow tổng thể:**

```
[User chat] → Hermes tạo kanban card + tạo task Auto-Devs (kèm kanban_task_id)
            → kanban_block(dependency) → kết thúc turn, session chết thoải mái

[Auto-Devs] task đổi status (PLAN_REVIEWING / CODE_REVIEWING / DONE / CANCELLED)
            → enqueue asynq job "kanban:notify"
            → job handler: POST comment + PATCH unblock vào Hermes Kanban API

[Hermes]    dispatcher tick (60s) thấy card ready → spawn worker (profile MẶC ĐỊNH — không tạo profile riêng)
            → worker đọc card + comment → send_message cho user / hoàn tất card
```

**Quyết định đã chốt:**
- Callback bắn cho cả 4 status: `PLAN_REVIEWING`, `CODE_REVIEWING`, `DONE`, `CANCELLED`.
- Bật/tắt hoàn toàn qua env (`HERMES_KANBAN_ENABLED`). Instance không config env → mọi code path skip êm, không side effect. Dev (8098) sẽ không config; prod (8888) có config.
- **Không tạo worker profile riêng** — card gán cho profile hermes mặc định. Lý do: profile riêng (`hermes profile create`) có home riêng tại `~/.hermes/profiles/<name>` với skills riêng → SKILL.md phải sync 2 nơi. Dùng profile mặc định thì dispatcher spawn worker với chính home `~/.hermes` → cùng 1 file skill, sửa 1 nơi. Trade-off (chấp nhận): worker mang full toolset của profile chính.

---

## Phần 1 — Backend Auto-Devs (Go)

### 1.1. Migration

Tạo migration mới (số tiếp theo sau `000020`):

```bash
make migrate-create name=add_kanban_task_id_to_tasks
```

`000021_add_kanban_task_id_to_tasks.up.sql`:
```sql
ALTER TABLE tasks ADD COLUMN kanban_task_id VARCHAR(64);
CREATE INDEX idx_tasks_kanban_task_id ON tasks(kanban_task_id) WHERE kanban_task_id IS NOT NULL;
```

`000021_add_kanban_task_id_to_tasks.down.sql`:
```sql
DROP INDEX IF EXISTS idx_tasks_kanban_task_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS kanban_task_id;
```

### 1.2. Entity

`internal/entity/task.go` — thêm field vào struct `Task`:

```go
KanbanTaskID *string `json:"kanban_task_id,omitempty" gorm:"size:64"`
```

### 1.3. DTO + Handler

- `internal/handler/dto/task.go`: thêm `KanbanTaskID *string \`json:"kanban_task_id,omitempty"\`` vào `CreateTaskRequest` và vào task response DTO.
- Handler create task: pass field xuống usecase → repository (GORM tự map).
- **Lưu ý:** REST `POST /api/v1/tasks` phải nhận được field này vì skill hermes tạo task bằng curl (không chỉ MCP).
- Regen swagger: `make swagger`.

### 1.4. Service mới: `internal/service/kanban`

File `internal/service/kanban/client.go`:

```go
type Client interface {
    // CommentTask đăng comment markdown lên card
    CommentTask(ctx context.Context, kanbanTaskID string, body string) error
    // UnblockTask chuyển card về trạng thái ready để dispatcher spawn worker
    UnblockTask(ctx context.Context, kanbanTaskID string) error
    // Enabled cho biết feature có được config không
    Enabled() bool
}
```

Implementation (`httpClient`):
- Config đọc từ env:

| Env | Ý nghĩa | Ví dụ |
|---|---|---|
| `HERMES_KANBAN_ENABLED` | bật/tắt feature | `true` |
| `HERMES_KANBAN_URL` | base URL hermes dashboard | `http://localhost:3000` |
| `HERMES_KANBAN_TOKEN` | bearer token (= `HERMES_DASHBOARD_SESSION_TOKEN` phía hermes) | `xxx` |

- Endpoints (đã verify trong source hermes-agent `plugins/kanban/dashboard/plugin_api.py`):
  - Comment: `POST {base}/api/plugins/kanban/tasks/{id}/comments` — body JSON `{"body": "<markdown>"}` (kiểm tra lại tên field chính xác trong plugin_api.py, có thể là `text`/`body`)
  - Unblock: `PATCH {base}/api/plugins/kanban/tasks/{id}` — body JSON `{"status": "ready"}`
  - Auth header: `Authorization: Bearer <token>`
- HTTP timeout 15s. Không tự retry trong client — retry do asynq lo (1.5).
- Khi `HERMES_KANBAN_ENABLED != true` → `Enabled()` = false, mọi method no-op trả nil.
- Thêm vào `.env.example` cả 3 biến, kèm comment.

### 1.5. Asynq job mới: `kanban:notify`

`internal/jobs/types.go`:

```go
const TypeKanbanNotify = "kanban:notify"

type KanbanNotifyPayload struct {
    TaskID       uuid.UUID         `json:"task_id"`
    KanbanTaskID string            `json:"kanban_task_id"`
    OldStatus    entity.TaskStatus `json:"old_status"`
    NewStatus    entity.TaskStatus `json:"new_status"`
}

func NewKanbanNotifyTask(p KanbanNotifyPayload) (*asynq.Task, error) { ... }
```

Options khi enqueue: `asynq.MaxRetry(10)`, `asynq.Queue("default")` — exponential backoff mặc định của asynq là đủ (dashboard hermes chết tạm thời sẽ tự retry).

**Handler** (đăng ký trong mux ở `internal/jobs/server.go`, cạnh các handler hiện có):
1. Load task từ DB (lấy title, plan/PR info mới nhất).
2. Build comment markdown theo format máy-đọc-được (worker hermes sẽ parse):

```
[auto-devs] status=PLAN_REVIEWING
task: <uuid> — <title>
old_status: PLANNING
plans: GET /api/v1/tasks/<uuid>/plans
pr: <url hoặc none>
error: <error_logs tail nếu CANCELLED, ngược lại none>
```

3. Gọi `kanbanClient.CommentTask(...)` rồi `kanbanClient.UnblockTask(...)`. Lỗi ở bước nào → return error để asynq retry. **Comment idempotent-friendly:** nếu retry làm comment trùng thì chấp nhận được (worker đọc comment mới nhất), không cần dedup phức tạp.

### 1.6. Điểm hook enqueue

Điểm hội tụ mọi status transition là **usecase layer** (đã verify: `processor.updateTaskStatus` → `taskUsecase.UpdateStatus`; handler flows → `UpdateStatusWithHistory`):

- `internal/usecase/task.go` → `UpdateStatus()` (line ~431) và `UpdateStatusWithHistory()` (line ~459).
- Sau khi update DB thành công, nếu thoả **cả 3 điều kiện** thì enqueue `kanban:notify`:
  1. `oldStatus != newStatus`
  2. `newStatus ∈ {PLAN_REVIEWING, CODE_REVIEWING, DONE, CANCELLED}`
  3. `task.KanbanTaskID != nil && *task.KanbanTaskID != ""`
- Enqueue fail → **log warning, KHÔNG fail transition** (callback là best-effort, không được chặn workflow chính).
- `taskUsecase` cần thêm dependency: asynq client (đã có `internal/jobs/client.go`) — inject qua Wire. Nếu muốn tránh usecase phụ thuộc jobs package, tạo interface nhỏ `KanbanNotifyEnqueuer` trong usecase, implement ở jobs.
- Chú ý cả `BulkUpdateStatus` nếu nó không đi qua 2 method trên.

### 1.7. DI + mocks + tests

- `internal/di/wire.go`: provider cho `kanban.Client` (đọc env) + inject vào usecase/jobs. Chạy `make wire`.
- `make mocks` cho interface mới.
- Tests:
  - `kanban/client_test.go`: httptest server — happy path, 401, 500, disabled mode.
  - `usecase/task_test.go`: enqueue đúng khi đủ 3 điều kiện, không enqueue khi thiếu, enqueue fail không làm fail transition.
  - Job handler test: mock kanban client, verify comment format + unblock được gọi, error → return error (để asynq retry).

### 1.8. Thứ tự thực hiện Phần 1

migration → entity → DTO/handler → service kanban → jobs types/handler → hook usecase → wire → mocks → swagger → tests → `make build && make build-worker`.

---

## Phần 2 — MCP Server (TypeScript)

`mcp-server/src/tools/task-tools.ts`:
- `task:create`: thêm optional param `kanban_task_id` (string) vào JSON schema + description "Hermes kanban card ID for callback".
- `mcp-server/src/client/autodevs-client.ts` (`createTask`): gửi `kanban_task_id: kanbanTaskId` — **snake_case** (đã từng có bug camelCase, xem SKILL.md section "MCP Bug Fixed").
- Rebuild: `cd mcp-server && npm run build`. Phía hermes chạy `/reload-mcp`.

---

## Phần 3 — Hạ tầng Hermes (setup một lần, làm TRƯỚC khi test end-to-end)

1. **Board:** `hermes kanban boards create autodevs` (dispatcher mặc định đã chạy trong gateway: `kanban.dispatch_in_gateway: true`).
2. **Assignee = profile mặc định (KHÔNG tạo profile mới):**
   - Xác định tên profile mặc định: `echo $HERMES_PROFILE` (hoặc xem docs/`hermes status`; kanban docs xác nhận "active default profile" là assignee hợp lệ — orchestrator mặc định chính là nó).
   - Verify profile mặc định đã có đủ: skill `auto-devs-workflow`, toolset kanban (enable `kanban` trong toolsets config nếu chưa), `send_message`, MCP `auto-devs` (port 8888). Thiếu gì bổ sung vào config profile chính.
   - Worker được dispatcher spawn (`hermes -p <default> --cli`) sẽ tự có kanban tools qua env `HERMES_KANBAN_TASK`.
3. **Token cố định cho dashboard:** set env `HERMES_DASHBOARD_SESSION_TOKEN=<token>` cho process hermes dashboard (mặc định token random mỗi lần start — phải fix cứng để Auto-Devs gọi được). Dashboard phải chạy thường trực.
4. **Notification:** `hermes kanban nsub autodevs telegram` (hoặc platform anh dùng) để biến động card tự báo về kênh chat.
5. **Smoke test bằng tay (quan trọng — làm trước khi tin hệ thống):**

```bash
# Tạo card + block (assignee = tên profile mặc định xác định ở bước 2)
hermes kanban create "[AD] smoke test" --board autodevs --assignee <default-profile>
hermes kanban block <card_id> --kind dependency --reason "waiting auto-devs"

# Giả lập Auto-Devs callback
curl -X POST "http://localhost:<port>/api/plugins/kanban/tasks/<card_id>/comments" \
  -H "Authorization: Bearer $HERMES_KANBAN_TOKEN" -H "Content-Type: application/json" \
  -d '{"body": "[auto-devs] status=DONE\ntask: test — smoke\npr: none"}'
curl -X PATCH "http://localhost:<port>/api/plugins/kanban/tasks/<card_id>" \
  -H "Authorization: Bearer $HERMES_KANBAN_TOKEN" -H "Content-Type: application/json" \
  -d '{"status": "ready"}'

# Verify: dispatcher spawn worker trong ~60s, worker đọc được card
hermes kanban tail <card_id>
```

Nếu auth fail hoặc dispatcher không nhặt card → sửa hạ tầng trước, đừng debug bằng code Go.

---

## Phần 4 — Sửa SKILL.md (`~/.hermes/skills/software-development/auto-devs-workflow/SKILL.md`)

### 4.1. Golden Rule — thêm bước 5

Trong chuỗi "Hành động DUY NHẤT", sửa thành: xác định project → tạo branch → log task → start-planning → **tạo kanban card + block (xem Kanban Bridge)**.

### 4.2. Section mới: `## Kanban Bridge — KHÔNG chờ Auto Devs trong session`

Đặt ngay sau "## Standard Workflow". Nội dung:

```markdown
## Kanban Bridge — KHÔNG chờ Auto Devs trong session

Nguyên tắc: session chat KHÔNG BAO GIỜ chờ/poll Auto Devs. Mọi chờ đợi đi qua
kanban board `autodevs`. Auto Devs sẽ tự callback (comment + unblock card) khi
task chuyển PLAN_REVIEWING / CODE_REVIEWING / DONE / CANCELLED.

### Sau khi start-planning (hoặc bất kỳ action nào đẩy task vào trạng thái chạy dài)

1. `kanban_create` trên board `autodevs`, assignee = profile mặc định (cùng profile đang chạy):
   - title: `[AD] <task title>`
   - body CHỈ chứa data (playbook nằm trong skill này, không nhét instructions vào card):
     ```
     autodevs_task_id: <uuid>
     project: <name> (<project_id>)
     branch: <feature-branch>
     awaiting: PLAN_REVIEWING
     notify: <platform>:<chat_id>
     ```
2. **QUAN TRỌNG: tạo task Auto-Devs PHẢI truyền `kanban_task_id=<card_id>`**
   (param mới trong cả MCP `task:create` lẫn REST POST /api/v1/tasks).
   → Vì cần card_id trước, thứ tự đúng là: kanban_create → task:create → start-planning.
3. `kanban_block(kind=dependency, reason="waiting auto-devs callback")`
4. Báo user "đã giao task, sẽ báo khi có kết quả" → KẾT THÚC TURN. Không poll, không sleep.

### Khi được spawn làm worker (env HERMES_KANBAN_TASK được set)

1. `kanban_show` → đọc body + comment MỚI NHẤT dạng `[auto-devs] status=...`
2. Xử lý theo status:
   - `PLAN_REVIEWING`: GET /api/v1/tasks/<id>/plans → tóm tắt options pros/cons →
     `send_message` cho user (theo `notify:` trong body) →
     `kanban_block(kind=needs_input, reason="user reviewing plan")`
   - `CODE_REVIEWING`: lấy PR link (github MCP hoặc comment) → `send_message` link
     cho user, KHÔNG tự merge → `kanban_block(kind=needs_input, reason="user reviewing PR")`
   - `DONE`: verify PR merged → `send_message` tổng kết → `kanban_complete(summary=...)`
   - `CANCELLED` / failed: đọc `error:` → `send_message` báo lỗi + gợi ý
     (quota limit → gợi ý retry cursor-agent) → `kanban_block(kind=needs_input)`
3. Worker KHÔNG tự approve plan, KHÔNG tự merge PR — quyết định thuộc user (qua chat).

### Khi user phản hồi trong chat (approve plan / yêu cầu merge / sửa yêu cầu)

1. Thực hiện action như workflow chuẩn (task_approve_plan / merge PR / update description).
2. **BẮT BUỘC trước khi kết thúc turn:** tìm card tương ứng
   (`kanban_list` board `autodevs`, match `autodevs_task_id` trong body) →
   cập nhật dòng `awaiting:` → `kanban_block(kind=dependency)` để chờ callback vòng sau.
   Quên bước này = card kẹt ở needs_input, callback tiếp theo vẫn unblock được
   nhưng flow kém rõ ràng.
```

### 4.3. Sửa section "Monitoring planning progress"

- Xoá "poll the task API periodically".
- Thay bằng: "KHÔNG poll. Auto Devs callback vào kanban khi status đổi. Chỉ inspect worktree git log khi user chủ động hỏi tiến độ."

### 4.4. Sửa "Multiple Tasks (Parallel)"

Thêm: mỗi task 1 kanban card riêng, block độc lập; callback từng task tự unblock card của nó.

### 4.5. Bump version

`version: 1.6.0` → `1.7.0`, thêm tag `kanban-bridge`.

---

## Phần 5 — End-to-end test trên môi trường thật

Test duy nhất, chạy trên môi trường thật sau khi deploy đầy đủ. Không cần test giả lập.

### 5.1. Deploy trước khi test

1. **Auto-Devs (prod 8888):** `make build && make build-worker` → restart server + worker với env mới:
   ```
   HERMES_KANBAN_ENABLED=true
   HERMES_KANBAN_URL=<hermes dashboard url>
   HERMES_KANBAN_TOKEN=<token>
   ```
2. **MCP server:** `cd mcp-server && npm run build`.
3. **Hermes:** sửa SKILL.md xong → restart hermes gateway + dashboard (dashboard chạy với `HERMES_DASHBOARD_SESSION_TOKEN` cố định). Trong session chạy `/reload-mcp` nếu cần.
4. Chạy migration: `make migrate-up`.

### 5.2. Kịch bản test

Dùng project **Test auto dev** (`thuanhd2/test-auto-devs`). Task đơn giản: **ghi "xin chào" vào file `hi.md`**.

Bắt đầu bằng cách nhắn hermes (qua kênh chat bình thường):

> "Tạo task trong project Test auto dev: ghi 'xin chào' vào file hi.md"

Sau đó theo dõi từng checkpoint:

| # | Bước | Hành vi mong đợi |
|---|---|---|
| 1 | Hermes nhận yêu cầu | Tạo branch → `kanban_create` card trên board `autodevs` (body có `autodevs_task_id`, `awaiting`, `notify`) → tạo task Auto-Devs **kèm `kanban_task_id`** → start-planning → `kanban_block(dependency)` → trả lời "đã giao task" và KẾT THÚC TURN. **Không poll.** |
| 2 | Auto-Devs planning xong → `PLAN_REVIEWING` | Asynq job `kanban:notify` chạy: comment `[auto-devs] status=PLAN_REVIEWING` xuất hiện trên card + card chuyển `ready`. Check: board UI hoặc `hermes kanban show <card_id>`. |
| 3 | Dispatcher tick (≤60s) | Worker được spawn (profile mặc định). Nó đọc card, GET plans, **send_message tóm tắt plan cho user**, rồi `kanban_block(needs_input)`. |
| 4 | User reply trong chat: "approve đi" | Session chat approve plan (`task_approve_plan`, ai_type `cursor-agent`) → tìm lại card, cập nhật `awaiting` → `kanban_block(dependency)`. |
| 5 | Auto-Devs implement xong → `CODE_REVIEWING` | Callback lần 2: comment mới + card `ready` → worker spawn → **send_message link PR cho user** → `kanban_block(needs_input)`. KHÔNG tự merge. |
| 6 | User reply: "merge đi" | Session chat merge PR qua github MCP → task chuyển `DONE`. Card được cập nhật + block(dependency) chờ callback DONE. |
| 7 | Callback `DONE` | Comment lần 3 + card `ready` → worker spawn → verify hi.md có "xin chào" trên branch → **send_message tổng kết** → `kanban_complete`. Card sang cột Done. |

### 5.3. Tiêu chí pass

- Cả 3 callback (PLAN_REVIEWING, CODE_REVIEWING, DONE) đều tạo comment + unblock card đúng.
- User nhận đủ 3 tin nhắn chủ động từ hermes (plan review, PR link, tổng kết) **mà không cần session nào sống chờ**.
- File `hi.md` chứa "xin chào" sau khi merge.
- Không có error lạ trong log worker Auto-Devs và asynq (job `kanban:notify` không retry liên tục).

### 5.4. Debug khi fail

- Callback không đến: check asynq queue/dead queue + log worker Auto-Devs → thường là token/URL sai hoặc dashboard chưa chạy.
- Card ready nhưng không có worker: check dispatcher log trong gateway (`hermes kanban tail <card_id>`), assignee có đúng tên profile mặc định không.
- Worker spawn nhưng hành xử sai: đọc transcript session worker (`hermes kanban show <card_id>` xem run) → thường do SKILL.md section Kanban Bridge chưa rõ hoặc comment format thiếu field.

## Ghi chú rủi ro

- **Tên field body comment API**: verify chính xác schema pydantic trong `plugins/kanban/dashboard/plugin_api.py` (router `POST /tasks/{task_id}/comments`) trước khi code client Go.
- **PATCH status hợp lệ**: verify `blocked → ready` được PATCH endpoint chấp nhận (nếu không, dùng transition qua `todo` hoặc endpoint unblock tương đương — check `hermes kanban unblock` dùng code path nào trong `kanban_db`).
- **Dashboard là single point**: callback phụ thuộc dashboard process sống. Asynq retry 10 lần backoff exponential cover được restart ngắn; nếu dashboard down lâu, comment sẽ đến khi nó sống lại.
- **Card bị xoá/archive trước khi callback**: API trả 404 → job retry vô ích → sau max retry job vào dead queue, log warning. Chấp nhận được.
