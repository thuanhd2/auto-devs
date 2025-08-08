import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { ExecutionStatusBadge } from './execution-status-badge'
import { ExecutionProgress } from './execution-progress'
import { ExecutionDuration } from './execution-duration'
import type { Execution } from '@/types/execution'
import { 
  Play, 
  Pause, 
  Square, 
  Trash2, 
  Eye, 
  AlertTriangle,
  ChevronRight,
  MoreHorizontal,
} from 'lucide-react'
import { useState } from 'react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'

interface ExecutionItemProps {
  execution: Execution
  onUpdate?: (executionId: string, updates: Record<string, unknown>) => void
  onDelete?: (executionId: string) => void
  onViewLogs?: (executionId: string) => void
  onViewDetails?: (executionId: string) => void
  showActions?: boolean
  compact?: boolean
  expandable?: boolean
  className?: string
}

export function ExecutionItem({
  execution,
  onUpdate,
  onDelete,
  onViewLogs,
  onViewDetails,
  showActions = true,
  compact = false,
  expandable = false,
  className,
}: ExecutionItemProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const isActive = execution.status === 'running' || execution.status === 'pending'
  const hasError = execution.status === 'failed' || !!execution.error

  const handleStatusUpdate = (newStatus: string) => {
    onUpdate?.(execution.id, { status: newStatus })
  }



  return (
    <Card className={cn(
      'transition-all duration-200 hover:shadow-md',
      isActive && 'ring-2 ring-blue-200 bg-blue-50/30',
      hasError && 'border-red-200 bg-red-50/30',
      className
    )}>
      <CardContent className={cn('p-4', compact && 'p-3')}>
        {expandable ? (
          <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
            <CollapsibleTrigger asChild>
              <div className="flex items-center justify-between cursor-pointer">
                <ExecutionHeader execution={execution} compact={compact} />
                <div className="flex items-center gap-2">
                  {showActions && (
                    <ExecutionActions 
                      execution={execution}
                      onUpdate={handleStatusUpdate}
                      onDelete={onDelete}
                      onViewLogs={onViewLogs}
                      onViewDetails={onViewDetails}
                    />
                  )}
                  <ChevronRight className={cn(
                    'h-4 w-4 transition-transform duration-200',
                    isExpanded && 'transform rotate-90'
                  )} />
                </div>
              </div>
            </CollapsibleTrigger>
            <CollapsibleContent className="space-y-4">
              <div className="pt-4 border-t">
                <ExecutionDetails execution={execution} />
              </div>
            </CollapsibleContent>
          </Collapsible>
        ) : (
          <div className="space-y-3">
            <div className="flex items-start justify-between">
              <ExecutionHeader execution={execution} compact={compact} />
              {showActions && (
                <ExecutionActions 
                  execution={execution}
                  onUpdate={handleStatusUpdate}
                  onDelete={onDelete}
                  onViewLogs={onViewLogs}
                  onViewDetails={onViewDetails}
                />
              )}
            </div>
            {!compact && <ExecutionDetails execution={execution} />}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ExecutionHeader({ execution, compact }: { execution: Execution; compact: boolean }) {
  return (
    <div className="flex items-start gap-3 flex-1 min-w-0">
      <div className="flex items-center gap-2">
        <ExecutionStatusBadge 
          status={execution.status} 
          size={compact ? 'sm' : 'md'}
        />
        {execution.error && (
          <Badge variant="destructive" className="gap-1">
            <AlertTriangle className="h-3 w-3" />
            Error
          </Badge>
        )}
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>#{execution.id.slice(-8)}</span>
          <span>•</span>
          <ExecutionDuration 
            startedAt={execution.started_at}
            completedAt={execution.completed_at}
            status={execution.status}
            showIcon={false}
          />
        </div>
        
        {!compact && (
          <div className="mt-2">
            <ExecutionProgress 
              progress={execution.progress}
              status={execution.status}
              size="md"
            />
          </div>
        )}
      </div>
    </div>
  )
}

function ExecutionDetails({ execution }: { execution: Execution }) {
  return (
    <div className="space-y-3">
      <ExecutionProgress 
        progress={execution.progress}
        status={execution.status}
        size="md"
      />
      
      <div className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span className="text-muted-foreground">Started:</span>
          <div className="font-mono text-xs">
            {new Date(execution.started_at).toLocaleString()}
          </div>
        </div>
        
        {execution.completed_at && (
          <div>
            <span className="text-muted-foreground">Completed:</span>
            <div className="font-mono text-xs">
              {new Date(execution.completed_at).toLocaleString()}
            </div>
          </div>
        )}
      </div>
      
      {execution.error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-start gap-2">
            <AlertTriangle className="h-4 w-4 text-red-500 mt-0.5 flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium text-red-800 mb-1">
                Execution Error
              </div>
              <div className="text-sm text-red-700 break-words">
                {execution.error}
              </div>
            </div>
          </div>
        </div>
      )}
      
      {execution.result && (
        <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
          <div className="text-sm font-medium text-green-800 mb-2">
            Execution Result
          </div>
          <div className="space-y-1 text-sm text-green-700">
            {execution.result.files && execution.result.files.length > 0 && (
              <div>
                <span className="font-medium">Files:</span> {execution.result.files.length}
              </div>
            )}
            {execution.result.output && (
              <div>
                <span className="font-medium">Output:</span>
                <div className="mt-1 p-2 bg-white border rounded font-mono text-xs max-h-20 overflow-y-auto">
                  {execution.result.output}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function ExecutionActions({
  execution,
  onUpdate,
  onDelete,
  onViewLogs,
  onViewDetails,
}: {
  execution: Execution
  onUpdate?: (status: string) => void
  onDelete?: (executionId: string) => void
  onViewLogs?: (executionId: string) => void
  onViewDetails?: (executionId: string) => void
}) {
  const canPause = execution.status === 'running'
  const canResume = execution.status === 'paused'
  const canStop = execution.status === 'running' || execution.status === 'paused'
  const canRetry = execution.status === 'failed' || execution.status === 'cancelled'

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
          <MoreHorizontal className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => onViewDetails?.(execution.id)}>
          <Eye className="mr-2 h-4 w-4" />
          View Details
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => onViewLogs?.(execution.id)}>
          <Eye className="mr-2 h-4 w-4" />
          View Logs
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        
        {canPause && (
          <DropdownMenuItem onClick={() => onUpdate?.('paused')}>
            <Pause className="mr-2 h-4 w-4" />
            Pause
          </DropdownMenuItem>
        )}
        
        {canResume && (
          <DropdownMenuItem onClick={() => onUpdate?.('running')}>
            <Play className="mr-2 h-4 w-4" />
            Resume
          </DropdownMenuItem>
        )}
        
        {canStop && (
          <DropdownMenuItem onClick={() => onUpdate?.('cancelled')}>
            <Square className="mr-2 h-4 w-4" />
            Stop
          </DropdownMenuItem>
        )}
        
        {canRetry && (
          <DropdownMenuItem onClick={() => onUpdate?.('pending')}>
            <Play className="mr-2 h-4 w-4" />
            Retry
          </DropdownMenuItem>
        )}
        
        <DropdownMenuSeparator />
        <DropdownMenuItem 
          onClick={() => onDelete?.(execution.id)}
          className="text-red-600 focus:text-red-600"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}