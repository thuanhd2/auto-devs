import type { Task } from '@/types/task'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { TaskCard } from './task-card'

interface DraggableTaskCardProps {
  task: Task
  index: number
  onEdit: (task: Task) => void
  onDelete: (taskId: string) => void
  onStatusChange: (taskId: string, status: TaskStatus) => void
  onMove: (taskId: string, newStatus: TaskStatus, newIndex: number) => void
}

export function DraggableTaskCard({
  task,
  index,
  onEdit,
  onDelete,
  onStatusChange,
  onMove,
}: DraggableTaskCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={` ${isDragging ? 'z-50' : ''} ${isSelected ? 'ring-2 ring-blue-500' : ''} touch-none transition-all duration-200 ease-out`}
      data-task-id={task.id}
      data-task-status={task.status}
    >
      <TaskCard
        task={task}
        isDragging={isDragging}
        onEdit={onEdit}
        onDelete={onDelete}
        onViewDetails={onViewDetails}
      />
    </div>
  )
}
