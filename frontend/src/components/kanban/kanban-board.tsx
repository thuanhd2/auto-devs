import { useState, useEffect, useMemo } from 'react'
import type { Task, TaskStatus } from '@/types/task'
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd'
import { KANBAN_COLUMNS } from '@/lib/kanban'
import { useWebSocketConnection } from '@/context/websocket-context'
import { useTasks } from '@/hooks/use-tasks'
import { KanbanColumn } from './kanban-column'
import { TaskCard } from './task-card'

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
  const updateTaskMutation = useTasks()

  // WebSocket integration
  const { setCurrentProjectId } = useWebSocketConnection()
  const { isConnected } = useWebSocketConnection()

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

  const handleDragStart = useCallback((event: DragStartEvent) => {
    setActiveTaskId(event.active.id as string)
  }, [])

  const handleDragOver = useCallback(
    (event: DragOverEvent) => {
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
    },
    [tasks]
  )

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
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
    },
    [tasks, updateTaskMutation]
  )

  const handleDragCancel = useCallback(() => {
    setActiveTaskId(null)
  }, [])

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
