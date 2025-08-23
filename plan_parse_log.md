# Plan: Parse Execution Log Structure Backend

## Current State Analysis

From analyzing the codebase, I understand the current log processing flow:

### Frontend Log Processing (Current State):

- Frontend receives raw log messages as strings in `ExecutionLog.message` field
- Frontend components (`execution-logs-modal.tsx`, `execution-logs-pannel.tsx`) parse JSON from raw message strings using `JSON.parse(message)`
- Frontend handles different log types: `user`, `assistant`, `tool_result`, `result` with complex rendering logic
- Each log type has specific formatting and display rules implemented in the frontend

### Backend Current Structure:

- `ExecutionLog` entity in `internal/entity/execution_log.go` stores logs with basic fields
- `message` field stores raw string content
- `metadata` JSONB field exists but appears underutilized
- AI executors like `ClaudeCodeExecutor` in `internal/ai-executors/claude-code.go` create basic logs with raw stdout

## Implementation Plan

### 1. Database Schema Enhancement

**File: Create new migration `000019_enhance_execution_logs_structure.up.sql`**

- Add structured fields to `execution_logs` table:
  - `log_type VARCHAR(20)` (user, assistant, tool_result, result, etc.)
  - `tool_name VARCHAR(100)` (for tool_use logs)
  - `tool_use_id VARCHAR(100)` (for tool result correlation)
  - `parsed_content JSONB` (structured content)
  - `is_error BOOLEAN` (for result logs)
  - `duration_ms INTEGER` (for result logs)
  - `num_turns INTEGER` (for result logs)
- Add indexes for new fields
- Update `internal/entity/execution_log.go` with new fields

### 2. Backend Log Parsing Service

**File: Create `internal/service/log_parser.go`**

- `ParseLogMessage(rawMessage string) (*ParsedLogData, error)` - Main parsing function
- `ParseUserMessage(data map[string]interface{}) *ParsedLogData`
- `ParseAssistantMessage(data map[string]interface{}) *ParsedLogData`
- `ParseToolResultMessage(data map[string]interface{}) *ParsedLogData`
- `ParseResultMessage(data map[string]interface{}) *ParsedLogData`
- Handle different log formats from AI CLI tools
- Extract structured data into appropriate fields

### 3. Update AI Executors Parsing

**File: Modify `internal/ai-executors/claude-code.go`**

- Update `ParseOutputToLogs()` method to use new log parser service
- Parse JSON from each stdout line
- Populate structured fields instead of just raw message
- Handle parsing errors gracefully with fallback to raw message

**Files: Update other executors**

- `internal/ai-executors/cursor-agent.go`
- `internal/ai-executors/fake-code.go`
- Ensure consistent log parsing across all AI executors

### 4. Repository Layer Updates

**File: `internal/repository/execution_log.go`**

- Update queries to include new structured fields
- Add methods for filtering by log_type, tool_name, etc.
- Maintain backward compatibility with existing queries

### 5. Handler/API Updates

**File: `internal/handler/execution_log.go`**

- Update DTOs to include structured fields
- Ensure API responses include parsed content
- Add filtering capabilities by log type

### 6. Frontend Updates

**File: `frontend/src/types/execution.ts`**

- Update `ExecutionLog` interface to include new structured fields:
  - `log_type: string`
  - `tool_name?: string`
  - `tool_use_id?: string`
  - `parsed_content?: any`
  - `is_error?: boolean`
  - `duration_ms?: number`
  - `num_turns?: number`

**Files: Simplify frontend components**

- `frontend/src/components/executions/execution-logs-modal.tsx`
- `frontend/src/components/executions/execution-logs-pannel.tsx`
- Remove `JSON.parse(message)` logic
- Use structured fields directly from API
- Simplify rendering logic by using pre-parsed data
- Keep fallback to raw message for legacy logs

### 7. Migration Strategy

- Use database migration to add new columns with default values
- Update parsing logic to populate both old and new fields during transition
- Frontend can gracefully handle both old (unparsed) and new (structured) log formats
- Gradually migrate existing logs through background job if needed

### 8. Testing

- Unit tests for log parser service with various log formats
- Integration tests for AI executor parsing
- API tests for structured log retrieval
- Frontend tests for rendering structured logs

## Benefits

- **Performance**: Frontend no longer needs to parse JSON for every log display
- **Reliability**: Server-side parsing with proper error handling
- **Queryability**: Structured data enables efficient filtering and searching
- **Maintainability**: Centralized parsing logic instead of duplicated frontend code
- **Extensibility**: Easy to add new log types and structured fields

## Implementation Order

1. Database migration and entity updates
2. Log parser service implementation
3. Update AI executors to use parser
4. Repository and handler updates
5. Frontend type and component updates
6. Testing and validation
