import { useDroppable } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { Plus, MoreHorizontal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { TaskCard } from './task-card'
import { EmptyColumn } from './empty-column'
import { DraggableTaskCard } from './draggable-task-card'
import type { Task } from '@/types/task'
import type { KanbanColumn } from '@/lib/kanban'

interface KanbanColumnProps {
  column: KanbanColumn
  tasks: Task[]
  onCreateTask?: () => void
  onEditTask?: (task: Task) => void
  onDeleteTask?: (taskId: string) => void
  onViewTaskDetails?: (task: Task) => void
  isCompactView?: boolean
  selectedTaskId?: string | null
  isSelectedColumn?: boolean
}

export function KanbanColumn({
  column,
  tasks,
  onCreateTask,
  onEditTask,
  onDeleteTask,
  onViewTaskDetails,
  isCompactView = false,
  selectedTaskId,
  isSelectedColumn = false,
}: KanbanColumnProps) {
  const { isOver, setNodeRef } = useDroppable({
    id: column.id,
  })

  const taskIds = tasks.map(task => task.id)

  return (
    <div 
      className={`flex flex-col h-full ${isSelectedColumn ? 'ring-2 ring-blue-500 ring-opacity-50' : ''}`}
      data-column={column.id}
    >
      {/* Column Header */}
      <div className="flex items-center justify-between p-4 border-b bg-white">
        <div className="flex items-center gap-2">
          <h2 className="font-semibold text-gray-900">{column.title}</h2>
          <Badge 
            variant="secondary" 
            className="text-xs task-count transition-transform duration-200"
            data-testid={`${column.id}-count`}
          >
            {tasks.length}
          </Badge>
        </div>
        
        <div className="flex items-center gap-1">
          {column.id === 'TODO' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onCreateTask}
              className="h-7 w-7 p-0 hover:scale-105 transition-transform duration-200"
            >
              <Plus className="h-3 w-3" />
            </Button>
          )}
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 hover:scale-105 transition-transform duration-200"
              >
                <MoreHorizontal className="h-3 w-3" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
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
        className={`
          flex-1 p-4 space-y-3 overflow-y-auto min-h-96
          ${isOver ? 'bg-blue-50 border-blue-200' : ''} 
          transition-all duration-200 ease-out
        `}
        data-column-content={column.id}
      >
        {tasks.length === 0 ? (
          <EmptyColumn
            column={column}
            onCreateTask={onCreateTask}
            canCreateTask={column.id === 'TODO'}
          />
        ) : (
          <SortableContext items={taskIds} strategy={verticalListSortingStrategy}>
            <div className="space-y-3">
              {tasks.map((task) => (
                <DraggableTaskCard
                  key={task.id}
                  task={task}
                  onEdit={onEditTask}
                  onDelete={onDeleteTask}
                  onViewDetails={onViewTaskDetails}
                  isCompact={isCompactView}
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