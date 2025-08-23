import { StructuredLogItem } from './structured-log-item'
import { ExecutionLog } from '@/types/execution'

// Demo component to test structured logs
export function StructuredLogDemo() {
  // Sample structured logs for testing
  const demoLogs: ExecutionLog[] = [
    {
      id: '1',
      execution_id: 'exec-1',
      level: 'info',
      message: 'Legacy log message',
      timestamp: new Date().toISOString(),
      source: 'demo',
      created_at: new Date().toISOString(),
      line: 1,
    },
    {
      id: '2',
      execution_id: 'exec-1',
      level: 'info',
      message: 'User message',
      timestamp: new Date().toISOString(),
      source: 'demo',
      created_at: new Date().toISOString(),
      line: 2,
      // New structured fields
      log_type: 'user',
      parsed_content: {
        text: 'Please help me implement a feature'
      }
    },
    {
      id: '3',
      execution_id: 'exec-1',
      level: 'info',
      message: 'Assistant message',
      timestamp: new Date().toISOString(),
      source: 'demo',
      created_at: new Date().toISOString(),
      line: 3,
      log_type: 'assistant',
      parsed_content: {
        content: [
          {
            type: 'text',
            text: 'I\'ll help you implement that feature. Let me start by analyzing the requirements.'
          },
          {
            type: 'tool_use',
            name: 'FileReader',
            input: { file_path: '/src/main.ts' }
          }
        ]
      }
    },
    {
      id: '4',
      execution_id: 'exec-1',
      level: 'info',
      message: 'Tool result',
      timestamp: new Date().toISOString(),
      source: 'demo',
      created_at: new Date().toISOString(),
      line: 4,
      log_type: 'tool_result',
      tool_name: 'FileReader',
      tool_use_id: 'tool-123',
      parsed_content: {
        content: 'File content: export function main() { console.log("Hello World"); }'
      }
    },
    {
      id: '5',
      execution_id: 'exec-1',
      level: 'info',
      message: 'Execution result',
      timestamp: new Date().toISOString(),
      source: 'demo',
      created_at: new Date().toISOString(),
      line: 5,
      log_type: 'result',
      duration_ms: 1500,
      num_turns: 3,
      is_error: false,
      parsed_content: {
        result: 'Feature implemented successfully'
      }
    },
    {
      id: '6',
      execution_id: 'exec-1',
      level: 'error',
      message: 'Error result',
      timestamp: new Date().toISOString(),
      source: 'demo',
      created_at: new Date().toISOString(),
      line: 6,
      log_type: 'result',
      duration_ms: 500,
      num_turns: 1,
      is_error: true,
      parsed_content: {
        error: 'Claude AI usage limit reached|1755338400'
      }
    }
  ]

  return (
    <div className="p-4 space-y-4">
      <h2 className="text-xl font-bold">Structured Log Demo</h2>
      <p className="text-sm text-gray-600">
        This demo shows how structured logs are rendered with the new fields.
      </p>
      
      <div className="space-y-2">
        {demoLogs.map((log) => (
          <StructuredLogItem key={log.id} log={log} />
        ))}
      </div>
    </div>
  )
}
