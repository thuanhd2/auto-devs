import { Plus, BrushCleaning } from 'lucide-react'
import type { KanbanColumn } from '@/lib/kanban'
import { Button } from '@/components/ui/button'

interface EmptyColumnProps {
  column: KanbanColumn
  onCreateTask?: () => void
  canCreateTask?: boolean
}

export function EmptyColumn({
  column,
  onCreateTask,
  canCreateTask = true,
}: EmptyColumnProps) {
  return (
    <div className='flex flex-col items-center justify-center py-8 text-center'>
      <div className='mb-3 flex h-12 w-12 items-center justify-center rounded-full'>
        <BrushCleaning className='h-6 w-6' />
      </div>

      <h3 className='mb-1 font-medium'>No tasks</h3>
      <p className='mb-4 max-w-40 text-sm'>{column.description}</p>

      {canCreateTask && column.id === 'TODO' && (
        <Button
          variant='outline'
          size='sm'
          onClick={onCreateTask}
          className='text-xs'
        >
          <Plus className='mr-1 h-3 w-3' />
          Add Task
        </Button>
      )}
    </div>
  )
}
