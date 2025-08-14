import type { TaskFilters, TaskStatus } from '@/types/task'
import { Search, X } from 'lucide-react'
import { KANBAN_COLUMNS } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface BoardFiltersProps {
  filters: TaskFilters
  onFiltersChange: (filters: TaskFilters) => void
  searchQuery: string
  onSearchChange: (query: string) => void
  taskCount: number
}

export function BoardFilters({
  filters,
  onFiltersChange,
  searchQuery,
  onSearchChange,
  taskCount,
}: BoardFiltersProps) {
  const activeFilterCount = [
    filters.status?.length || 0,
    filters.git_status?.length || 0,
    filters.sortBy ? 1 : 0,
    filters.branch_search ? 1 : 0,
  ].reduce((a, b) => a + b, 0)

  const handleStatusToggle = (status: TaskStatus, checked: boolean) => {
    const currentStatuses = filters.status || []
    const newStatuses = checked
      ? [...currentStatuses, status]
      : currentStatuses.filter((s) => s !== status)

    onFiltersChange({
      ...filters,
      status: newStatuses.length > 0 ? newStatuses : undefined,
    })
  }

  const hasActiveFilters =
    activeFilterCount > 0 ||
    searchQuery.length > 0 ||
    (filters.branch_search && filters.branch_search.length > 0)

  return (
    <div className='flex items-center gap-3 p-4 py-0'>
      {/* Search */}
      <div className='relative max-w-md flex-1'>
        <Search className='absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform' />
        <Input
          placeholder='Search tasks...'
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className='pl-10'
        />
        {searchQuery && (
          <Button
            variant='ghost'
            size='sm'
            onClick={() => onSearchChange('')}
            className='absolute top-1/2 right-1 h-6 w-6 -translate-y-1/2 transform p-0'
          >
            <X className='h-3 w-3' />
          </Button>
        )}
      </div>

      {/* Task Count */}
      <div className='text-sm text-gray-500'>
        {taskCount} task{taskCount !== 1 ? 's' : ''}
      </div>

      {/* Active Filter Tags */}
      {hasActiveFilters && (
        <div className='flex items-center gap-2'>
          {filters.status?.map((status) => {
            const column = KANBAN_COLUMNS.find((col) => col.id === status)
            return (
              <Badge key={status} variant='secondary' className='text-xs'>
                {column?.title}
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => handleStatusToggle(status, false)}
                  className='ml-1 h-3 w-3 p-0 hover:bg-transparent'
                >
                  <X className='h-2 w-2' />
                </Button>
              </Badge>
            )
          })}

          {searchQuery && (
            <Badge variant='secondary' className='text-xs'>
              "{searchQuery}"
              <Button
                variant='ghost'
                size='sm'
                onClick={() => onSearchChange('')}
                className='ml-1 h-3 w-3 p-0 hover:bg-transparent'
              >
                <X className='h-2 w-2' />
              </Button>
            </Badge>
          )}
        </div>
      )}
    </div>
  )
}
