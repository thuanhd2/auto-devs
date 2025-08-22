import type { Task } from '@/types/task'
import { useDroppable } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { Plus, MoreHorizontal } from 'lucide-react'
import type { KanbanColumn } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { DraggableTaskCard } from './draggable-task-card'
import { EmptyColumn } from './empty-column'

interface KanbanColumnProps {
  column: KanbanColumn
  tasks: Task[]
  onCreateTask?: () => void
  onEditTask?: (task: Task) => void
  onDeleteTask?: (taskId: string) => void
  onViewTaskDetails?: (task: Task) => void
  selectedTaskId?: string | null
  isSelectedColumn?: boolean
  onLoadDoneTasks?: () => void
  showLoadDoneAction?: boolean
}

export function KanbanColumn({
  column,
  tasks,
  onCreateTask,
  onViewTaskDetails,
  selectedTaskId,
  isSelectedColumn = false,
  onLoadDoneTasks,
  showLoadDoneAction = false,
}: KanbanColumnProps) {
  const { isOver, setNodeRef } = useDroppable({
    id: column.id,
  })

  const taskIds = tasks.map((task) => task.id)

  return (
    <div
      className={`flex h-full flex-col ${isSelectedColumn ? 'ring-opacity-50 ring-2 ring-blue-500' : ''}`}
      data-column={column.id}
    >
      {/* Column Header */}
      <div className='flex items-center justify-between border-b p-4'>
        <div className='flex items-center gap-2'>
          <h2 className='font-semibold'>{column.title}</h2>
          <Badge
            variant='secondary'
            className='task-count text-xs transition-transform duration-200'
            data-testid={`${column.id}-count`}
          >
            {tasks.length}
          </Badge>
        </div>

        <div className='flex items-center gap-1'>
          {column.id === 'TODO' && (
            <Button
              variant='ghost'
              size='sm'
              onClick={onCreateTask}
              className='h-7 w-7 p-0 transition-transform duration-200 hover:scale-105'
            >
              <Plus className='h-3 w-3' />
            </Button>
          )}

          {column.id === 'DONE' && showLoadDoneAction && (
            <Button
              variant='outline'
              size='sm'
              onClick={onLoadDoneTasks}
              className='h-7 px-2 text-xs'
            >
              Load tasks
            </Button>
          )}

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant='ghost'
                size='sm'
                className='h-7 w-7 p-0 transition-transform duration-200 hover:scale-105'
              >
                <MoreHorizontal className='h-3 w-3' />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align='end'>
              <DropdownMenuItem>Sort by Date</DropdownMenuItem>
              <DropdownMenuItem>Sort by Title</DropdownMenuItem>
              <DropdownMenuItem>Filter Tasks</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Column Content */}
      <div
        ref={setNodeRef}
        className={`min-h-96 flex-1 space-y-3 overflow-y-auto p-4 ${isOver ? 'border-blue-200 bg-blue-50' : ''} transition-all duration-200 ease-out`}
        data-column-content={column.id}
      >
        {tasks.length === 0 ? (
          <EmptyColumn
            column={column}
            onCreateTask={onCreateTask}
            canCreateTask={column.id === 'TODO'}
          />
        ) : (
          <SortableContext
            items={taskIds}
            strategy={verticalListSortingStrategy}
          >
            <div className='space-y-3'>
              {tasks.map((task) => (
                <DraggableTaskCard
                  key={task.id}
                  task={task}
                  onViewDetails={onViewTaskDetails}
                  isSelected={selectedTaskId === task.id}
                />
              ))}
            </div>
          </SortableContext>
        )}
      </div>
    </div>
  )
}
