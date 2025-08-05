import { useState } from 'react'
import { Filter, Search, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Separator } from '@/components/ui/separator'
import { KANBAN_COLUMNS } from '@/lib/kanban'
import { GitStatusBadge, getGitStatusLabel } from './git-status-badge'
import type { TaskFilters, TaskStatus, TaskGitStatus } from '@/types/task'

const GIT_STATUS_OPTIONS: { id: TaskGitStatus; title: string }[] = [
  { id: 'NO_GIT', title: 'No Git' },
  { id: 'WORKTREE_PENDING', title: 'Creating...' },
  { id: 'WORKTREE_CREATED', title: 'Worktree Ready' },
  { id: 'BRANCH_CREATED', title: 'Branch Ready' },
  { id: 'CHANGES_PENDING', title: 'Changes' },
  { id: 'CHANGES_STAGED', title: 'Staged' },
  { id: 'CHANGES_COMMITTED', title: 'Committed' },
  { id: 'PR_CREATED', title: 'PR Created' },
  { id: 'PR_MERGED', title: 'Merged' },
  { id: 'WORKTREE_ERROR', title: 'Git Error' },
]

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
  const [isFilterOpen, setIsFilterOpen] = useState(false)

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
      : currentStatuses.filter(s => s !== status)
    
    onFiltersChange({
      ...filters,
      status: newStatuses.length > 0 ? newStatuses : undefined,
    })
  }
  const handleGitStatusToggle = (gitStatus: TaskGitStatus, checked: boolean) => {
    const currentGitStatuses = filters.git_status || []
    const newGitStatuses = checked
      ? [...currentGitStatuses, gitStatus]
      : currentGitStatuses.filter(s => s !== gitStatus)
    
    onFiltersChange({
      ...filters,
      git_status: newGitStatuses.length > 0 ? newGitStatuses : undefined,
    })
  }

  const handleBranchSearchChange = (branchSearch: string) => {
    onFiltersChange({
      ...filters,
      branch_search: branchSearch.trim() || undefined,
    })
  }

  const handleSortChange = (sortBy: TaskFilters['sortBy']) => {
    onFiltersChange({
      ...filters,
      sortBy,
      sortOrder: sortBy ? 'desc' : undefined,
    })
  }

  const handleSortOrderToggle = () => {
    onFiltersChange({
      ...filters,
      sortOrder: filters.sortOrder === 'asc' ? 'desc' : 'asc',
    })
  }

  const clearAllFilters = () => {
    onFiltersChange({})
    onSearchChange('')
  }

  const hasActiveFilters = activeFilterCount > 0 || searchQuery.length > 0 || (filters.branch_search && filters.branch_search.length > 0)

  return (
    <div className="flex items-center gap-3 p-4 bg-white border-b">
      {/* Search */}
      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          placeholder="Search tasks..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10"
        />
        {searchQuery && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onSearchChange('')}
            className="absolute right-1 top-1/2 transform -translate-y-1/2 h-6 w-6 p-0"
          >
            <X className="h-3 w-3" />
          </Button>
        )}
      </div>

      {/* Filters */}
      <Popover open={isFilterOpen} onOpenChange={setIsFilterOpen}>
        <PopoverTrigger asChild>
          <Button variant="outline" className="relative">
            <Filter className="h-4 w-4 mr-2" />
            Filters
            {activeFilterCount > 0 && (
              <Badge
                variant="secondary"
                className="ml-2 h-5 w-5 rounded-full p-0 flex items-center justify-center text-xs"
              >
                {activeFilterCount}
              </Badge>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80" align="end">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium">Filters</h4>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearAllFilters}
                disabled={!hasActiveFilters}
              >
                Clear all
              </Button>
            </div>

            <Separator />

            {/* Status Filter */}
            <div>
              <h5 className="text-sm font-medium mb-2">Status</h5>
              <div className="space-y-2">
                {KANBAN_COLUMNS.map((column) => (
                  <div key={column.id} className="flex items-center space-x-2">
                    <Checkbox
                      id={column.id}
                      checked={filters.status?.includes(column.id) || false}
                      onCheckedChange={(checked) =>
                        handleStatusToggle(column.id, !!checked)
                      }
                    />
                    <label
                      htmlFor={column.id}
                      className="text-sm font-normal cursor-pointer flex-1"
                    >
                      {column.title}
                    </label>
                  </div>
                ))}
              </div>
            </div>

            <Separator />

            {/* Sort */}
            <div>
              <h5 className="text-sm font-medium mb-2">Sort by</h5>
              <div className="flex gap-2">
                <Select
                  value={filters.sortBy || ''}
                  onValueChange={handleSortChange}
                >
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder="Select..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="created_at">Created Date</SelectItem>
                    <SelectItem value="updated_at">Updated Date</SelectItem>
                    <SelectItem value="title">Title</SelectItem>
                  </SelectContent>
                </Select>
                
                {filters.sortBy && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleSortOrderToggle}
                    className="px-3"
                  >
                    {filters.sortOrder === 'asc' ? '↑' : '↓'}
                  </Button>
                )}
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>

      {/* Task Count */}
      <div className="text-sm text-gray-500">
        {taskCount} task{taskCount !== 1 ? 's' : ''}
      </div>

      {/* Active Filter Tags */}
      {hasActiveFilters && (
        <div className="flex items-center gap-2">
          {filters.status?.map((status) => {
            const column = KANBAN_COLUMNS.find(col => col.id === status)
            return (
              <Badge key={status} variant="secondary" className="text-xs">
                {column?.title}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleStatusToggle(status, false)}
                  className="ml-1 h-3 w-3 p-0 hover:bg-transparent"
                >
                  <X className="h-2 w-2" />
                </Button>
              </Badge>
            )
          })}
          
          {searchQuery && (
            <Badge variant="secondary" className="text-xs">
              "{searchQuery}"
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onSearchChange('')}
                className="ml-1 h-3 w-3 p-0 hover:bg-transparent"
              >
                <X className="h-2 w-2" />
              </Button>
            </Badge>
          )}
        </div>
      )}
    </div>
  )
}