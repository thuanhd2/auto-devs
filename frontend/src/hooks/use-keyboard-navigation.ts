import { useEffect, useCallback, useState } from 'react'
import type { Task, TaskStatus } from '@/types/task'
import { KANBAN_COLUMNS } from '@/lib/kanban'

interface UseKeyboardNavigationProps {
  tasks: Task[]
  onEditTask?: (task: Task) => void
  onCreateTask?: () => void
  onDeleteTask?: (taskId: string) => void
}

export function useKeyboardNavigation({
  tasks,
  onEditTask,
  onCreateTask,
  onDeleteTask,
}: UseKeyboardNavigationProps) {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)
  const [selectedColumnId, setSelectedColumnId] = useState<TaskStatus>('TODO')

  const tasksByStatus = tasks.reduce(
    (acc, task) => {
      if (!acc[task.status]) acc[task.status] = []
      acc[task.status].push(task)
      return acc
    },
    {} as Record<TaskStatus, Task[]>
  )

  const selectedTask = selectedTaskId
    ? tasks.find((t) => t.id === selectedTaskId)
    : null

  const moveToNextColumn = useCallback(() => {
    const currentIndex = KANBAN_COLUMNS.findIndex(
      (col) => col.id === selectedColumnId
    )
    const nextIndex = Math.min(currentIndex + 1, KANBAN_COLUMNS.length - 1)
    setSelectedColumnId(KANBAN_COLUMNS[nextIndex].id)
    setSelectedTaskId(null)
  }, [selectedColumnId])

  const moveToPrevColumn = useCallback(() => {
    const currentIndex = KANBAN_COLUMNS.findIndex(
      (col) => col.id === selectedColumnId
    )
    const prevIndex = Math.max(currentIndex - 1, 0)
    setSelectedColumnId(KANBAN_COLUMNS[prevIndex].id)
    setSelectedTaskId(null)
  }, [selectedColumnId])

  const moveToNextTask = useCallback(() => {
    const columnTasks = tasksByStatus[selectedColumnId] || []
    if (columnTasks.length === 0) return

    if (!selectedTaskId) {
      setSelectedTaskId(columnTasks[0].id)
      return
    }

    const currentIndex = columnTasks.findIndex((t) => t.id === selectedTaskId)
    const nextIndex = Math.min(currentIndex + 1, columnTasks.length - 1)
    setSelectedTaskId(columnTasks[nextIndex].id)
  }, [selectedColumnId, selectedTaskId, tasksByStatus])

  const moveToPrevTask = useCallback(() => {
    const columnTasks = tasksByStatus[selectedColumnId] || []
    if (columnTasks.length === 0) return

    if (!selectedTaskId) {
      setSelectedTaskId(columnTasks[columnTasks.length - 1].id)
      return
    }

    const currentIndex = columnTasks.findIndex((t) => t.id === selectedTaskId)
    const prevIndex = Math.max(currentIndex - 1, 0)
    setSelectedTaskId(columnTasks[prevIndex].id)
  }, [selectedColumnId, selectedTaskId, tasksByStatus])

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Don't handle keyboard events when user is typing in an input
      if (
        event.target instanceof HTMLInputElement ||
        event.target instanceof HTMLTextAreaElement
      ) {
        return
      }

      switch (event.key) {
        case 'ArrowLeft':
          event.preventDefault()
          moveToPrevColumn()
          break
        case 'ArrowRight':
          event.preventDefault()
          moveToNextColumn()
          break
        case 'ArrowUp':
          event.preventDefault()
          moveToPrevTask()
          break
        case 'ArrowDown':
          event.preventDefault()
          moveToNextTask()
          break
        case 'Enter':
          event.preventDefault()
          if (selectedTask) {
            onEditTask?.(selectedTask)
          } else {
            onCreateTask?.()
          }
          break
        case 'Delete':
        case 'Backspace':
          event.preventDefault()
          if (selectedTaskId) {
            onDeleteTask?.(selectedTaskId)
          }
          break
        case 'Escape':
          event.preventDefault()
          setSelectedTaskId(null)
          break
        case 'n':
        case 'N':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault()
            onCreateTask?.()
          }
          break
      }
    },
    [
      selectedTask,
      selectedTaskId,
      moveToNextColumn,
      moveToPrevColumn,
      moveToNextTask,
      moveToPrevTask,
      onEditTask,
      onCreateTask,
      onDeleteTask,
    ]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  return {
    selectedTaskId,
    selectedColumnId,
    setSelectedTaskId,
    setSelectedColumnId,
  }
}
