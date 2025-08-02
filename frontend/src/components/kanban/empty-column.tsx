import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { KanbanColumn } from '@/lib/kanban'

interface EmptyColumnProps {
  column: KanbanColumn
  onCreateTask?: () => void
  canCreateTask?: boolean
}

export function EmptyColumn({ column, onCreateTask, canCreateTask = true }: EmptyColumnProps) {
  return (
    <div className="flex flex-col items-center justify-center py-8 text-center">
      <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
        <div className="w-6 h-6 rounded bg-gray-300" />
      </div>
      
      <h3 className="font-medium text-gray-900 mb-1">No tasks</h3>
      <p className="text-sm text-gray-500 mb-4 max-w-40">
        {column.description}
      </p>
      
      {canCreateTask && column.id === 'TODO' && (
        <Button
          variant="outline"
          size="sm"
          onClick={onCreateTask}
          className="text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add Task
        </Button>
      )}
    </div>
  )
}