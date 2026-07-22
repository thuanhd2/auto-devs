---
name: auto-devs-workflow
description: "Manage work via Auto Devs — the user's custom AI-powered development workflow platform. Log tasks, don't code yourself."
version: 1.7.1
author: Hermes Agent
metadata:
  hermes:
    tags: [auto-devs, task-management, ai-executor, kanban, workflow, kanban-bridge]
---

# Auto Devs Workflow

The user (Thuan Ho) built and uses **Auto Devs** daily — an AI-powered development workflow platform at **http://localhost:8888**. It automates coding tasks with AI agents (Claude Code).

> 🚨 **🚨 GOLDEN RULE — Auto Devs First, ALWAYS 🚨**
>
> Khi user yêu cầu **bất kỳ** feature dev / code change / sửa bug nào:
> 1. **KHÔNG** explore codebase — không đọc file, không phân tích, không grep
> 2. **KHÔNG** load `plan` skill, không viết markdown plan, không propose approach
> 3. **KHÔNG** làm bất kỳ investigation nào trước khi log task
> 4. **Hành động DUY NHẤT:** xác định project → tạo branch → tạo kanban card → log task (kèm `kanban_task_id`) → start-planning → block card (xem **Kanban Bridge**)
>
> Dù user có hỏi "plan của bạn là gì?" trước — câu trả lời vẫn là "tôi sẽ log task vào Auto Devs".
> Đừng load `plan` skill rồi viết plan cho user xem — việc đó thuộc về Claude Code bên trong Auto Devs.
> AI executor (Claude Code + Cursor Agent) đã được user setup để làm tất cả. Làm thủ công == phí thời gian và đã bị correct nhiều lần.
>
> **Nếu user hỏi "plan" trước khi bạn kịp log task:** reply ngay "tôi sẽ log task vào Auto Devs, AI sẽ tự plan", không hỏi lại, không propose approach.

## Projects & When to Use Them

| Project | Git Repo | Purpose |
|---------|----------|---------|
| **SCEX** | thuanhd2/scex (monorepo + submodules) | **Tài liệu, nghiên cứu**, viết PRD/BRD/technical design. Chứa toàn bộ platform (dax-be, dax-fe, fiat-service...). Không implement code ở đây. |
| **Dax Be** | thuanhd2/dax-be | Backend Go (API, orders, wallets, WebSocket). Task implementation. |
| **Dax FE** | lpex-fe-source | Frontend React Native + Web. |
| **Dax Kyc Service** | thuanhd2/dax-kyc-service | KYC identity verification. |
| **Fiat Service** | thuanhd2/fiat-service | Bank deposit/withdraw. |
| **LPEX Reporting** | thuanhd2/lpex-reporting-service | Reporting service. |
| **Dax Admin SSO** | thuanhd2/dax-admin-sso | Admin SSO portal. |
| **Azasend Za Be** | thuanhd2/azasend-za-be | Azasend backend. |
| **Auto Devs** | thuanhd2/auto-devs | Fix Auto Devs itself. |
| **Test auto dev** | thuanhd2/test-auto-devs | Testing Auto Devs workflow. |
| **Markdown Viewer Extension** | thuanhd2/markdown-viewer-extension | Markdown Viewer Chrome/Firefox/Edge extension. Branch pattern: `feat/md-to-docx-cli` |
| **Release Manager** | thuanhd2/release-manager | Release management dashboard (React + Vite + Supabase). Project ID: `f2b47692-0702-4ca3-b8c2-ab4640f96104`. Base branch: `master`. |

## Creating a New Project

There is **no MCP tool** (`project:create` does not exist) and **no REST API endpoint** for creating projects. You must use the **browser UI** at `http://localhost:8888/projects`.

Steps:
1. Click **"New Project"** button
2. Fill in the dialog fields:
   - **Project Name** — display name (e.g., "Markdown Viewer Extension")
   - **Description** — short description
   - **Worktree Base Path** — absolute path to the local repo directory
   - **Init Workspace Script** — optional, leave blank unless an init script exists
