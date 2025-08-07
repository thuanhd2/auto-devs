import { useState, useMemo, useEffect } from 'react'
import { taskOptimisticUpdates } from '@/services/optimisticUpdates'
import type { Task, TaskStatus } from '@/types/task'
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
import { KANBAN_COLUMNS, canTransitionTo } from '@/lib/kanban'
import {
  useWebSocketProject,
  useWebSocketContext,
} from '@/context/websocket-context'
import { useKeyboardNavigation } from '@/hooks/use-keyboard-navigation'
import { useUpdateTask, useOptimisticTaskUpdate } from '@/hooks/use-tasks'
import { ScrollArea } from '@/components/ui/scroll-area'
import { KanbanColumn } from './kanban-column'

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
  const [localTasks, setLocalTasks] = useState<Task[]>(tasks)
  const updateTaskMutation = useUpdateTask()
  const optimisticUpdate = useOptimisticTaskUpdate()

  // WebSocket integration
  const { setCurrentProjectId } = useWebSocketProject(projectId)
  const { isConnected } = useWebSocketContext()

  // Keep local tasks in sync with props
  useEffect(() => {
    setLocalTasks(tasks)
  }, [tasks])

  // Set current project for WebSocket subscriptions
  useEffect(() => {
    if (projectId) {
      setCurrentProjectId(projectId)
    }
  }, [projectId, setCurrentProjectId])

  // Filter and group tasks by status
  const filteredTasks = useMemo(() => {
    return localTasks.filter((task) => {
      if (!searchQuery) return true

      const query = searchQuery.toLowerCase()
      return (
        task.title.toLowerCase().includes(query) ||
        task.description.toLowerCase().includes(query)
      )
    })
  }, [localTasks, searchQuery])

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

    filteredTasks.forEach((task) => {
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
    const activeTask = tasks.find((task) => task.id === activeTaskId)
    if (!activeTask) return

    // Check if we're moving to a different column
    if (activeTask.status !== overColumnId) {
      // Check if the transition is valid
      if (!canTransitionTo(activeTask.status, overColumnId)) {
        return // Invalid transition, don't allow drop
      }

      // Apply optimistic update using WebSocket service
      if (activeTask) {
        taskOptimisticUpdates.updateTaskStatus(
          activeTaskId,
          overColumnId,
          activeTask,
          (updatedTask) => {
            setLocalTasks((prev) =>
              prev.map((t) => (t.id === activeTaskId ? updatedTask : t))
            )
          }
        )
      }
    }
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    setActiveTaskId(null)

    if (!over) return

    const activeTaskId = active.id as string
    const overColumnId = over.id as TaskStatus

    // Find the active task
    const activeTask = tasks.find((task) => task.id === activeTaskId)
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
      <div className='h-full w-full overflow-hidden'>
        <ScrollArea className='h-full w-full' orientation='horizontal'>
          <div className='flex gap-6 p-6' style={{ minWidth: 'max-content' }}>
            {KANBAN_COLUMNS.map((column) => (
              <div
                key={column.id}
                className='w-80 flex-shrink-0 rounded-lg border shadow-sm'
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
