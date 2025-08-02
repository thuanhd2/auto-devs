import { useState, useMemo } from 'react'
import { useKeyboardNavigation } from '@/hooks/use-keyboard-navigation'
import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  DragStartEvent,
  closestCorners,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { arrayMove } from '@dnd-kit/sortable'
import { ScrollArea } from '@/components/ui/scroll-area'
import { KanbanColumn } from './kanban-column'
import { KANBAN_COLUMNS, canTransitionTo } from '@/lib/kanban'
import { useUpdateTask, useOptimisticTaskUpdate } from '@/hooks/use-tasks'
import type { Task, TaskStatus } from '@/types/task'

interface KanbanBoardProps {
  tasks: Task[]
  projectId: string
  onCreateTask?: () => void
  onEditTask?: (task: Task) => void
  onDeleteTask?: (taskId: string) => void
  onViewTaskDetails?: (task: Task) => void
  isCompactView?: boolean
  searchQuery?: string
}

export function KanbanBoard({
  tasks,
  projectId,
  onCreateTask,
  onEditTask,
  onDeleteTask,
  onViewTaskDetails,
  isCompactView = false,
  searchQuery = '',
}: KanbanBoardProps) {
  const [activeTaskId, setActiveTaskId] = useState<string | null>(null)
  const updateTaskMutation = useUpdateTask()
  const optimisticUpdate = useOptimisticTaskUpdate()
  
  const { selectedTaskId, selectedColumnId } = useKeyboardNavigation({
    tasks: filteredTasks,
    onEditTask,
    onCreateTask,
    onDeleteTask,
  })

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  )

  // Filter and group tasks by status
  const filteredTasks = useMemo(() => {
    return tasks.filter(task => {
      if (!searchQuery) return true
      
      const query = searchQuery.toLowerCase()
      return (
        task.title.toLowerCase().includes(query) ||
        task.description.toLowerCase().includes(query)
      )
    })
  }, [tasks, searchQuery])

  const tasksByStatus = useMemo(() => {
    const grouped: Record<TaskStatus, Task[]> = {
      TODO: [],
      PLANNING: [],
      PLAN_REVIEWING: [],
      IMPLEMENTING: [],
      CODE_REVIEWING: [],
      DONE: [],
      CANCELLED: [],
    }

    filteredTasks.forEach(task => {
      grouped[task.status].push(task)
    })

    return grouped
  }, [filteredTasks])

  const handleDragStart = (event: DragStartEvent) => {
    setActiveTaskId(event.active.id as string)
  }

  const handleDragOver = (event: DragOverEvent) => {
    const { active, over } = event
    
    if (!over) return

    const activeTaskId = active.id as string
    const overColumnId = over.id as TaskStatus

    // Find the active task
    const activeTask = tasks.find(task => task.id === activeTaskId)
    if (!activeTask) return

    // Check if we're moving to a different column
    if (activeTask.status !== overColumnId) {
      // Check if the transition is valid
      if (!canTransitionTo(activeTask.status, overColumnId)) {
        return // Invalid transition, don't allow drop
      }

      // Apply optimistic update
      optimisticUpdate(projectId, activeTaskId, overColumnId)
    }
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    setActiveTaskId(null)

    if (!over) return

    const activeTaskId = active.id as string
    const overColumnId = over.id as TaskStatus

    // Find the active task
    const activeTask = tasks.find(task => task.id === activeTaskId)
    if (!activeTask) return

    // If status changed, update via API
    if (activeTask.status !== overColumnId) {
      if (canTransitionTo(activeTask.status, overColumnId)) {
        updateTaskMutation.mutate({
          taskId: activeTaskId,
          updates: { status: overColumnId },
        })
      }
    }
  }

  const handleDragCancel = () => {
    setActiveTaskId(null)
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <div className="h-full">
        <ScrollArea className="h-full">
          <div className="flex gap-6 p-6 min-w-max">
            {KANBAN_COLUMNS.map((column) => (
              <div
                key={column.id}
                className="w-80 bg-gray-50 rounded-lg border shadow-sm flex-shrink-0"
              >
                <KanbanColumn
                  column={column}
                  tasks={tasksByStatus[column.id]}
                  onCreateTask={column.id === 'TODO' ? onCreateTask : undefined}
                  onEditTask={onEditTask}
                  onDeleteTask={onDeleteTask}
                  onViewTaskDetails={onViewTaskDetails}
                  isCompactView={isCompactView}
                  selectedTaskId={selectedTaskId}
                  isSelectedColumn={selectedColumnId === column.id}
                />
              </div>
            ))}
          </div>
        </ScrollArea>
      </div>
    </DndContext>
  )
}