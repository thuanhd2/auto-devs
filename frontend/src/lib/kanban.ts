import type { TaskStatus } from '@/types/task'

export interface KanbanColumn {
  id: TaskStatus
  title: string
  color: string
  description: string
  maxItems?: number
}

export const KANBAN_COLUMNS: KanbanColumn[] = [
  {
    id: 'TODO',
    title: 'To Do',
    color: 'bg-slate-100 border-slate-200',
    description: 'Tasks ready to be started',
  },
  {
    id: 'PLANNING',
    title: 'Planning',
    color: 'bg-blue-100 border-blue-200',
    description: 'Tasks being planned by AI',
  },
  {
    id: 'PLAN_REVIEWING',
    title: 'Plan Review',
    color: 'bg-amber-100 border-amber-200',
    description: 'Plans awaiting review',
  },
  {
    id: 'IMPLEMENTING',
    title: 'Implementing',
    color: 'bg-orange-100 border-orange-200',
    description: 'Tasks being implemented',
  },
  {
    id: 'CODE_REVIEWING',
    title: 'Code Review',
    color: 'bg-purple-100 border-purple-200',
    description: 'Code awaiting review',
  },
  {
    id: 'DONE',
    title: 'Done',
    color: 'bg-green-100 border-green-200',
    description: 'Completed tasks',
  },
  {
    id: 'CANCELLED',
    title: 'Cancelled',
    color: 'bg-red-100 border-red-200',
    description: 'Cancelled tasks',
  },
]

export const TASK_STATUS_TRANSITIONS: Record<TaskStatus, TaskStatus[]> = {
  TODO: ['PLANNING', 'CANCELLED'],
  PLANNING: ['PLAN_REVIEWING', 'CANCELLED'],
  PLAN_REVIEWING: ['IMPLEMENTING', 'PLANNING', 'CANCELLED'],
  IMPLEMENTING: ['CODE_REVIEWING', 'CANCELLED'],
  CODE_REVIEWING: ['DONE', 'CANCELLED'],
  DONE: [],
  CANCELLED: [],
}

export function canTransitionTo(fromStatus: TaskStatus, toStatus: TaskStatus): boolean {
  return TASK_STATUS_TRANSITIONS[fromStatus].includes(toStatus)
}

export function getColumnById(columnId: TaskStatus): KanbanColumn | undefined {
  return KANBAN_COLUMNS.find(col => col.id === columnId)
}

export function getStatusColor(status: TaskStatus): string {
  const column = getColumnById(status)
  return column?.color || 'bg-gray-100 border-gray-200'
}

export function getStatusTitle(status: TaskStatus): string {
  const column = getColumnById(status)
  return column?.title || status
}