3. Click **"Create Project"**
4. The `repository_url` is **auto-detected** from the git remote in the worktree path (uses `origin` remote)

> **Note:** There is no `repository_url` field in the create dialog. Auto Devs reads it automatically from the git config at the worktree path.

After creating the project, create tasks via MCP `task:create` tool or REST API as usual.

## Standard Workflow

### Before creating a task — create branch

1. **Xác định base remote và base branch:**
   - Base remote: nếu có remote `company` thì dùng `company`, ngược lại là `origin`
   - Base branch: nếu remote có branch `main` thì dùng `main`, ngược lại là `master`

2. **Tạo branch từ base, push lên origin:**

```bash
cd /path/to/repo
git fetch <base_remote> <base_branch>
git checkout -b my-task-branch
git reset <base_remote>/<base_branch> --hard
git push origin my-task-branch
```

Ví dụ với dax-be (có remote company, base branch main):
```bash
cd /Users/thuanho/Documents/scex/dax-be
git fetch company main
git checkout -b feat/my-task
git reset company/main --hard
git push origin feat/my-task
```

### Creating the task

Create via REST API (port 8888) rather than the SPA browser UI — more reliable:

```bash
curl -s -X POST "http://localhost:8888/api/v1/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "<project-uuid>",
    "title": "Short descriptive title",
    "description": "Concise description with just enough context for the AI to plan."
  }'
```

**Title format:** `<verb> <file/thing> <context>` — ngắn gọn, đủ ý. Ví dụ: `Update /docs/apis/api_docs.md 1.0.6`

**Description rules (QUAN TRỌNG — user đã correct lỗi verbose task):**
- **CHỈ nói WHAT, không nói HOW** — không viết plan hộ Claude, không include git steps, branch strategy, repo paths
- **Tối đa 2-3 câu** — concise, súc tích
- **Có thể include GitHub link** — user dùng link PR trực tiếp, Auto Devs xử lý được
- **Include output expectation cụ thể** nếu cần — ví dụ: "có breaking change nào thì note vào file breaking_changes_1.0.6.md"
- **Không bao gồm**: repo path, remote info, git branch strategy, numbered steps, markdown sections — Auto Devs và AI executor tự biết project context
- **Nếu là SCEX project, không cần nói file ở submodule nào** — AI tự tìm

