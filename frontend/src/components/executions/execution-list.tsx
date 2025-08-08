import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'

import { ExecutionItem } from './execution-item'

import { cn } from '@/lib/utils'
import type { Execution, ExecutionStatus, ExecutionFilters } from '@/types/execution'
import { 
  Search, 
  RefreshCw, 
  Plus,
  Clock,
  Play,
  CheckCircle,
  XCircle,
  AlertTriangle,
} from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'


interface ExecutionListProps {
  executions: Execution[]
  loading?: boolean
  error?: string
  onRefresh?: () => void
  onCreateExecution?: () => void
  onUpdateExecution?: (executionId: string, updates: Record<string, unknown>) => void
  onDeleteExecution?: (executionId: string) => void
  onViewLogs?: (executionId: string) => void
  onViewDetails?: (executionId: string) => void
  filters?: ExecutionFilters
  onFiltersChange?: (filters: ExecutionFilters) => void
  showCreateButton?: boolean
  showFilters?: boolean
  compact?: boolean
  expandable?: boolean
  emptyState?: React.ReactNode
  className?: string
}

const statusStats = (executions: Execution[]) => {
  const stats = executions.reduce((acc, execution) => {
    acc[execution.status] = (acc[execution.status] || 0) + 1
    return acc
  }, {} as Record<ExecutionStatus, number>)

  return {
    running: stats.running || 0,
    pending: stats.pending || 0,
    completed: stats.completed || 0,
    failed: stats.failed || 0,
    cancelled: stats.cancelled || 0,
    paused: stats.paused || 0,
  }
}

export function ExecutionList({
  executions,
  loading = false,
  error,
  onRefresh,
  onCreateExecution,
  onUpdateExecution,
  onDeleteExecution,
  onViewLogs,
  onViewDetails,
  filters,
  onFiltersChange,
  showCreateButton = true,
  showFilters = true,
  compact = false,
  expandable = false,
  emptyState,
  className,
}: ExecutionListProps) {
  const [searchTerm, setSearchTerm] = useState('')
  
  const stats = statusStats(executions)
  const hasActiveExecutions = stats.running > 0 || stats.pending > 0
  
  // Filter executions based on search term
  const filteredExecutions = executions.filter(execution => {
    if (!searchTerm) return true
    
    const searchLower = searchTerm.toLowerCase()
    return (
      execution.id.toLowerCase().includes(searchLower) ||
      execution.status.toLowerCase().includes(searchLower) ||
      execution.error?.toLowerCase().includes(searchLower)
    )
  })

  const handleStatusFilter = (status: ExecutionStatus | 'all') => {
    if (status === 'all') {
      onFiltersChange?.({ ...filters, status: undefined, statuses: undefined })
    } else {
      onFiltersChange?.({ ...filters, status, statuses: undefined })
    }
  }

  const handleSortChange = (sortBy: string) => {
    const [orderBy, orderDir] = sortBy.split('-')
    onFiltersChange?.({ 
      ...filters, 
      order_by: orderBy as 'started_at' | 'completed_at' | 'progress' | 'status', 
      order_dir: orderDir as 'asc' | 'desc' 
    })
  }

  return (
    <div className={cn('space-y-4', className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="text-lg font-semibold">
            Executions ({executions.length})
          </h3>
          {hasActiveExecutions && (
            <div className="flex items-center gap-1 text-sm text-blue-600">
              <Play className="h-3 w-3" />
              <span>{stats.running + stats.pending} active</span>
            </div>
          )}
        </div>
        
        <div className="flex items-center gap-2">
          {onRefresh && (
            <Button
              variant="outline"
              size="sm"
              onClick={onRefresh}
              disabled={loading}
              className="gap-1"
            >
              <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              Refresh
            </Button>
          )}
          
          {showCreateButton && onCreateExecution && (
            <Button
              size="sm"
              onClick={onCreateExecution}
              className="gap-1"
            >
              <Plus className="h-4 w-4" />
              New Execution
            </Button>
          )}
        </div>
      </div>

      {/* Status Stats */}
      <div className="flex flex-wrap gap-2">
        <StatusStatBadge
          label="Running"
          count={stats.running}
          status="running"
          icon={Play}
          onClick={() => handleStatusFilter('running')}
          active={filters?.status === 'running'}
        />
        <StatusStatBadge
          label="Pending"
          count={stats.pending}
          status="pending"
          icon={Clock}
          onClick={() => handleStatusFilter('pending')}
          active={filters?.status === 'pending'}
        />
        <StatusStatBadge
          label="Completed"
          count={stats.completed}
          status="completed"
          icon={CheckCircle}
          onClick={() => handleStatusFilter('completed')}
          active={filters?.status === 'completed'}
        />
        <StatusStatBadge
          label="Failed"
          count={stats.failed}
          status="failed"
          icon={XCircle}
          onClick={() => handleStatusFilter('failed')}
          active={filters?.status === 'failed'}
        />
        {filters?.status && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleStatusFilter('all')}
            className="text-muted-foreground"
          >
            Clear filter
          </Button>
        )}
      </div>

      {/* Filters */}
      {showFilters && (
        <div className="flex items-center gap-2">
          <div className="flex-1 max-w-xs">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search executions..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
          
          <Select onValueChange={handleSortChange} defaultValue="started_at-desc">
            <SelectTrigger className="w-40">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="started_at-desc">Latest first</SelectItem>
              <SelectItem value="started_at-asc">Oldest first</SelectItem>
              <SelectItem value="progress-desc">Progress ↓</SelectItem>
              <SelectItem value="progress-asc">Progress ↑</SelectItem>
              <SelectItem value="status-asc">Status A-Z</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}

      {/* Error State */}
      {error && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-red-800">
              <AlertTriangle className="h-4 w-4" />
              <span className="font-medium">Error loading executions</span>
            </div>
            <p className="text-sm text-red-700 mt-1">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Execution List */}
      {filteredExecutions.length === 0 ? (
        <div className="text-center py-8">
          {emptyState || (
            <div className="space-y-2">
              <div className="text-muted-foreground">
                {searchTerm ? 'No executions match your search' : 'No executions found'}
              </div>
              {showCreateButton && onCreateExecution && !searchTerm && (
                <Button variant="outline" onClick={onCreateExecution}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create First Execution
                </Button>
              )}
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {filteredExecutions.map((execution) => (
            <ExecutionItem
              key={execution.id}
              execution={execution}
              onUpdate={onUpdateExecution}
              onDelete={onDeleteExecution}
              onViewLogs={onViewLogs}
              onViewDetails={onViewDetails}
              compact={compact}
              expandable={expandable}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function StatusStatBadge({
  label,
  count,
  icon: Icon,
  onClick,
  active,
}: {
  label: string
  count: number
  icon: React.ComponentType<{ className?: string }>
  onClick: () => void
  active?: boolean
}) {
  if (count === 0) return null

  return (
    <Button
      variant={active ? "secondary" : "ghost"}
      size="sm"
      onClick={onClick}
      className={cn(
        'gap-1 h-auto py-1.5 px-2 text-xs font-medium',
        active && 'ring-2 ring-primary'
      )}
    >
      <Icon className="h-3 w-3" />
      <span>{label}</span>
      <Badge variant="secondary" className="ml-1 h-4 text-xs px-1">
        {count}
      </Badge>
    </Button>
  )
}