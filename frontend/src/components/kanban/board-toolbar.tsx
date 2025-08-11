import type { TaskFilters } from '@/types/task'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { BoardFilters } from './board-filters'

interface BoardToolbarProps {
  filters: TaskFilters
  onFiltersChange: (filters: TaskFilters) => void
  searchQuery: string
  onSearchChange: (query: string) => void
  onCreateTask: () => void
  taskCount: number
}

export function BoardToolbar({
  filters,
  onFiltersChange,
  searchQuery,
  onSearchChange,
  onCreateTask,
  taskCount,
}: BoardToolbarProps) {
  return (
    <div className='bg-background/95 supports-[backdrop-filter]:bg-background/60 flex items-center justify-between border-b backdrop-blur'>
      <div className='flex items-center gap-2'>
        <Button onClick={onCreateTask} size='sm'>
          <Plus className='mr-2 h-4 w-4' />
          New Task
        </Button>
      </div>

      <BoardFilters
        filters={filters}
        onFiltersChange={onFiltersChange}
        searchQuery={searchQuery}
        onSearchChange={onSearchChange}
        taskCount={taskCount}
      />
    </div>
  )
}