**Lỗi thường gặp (đã từng gây hậu quả):**
❌ Title dài dòng có "(Copy)" suffix — title bị rác
❌ Description có markdown sections (## Mục tiêu, ## Yêu cầu...) — quá verbose, confused AI
❌ Bao gồm repo path, remote config, git command — thông tin thừa
❌ Viết numbered steps execution plan — Claude bị confused, dẫn đến xoá toàn bộ files trong worktree
❌ Bao gồm repo path, remote config, git command trong description — chỉ gây confused cho AI
❌ Viết numbered steps execution plan — Claude bị confused (đã từng dẫn đến xoá toàn bộ worktree files)

### Starting planning

Preferred — use MCP tool:
```
mcp__auto_devs__task_start_planning(taskId, branchName="my-feature-branch", aiType="claude-code", useRemoteBranch=true, autoImplement=false)
```

Fallback — REST API:
```bash
curl -s -X POST "http://localhost:8888/api/v1/tasks/<task-id>/start-planning" \
  -H "Content-Type: application/json" \
  -d '{
    "branch_name": "my-feature-branch",
    "ai_type": "claude-code",
    "use_remote_branch": true
  }'
```

**`branch_name` là feature branch, KHÔNG phải base branch.** Backend lưu `branch_name` vào `BaseBranchName` của task (dùng làm base cho PR sau này). Worktree sẽ tự tạo generated branch (`task-<uuid>-...`) để Claude Code làm việc. Auto Devs sau đó tự tạo PR từ generated branch → feature branch.

**Không truyền `base_branch_name`** — field này tồn tại trong DTO nhưng handler bỏ qua, không forward vào usecase. Chỉ có `branch_name` mới có hiệu lực.

Xác định branch name phù hợp:
- Branch mới tạo (step 2) → dùng chính branch đó
- Nếu task đã có branch → dùng branch đó

### Multiple Tasks (Parallel)

Khi user đưa ra **nhiều feature request cùng lúc** (vd: "sửa A, thêm B, thêm C" trong cùng 1 tin nhắn):

1. **Tạo branch riêng cho mỗi task** — mỗi task 1 branch, không gộp
2. **Tạo task riêng cho mỗi feature** — mỗi task độc lập
3. **Start planning cho tất cả** — gọi start-planning cho từng task ngay sau khi tạo (không cần đợi task trước xong planning)
4. **Mỗi task 1 kanban card riêng** (xem Kanban Bridge) — block độc lập; callback của
   từng task tự unblock đúng card của nó, không ảnh hưởng task khác
5. **Review plan từng task khi chúng chuyển sang PLAN_REVIEWING** — mỗi task có thể review và approve riêng
6. Mỗi task sẽ tạo PR riêng vào feature branch của nó

### Monitoring planning progress

After calling start-planning:
1. Task status goes to `PLANNING`
2. An execution record is created with status `PENDING` — **this stays PENDING** even while Claude is actively working
3. The actual Claude process (e.g. `claude -p --permission-mode=plan`) runs as a separate OS process
4. Check real progress by inspecting the worktree git log:
   ```bash
   # Worktree path pattern:
   # /Users/thuanho/autodevs/project-<project-id>/task-<task-id>
   git -C /Users/thuanho/autodevs/project-<project-id>/task-<task-id> log --oneline -5
   ```
5. When planning completes, status moves to `PLAN_REVIEWING`
6. **KHÔNG poll.** Auto Devs tự callback vào kanban board `autodevs` khi status đổi
   (xem **Kanban Bridge**). Chỉ inspect worktree git log (bước 4) khi user chủ động hỏi tiến độ.

### After planning completes

1. Check status: task moves to `PLAN_REVIEWING`
2. Review plan via API: `GET /api/v1/tasks/<task-id>/plans`
3. **Show the plan to user** — present the options (solutions, proposals) with pros/cons, let user pick.

4. **Handle user modifications** — If user chooses a solution but provides custom requirements/modifications:
   a. **Update task description** with the chosen solution + specific requirements before approving:
      ```bash
      curl -s -X PUT "http://localhost:8888/api/v1/tasks/<task-id>" \
        -H "Content-Type: application/json" \
        -d '{"description": "<original description + solution chosen + specific requirements>"}'
      ```
   b. This ensures the implementation agent has the full context of what the user wants.
   c. **Important:** Use PUT, not PATCH — the backend treats PUT as full update.

5. Approve using MCP tool (preferred) — **dùng `cursor-agent` (hyphen) cho implementation:**
   ```
   mcp__auto_devs__task_approve_plan(taskId, aiType="cursor-agent")
   ```
6. If MCP tool is unavailable, use REST API (send `cursor-agent` with hyphen!):
   ```bash
   curl -s -X POST "http://localhost:8888/api/v1/tasks/<task-id>/approve-plan" \
     -H "Content-Type: application/json" \
     -d '{"ai_type": "cursor-agent"}'
   ```
7. Task moves to `IMPLEMENTING`
8. Wait for `CODE_REVIEWING` status
9. Check the worktree for generated files

### AI Types & Workflow Split

**Workflow (theo chỉ định của user):**
- **Planning** → luôn dùng **`claude-code`**
- **Implementation** → luôn dùng **`cursor_agent`** (concept name), nhưng API value là **`cursor-agent`** (hyphen) — xem lưu ý bên dưới

| Phase | AI Type (concept) | API value (gửi lên backend) | Notes |
|-------|----------|----------|-------|
| Planning | `claude-code` | `claude-code` | Dùng `npx @anthropic-ai/claude-code` |
| Implementation | `cursor_agent` | **`cursor-agent`** (hyphen!) | Dùng Cursor agent (không bị quota limit) |

> **⚠️ Lưu ý:** Backend processor's `getAiExecutor()` expects `"cursor-agent"` (hyphen) — **luôn gửi `"cursor-agent"`** khi gọi API, không gửi `"cursor_agent"` (underscore) vì sẽ fail ở switch case, task stuck IMPLEMENTING không execution.

Nếu implementation với cursor_agent fail, có thể retry với claude-code nếu cần.

### Starting implementation directly (nếu plan đã approve nhưng implementation fail)

Preferred — MCP tool:
```
mcp__auto_devs__task_start_implementing_direct(taskId, branchName="feature-branch-name", aiType="cursor-agent", useRemoteBranch=false)
```

Fallback — REST API:
```bash
curl -s -X POST "http://localhost:8888/api/v1/tasks/<task-id>/start-implementing-direct" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_type": "cursor-agent",
    "branch_name": "feature-branch-name"
  }'
```

Lưu ý: `branch_name` là **required field** cho API start-implementing-direct.

### Post-Implementation: Code Review & Merge (via GitHub)

Sau khi AI finish implementation (task status → `CODE_REVIEWING`):

1. **Lấy Pull Request từ task** — dùng GitHub MCP để lấy PR info:
   ```
   mcp__github__get_pull_request
   ```
   PR thường được Auto Devs tạo tự động. Nếu chưa có PR, kiểm tra bằng `github:search-issues`.

2. **Gửi link PR cho user review trước** — user yêu cầu xem PR trước khi merge. KHÔNG tự merge. Gửi link PR và chờ user phản hồi.

3. **Review PR diff nếu user yêu cầu** — dùng GitHub MCP:
   - `mcp__github__get_pull_request_files` — xem files changed
   - `mcp__github__get_pull_request` — xem status

3. **Merge PR** — dùng GitHub MCP:
   ```
   mcp__github__merge_pull_request
   ```
   - merge_method: "squash" hoặc "merge"
   - Sau merge, task chuyển: `CODE_REVIEWING` → `DONE`

4. **Pull code về local và merge vào base branch:**
   ```bash
   git checkout task-branch-name
   git pull origin task-branch-name
   git checkout base_branch
   git merge task-branch-name
   ```

### Alternative: Manual Merge (nếu Auto Devs không tạo PR)

Nếu task không có PR, dùng cách thủ công:

1. **Inspect the worktree** at worktree path
2. **Review the commit(s)** — `git log --oneline -5` then `git diff HEAD~1 --stat`
3. **Clean up Claude Code artifacts** — Claude Code may create `.serena/` directory:
   ```bash
   git reset --soft HEAD~1
   git reset HEAD .serena/   # unstage .serena/
   git commit -m "<original message>"
   git push --force origin <branch>
   ```
4. **Merge to the real branch** — the task's worktree branch is `task-<uuid>-<slug>`. Merge it into `feat/<task-slug>`:
   ```bash
   cd ~/Documents/personal/<repo>
   git checkout feat/<branch-name>
   git fetch origin task-<uuid>-<slug>
   git merge origin/task-<uuid>-<slug> --no-ff -m "Merge: <title>"
   git push origin feat/<branch-name>
   ```
5. **Test the implementation** — build and run with real project files. For markdown-related tools, use files from the scex project:
   ```bash
   # scex has rich markdown files at:
   ~/Documents/scex/docs/brds/*.md
   ~/Documents/scex/docs/prds/*.md
   ~/Documents/scex/docs/technical-designs/**/*.md
   ~/Documents/scex/docs/apis/*.md
   ```
6. **Fix any issues** found during testing — commit fixes to the feat branch
7. **Push** the final feat branch

## Kanban Bridge — KHÔNG chờ Auto Devs trong session

Nguyên tắc: session chat KHÔNG BAO GIỜ chờ/poll Auto Devs. Mọi chờ đợi đi qua
kanban board `autodevs`. Auto Devs sẽ tự callback (comment + unblock card) khi
task chuyển `PLAN_REVIEWING` / `CODE_REVIEWING` / `DONE` / `CANCELLED`.

### Khi giao task — thứ tự BẮT BUỘC: kanban_create → task:create → start-planning

(Cần `card_id` trước khi tạo task, vì `kanban_task_id` chỉ set được lúc create task, không update sau được.)

1. `kanban_create` trên board `autodevs`, assignee `default` (profile mặc định):
   - title: `[AD] <task title>`
   - body CHỈ chứa data (playbook nằm trong skill này, không nhét instructions vào card):
     ```
     project: <name> (<project_id>)
     branch: <feature-branch>
     awaiting: PLAN_REVIEWING
     notify: <platform>:<chat_id>
     ```
2. **QUAN TRỌNG: tạo task Auto-Devs PHẢI truyền `kanban_task_id=<card_id>`**
   (param có trong cả MCP `task:create` lẫn REST `POST /api/v1/tasks` — snake_case).
3. Ghi task uuid lên card: `kanban_comment(task_id=<card_id>, body="autodevs_task_id: <task-uuid>")`
   (body card không edit được bằng tool, dùng comment).
4. `task_start_planning` như workflow chuẩn.
5. `kanban_block(kind=needs_input, reason="waiting auto-devs callback")`
   — **PHẢI dùng `needs_input`, KHÔNG dùng `dependency`**: kind dependency đưa card
   về `todo` và dispatcher sẽ tự promote card có assignee → spawn worker sớm vô nghĩa.
   `needs_input` giữ card ở `blocked` cho đến khi Auto Devs callback PATCH `ready`.
6. Báo user "đã giao task, sẽ báo khi có kết quả" → KẾT THÚC TURN. Không poll, không sleep.

### Khi được spawn làm worker (env `HERMES_KANBAN_TASK` được set)

1. `kanban_show` → đọc body + comment MỚI NHẤT dạng `[auto-devs] status=...`
   (comment có đủ: status, task uuid + title, old_status, pr, error).
   **Nếu KHÔNG có comment `[auto-devs]` nào mới hơn lần xử lý trước (spawn sớm do race):
   chỉ `kanban_comment` ghi chú trạng thái rồi `kanban_block(kind=needs_input,
   reason="waiting auto-devs callback")` — TUYỆT ĐỐI KHÔNG approve plan, KHÔNG merge,
   KHÔNG `kanban_complete`.**
2. **Nhắn user bằng `hermes send` qua terminal** (worker KHÔNG có tool `send_message`):
   ```bash
   hermes send "<message>" -t <target theo dòng notify: trong body, vd telegram:1428166637>
   ```
3. Xử lý theo status trong comment:
   - `PLAN_REVIEWING`: `GET /api/v1/tasks/<uuid>/plans` → tóm tắt options pros/cons →
     `hermes send` cho user → `kanban_block(kind=needs_input, reason="user reviewing plan")`
   - `CODE_REVIEWING`: lấy PR link (dòng `pr:` trong comment, hoặc github MCP) →
     `hermes send` link cho user, KHÔNG tự merge →
     `kanban_block(kind=needs_input, reason="user reviewing PR")`
   - `DONE`: verify PR merged → `hermes send` tổng kết → `kanban_complete(summary=...)`
   - `CANCELLED` / failed: đọc dòng `error:` → `hermes send` báo lỗi + gợi ý
     (quota limit → gợi ý retry cursor-agent) → `kanban_block(kind=needs_input)`
4. Worker KHÔNG tự approve plan, KHÔNG tự merge PR — quyết định thuộc user (qua chat).

### Khi user phản hồi trong chat (approve plan / yêu cầu merge / sửa yêu cầu)

1. Thực hiện action như workflow chuẩn (task_approve_plan / merge PR / update description).
2. **BẮT BUỘC trước khi kết thúc turn:** tìm card tương ứng
   (`kanban_list` board `autodevs`, match title `[AD]` hoặc comment `autodevs_task_id`) →
   `kanban_comment` cập nhật `awaiting: <status kế tiếp>` →
   `kanban_block(kind=needs_input, reason="waiting auto-devs callback")` để chờ callback vòng sau
   (needs_input, không dùng dependency — xem lý do ở trên).
   Callback của Auto Devs sẽ tự PATCH card về `ready` khi status đổi.

## MCP Auto-Devs Tools

All MCP tools are available after MCP server rebuild + `/reload-mcp`:

```
project:list, project:get
task:list, task:create, task:get, task:update-status, task:delete
task:start-planning, task:approve-plan, task:start-implementing-direct
execution:list, execution:get, execution:create
worktree:get-status
```

**Important — use MCP tools, NOT curl, for workflow actions (planning, approve, implement):**
- ✅ `mcp__auto_devs__task_start_planning`
- ✅ `mcp__auto_devs__task_approve_plan`
- ✅ `mcp__auto_devs__task_start_implementing_direct`

MCP tools map to the correct route handler with proper error handling. Curl bypasses the MCP layer and can produce different results. Only fall back to curl/API when the MCP tools return errors.

If MCP tools are not visible, run `/reload-mcp` first. If still missing, rebuild MCP server:
```bash
cd ~/Documents/personal/auto-devs/mcp-server && npm run build
```

**Note:** MCP server defaults to port 8098 (separate dev instance with only 4 projects). The production instance is on port 8888 with all 16 projects. If MCP tools can't find a project, use the REST API on port 8888 instead.

## References

- `references/auto-devs-api.md` — Full REST API reference with endpoints, payloads, and MCP tool equivalents.
- `references/monitoring-progress.md` — How to check real-time task progress when execution status is misleading.
- `references/reviewing-large-prs-for-api-docs.md` — How to extract API changes from PRs >300 files (where `gh pr diff` fails) for updating api_docs.md and detecting breaking changes.
- `references/pr-creation-status.md` — PR creation works automatically; what to do if you don't see one.
- `references/release-manager-deployment.md` — Release Manager deploy: pull, dry-run migration check (STOP and ask user if pending), build with deploy-s3-cloudfront.sh.

## MCP Server Config (Fixed)

The MCP server config in `~/.hermes/config.yaml` MUST include `cwd` and `env`:

```yaml
mcp_servers:
  auto-devs:
    command: node
    args:
      - /Users/thuanho/Documents/personal/auto-devs/mcp-server/dist/index.js
    cwd: /Users/thuanho/Documents/personal/auto-devs/mcp-server
    env:
      AUTO_DEVS_API_URL: http://localhost:8888
```

Without this, the MCP server defaults to port 8098 (a separate dev instance with only 4 projects). Port 8888 has all 16 projects.

After changing config, run `/reload-mcp` in-session.

## MCP Bug Fixed: JSON Field Names

The MCP client was sending `projectId` (camelCase) but Go backend expects `project_id` (snake_case). Fix applied to:
- `autodevs-client.ts`: `projectId` → `project_id: projectId` in `createTask()`
- `autodevs-client.ts`: `taskId` → `task_id: taskId` in `createExecution()`

## Branch Default: `master` not `main`

On thuanhd2's fork (remote `origin`), the default branch is `master`, not `main`:
```bash
git push origin origin/master:refs/heads/my-branch
```

## Daily Cron Job

Cron job at 8am daily reads Slack reminders + saved items + Kanban board and sends summary to user via Telegram. Script at `~/.hermes/scripts/slack-reader.py` reads Slack API.

## Known Pitfalls

### ⚠️ Branch phải tạo trong repo của project, không phải repo trong description

Khi tạo branch cho task, dùng thông tin từ `project:get` (worktree_base_path) để biết chính xác repo nào. Task thuộc project nào thì branch tạo trong repo của project đó, không phải repo được mention trong description.

Ví dụ: SCEX project có worktree_base_path là `/Users/thuanho/Documents/scex` → branch tạo trong `/Users/thuanho/Documents/scex`. Dù task mô tả nói về `dax-be`.

### ⚠️ Branch phải tồn tại trên origin TRƯỚC khi start-planning với `use_remote_branch: true`

Khi gọi `start-planning` với `use_remote_branch: true`, backend cố gắng tạo worktree từ `origin/<branch_name>`. Nếu branch chưa được push lên origin, worktree creation fail với:
```
fatal: not a valid object name: 'origin/<branch_name>'
```

**Workaround:** Tạo branch local, push lên origin **trước** khi start-planning:
```bash
cd <repo-path>
git fetch <remote> <base_branch>
git checkout -b my-branch
git reset <remote>/<base_branch> --hard
git push origin my-branch
# Sau đó mới gọi start-planning với branch_name="my-branch", use_remote_branch=true
```

### ⚠️ Research/investigation tasks dùng branch prefix `research/`, SCEX project

Khi task là **research/investigation** (không thay đổi code — chỉ phân tích codebase, viết report):
- **SCEX project** là nơi phù hợp (dùng cho "Tài liệu, nghiên cứu")
- Branch name dùng prefix `research/` thay vì `feat/` (vd: `research/websocket-bandwidth`)
- Report output được viết vào `docs/technical-designs/<topic>.md`
- Không cần code change, không cần build/test
- AI sẽ tạo PR từ worktree branch → research branch. User review PR và merge để lưu report.

### ⚠️ Kiểm tra PR sau CODE_REVIEWING

Auto Devs tự động tạo PR từ worktree branch (`task-<uuid>-...`) vào feature branch (`branch_name` khi start-planning). PR được tạo trong cùng goroutine với completion handler.

Nếu chưa thấy PR ngay khi task ở CODE_REVIEWING:
- Đợi vài giây — worker tạo PR đồng bộ
- Nếu chưa thấy sau 1-2 phút, kiểm tra bằng `github:search-issues`:
  ```
  repo:<owner>/<repo> <task-uuid-prefix> type:pr
  ```
- **KHÔNG tự tạo PR thủ công** — nếu PR thực sự không được tạo, log task fix vào Auto Devs project

### ⚠️ KNOWN BUG: `ai_type` mismatch — `cursor-agent` vs `cursor_agent`

**Bug location:** `internal/jobs/processor.go` — `getAiExecutor()` switch case (line 317)

The Go backend's `getAiExecutor()` switch expects `"cursor-agent"` (hyphen), but the caller (whether MCP tool, curl, or frontend) may send `"cursor_agent"` (underscore). When `"cursor_agent"` is sent, it falls through to `default`, returning `"invalid execution type: cursor_agent"`.

**Symptoms when this bug triggers:**
- Task status changes to `IMPLEMENTING` but no execution record is created
- Worker logs show `"Failed to get AI executor"`
- No fallback status revert (task stuck at `IMPLEMENTING`)
- `start-implementing-direct` returns HTTP 500

**Workaround (current):** Luôn gửi `"cursor-agent"` (hyphen) khi gọi API — cả MCP tool lẫn curl.

**Permanent fix (chọn 1 trong 2):**
- **Option A:** Sửa Go backend: thêm `case "cursor_agent":` cùng handler với `case "cursor-agent":` trong `getAiExecutor()` → chấp nhận cả 2 định dạng
- **Option B:** Sửa Go backend: đổi `case "cursor-agent":` → `case "cursor_agent":` và rebuild worker

`branch_name` truyền vào start-planning được backend lưu vào `BaseBranchName` của task. Sau đó:
1. Worktree tạo generated branch (`task-<uuid>-...`)
2. Claude Code code trên generated branch
3. Auto Devs tạo PR: generated branch → feature branch (`branch_name`)

Nghĩa là `branch_name` quyết định PR sẽ nhắm vào branch nào. Chọn đúng branch (feature branch đã push lên origin) để PR có base chính xác.

### SPA Browser Dialog Issues
The Auto Devs web UI is a React SPA. Browser interactions (snapshots, clicks) can lose state:
- Dialogs may not render in snapshots
- Navigation between dialogs may break
- The approval confirmation dialog ("Approve Plan and Start Implementation") appears as a **separate overlay** — clicking "Approve Plan and Start Implement" on the task view opens it, then you must click "Approve and Start Implementation" in the overlay
- **Fix:** Use the REST API directly via curl instead of fighting with the browser

### Execution Status Stays PENDING During Planning/Implementation
After starting planning or implementation, the execution record (from `execution:list` or `execution:get`) may show `PENDING` status even though Claude Code is actively running as an OS process. This is normal — the execution record updates only on completion/failure. Do not re-trigger planning thinking it's stuck. Monitor the **task status** (which transitions to `PLANNING`, `IMPLEMENTING`, etc.) rather than the execution status.

### Claude Code Auth Failures
Auto Devs runs `npx @anthropic-ai/claude-code` via the worker. If planning fails with 401:
- The user needs to re-authenticate Claude Code
- The issue is that `npx` Claude Code has separate auth from the CLI `claude` command
- After re-auth, click "Start Planning" again or restart via API
### Claude Code Session / Quota Limit

Claude Code có **daily session limit**. Khi planning hoặc implementation fail với `exit code: 1` không rõ lý do, kiểm tra:

```bash
# Check CLI claude (dùng bởi Hermes)
cd <worktree>
echo 'test' | claude -p --print

# Check npx claude (dùng bởi Auto Devs worker - QUOTA KHÁC!)
npx --yes @anthropic-ai/claude-code -p "hi" --print
```

Nếu thấy `"You've hit your session limit · resets <time>"`:
- **DO NOT cancel the task** — quota tự reset theo thời gian của nó
- **DO NOT retry với claude-code ngay** — sẽ fail lại
- **Option A (nhanh hơn)**: Retry implementation với `cursor-agent` thay vì claude-code (không bị quota limit) — dùng `start-implementing-direct` API với `"ai_type": "cursor-agent"` (hyphen, không underscore)
- **Option B**: Set cron job retry sau 10 phút kể từ thời gian reset

### Auto Devs Uses `origin` Remote
The worktree clones use `origin` → thuanhd2's fork. Push branches to `origin`, NOT `company`.

### Claude Code Creates `.serena/` Directory
Claude Code auto-creates a `.serena/` directory with project metadata (Claude's code graph memory). This directory **must not be committed** to the repo:
- When reviewing Claude Code's commit, check for `.serena/` in `git diff --stat`
- If present, use `git reset --soft HEAD~1` then `git reset HEAD .serena/` to unstage it before recommitting
- The `.serena/` directory itself can remain as an untracked file — it doesn't affect the working tree
- Always force-push after cleaning up, since the commit hash changes

### Theme/Path Mismatches in Generated Help Text
When Claude Code generates CLI tools, it may include incorrect or non-existent examples in help text (e.g., `--theme github-dark` when that theme doesn't exist). Always test the CLI with --help or --verbose flags to verify examples match actual available values.

### `gh pr diff` Fails for Large PRs (>300 files)
GitHub's diff endpoint returns HTTP 406 for PRs exceeding ~300 files. When reviewing release PRs (common for SCEX releases), use local git workflow instead:
```bash
git fetch company  # or the upstream remote
git diff main...branch-name -- controllers/ routers/ --stat
```
See `references/reviewing-large-prs-for-api-docs.md` for the full workflow.
