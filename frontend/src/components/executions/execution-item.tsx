import { useState } from 'react'
import type { Execution } from '@/types/execution'
import {
  Play,
  Pause,
  Square,
  Eye,
  AlertTriangle,
  ChevronRight,
  MoreHorizontal,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ExecutionDuration } from './execution-duration'
import { ExecutionLogsPannel } from './execution-logs-pannel'
import { ExecutionProgress } from './execution-progress'
import { ExecutionStatusBadge } from './execution-status-badge'

interface ExecutionItemProps {
  execution: Execution
  onUpdate?: (executionId: string, updates: Record<string, unknown>) => void
  showActions?: boolean
  compact?: boolean
  expandable?: boolean
  className?: string
}

export function ExecutionItem({
  execution,
  onUpdate,
  showActions = true,
  compact = false,
  expandable = false,
  className,
}: ExecutionItemProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const isActive =
    execution.status === 'running' || execution.status === 'pending'
  const hasError = execution.status === 'failed' || !!execution.error

  const handleStatusUpdate = (newStatus: string) => {
    onUpdate?.(execution.id, { status: newStatus })
  }

  return (
    <Card
      className={cn(
        'transition-all duration-200 hover:shadow-md',
        isActive && 'bg-blue-50/30 ring-2 ring-blue-200',
        hasError && 'border-red-200 bg-red-50/30',
        className
      )}
    >
      <CardContent className={cn('p-4', compact && 'p-3 py-0')}>
        {expandable ? (
          <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
            <CollapsibleTrigger asChild>
              <div className='flex cursor-pointer items-center justify-between py-2'>
                <ExecutionHeader execution={execution} compact={compact} />
                <div className='flex items-center gap-2'>
                  {showActions && (
                    <ExecutionActions
                      execution={execution}
                      onUpdate={handleStatusUpdate}
                    />
                  )}
                  <ChevronRight
                    className={cn(
                      'h-4 w-4 transition-transform duration-200',
                      isExpanded && 'rotate-90 transform'
                    )}
                  />
                </div>
              </div>
            </CollapsibleTrigger>
            <CollapsibleContent className='space-y-4'>
              <div className='border-t pt-4'>
                <ExecutionDetails execution={execution} />
              </div>
            </CollapsibleContent>
          </Collapsible>
        ) : (
          <div className='space-y-3'>
            <div className='flex items-start justify-between'>
              <ExecutionHeader execution={execution} compact={compact} />
              {showActions && (
                <ExecutionActions
                  execution={execution}
                  onUpdate={handleStatusUpdate}
                  onDelete={onDelete}
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

function ExecutionHeader({
  execution,
  compact,
}: {
  execution: Execution
  compact: boolean
}) {
  return (
    <div className='flex min-w-0 flex-1 items-start gap-3'>
      <div className='flex items-center gap-2'>
        <ExecutionStatusBadge
          status={execution.status}
          size={compact ? 'sm' : 'md'}
        />
        {execution.error && (
          <Badge variant='destructive' className='gap-1'>
            <AlertTriangle className='h-3 w-3' />
            Error
          </Badge>
        )}
      </div>

      <div className='min-w-0 flex-1'>
        <div className='text-muted-foreground flex items-center gap-2 text-sm'>
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
          <div className='mt-2'>
            <ExecutionProgress
              progress={execution.progress}
              status={execution.status}
              size='md'
            />
          </div>
        )}
      </div>
    </div>
  )
}

function ExecutionDetails({ execution }: { execution: Execution }) {
  return (
    <div className='space-y-3'>
      {execution.error && (
        <div className='rounded-lg border border-red-200 bg-red-50 p-3'>
          <div className='flex items-start gap-2'>
            <AlertTriangle className='mt-0.5 h-4 w-4 flex-shrink-0 text-red-500' />
            <div className='min-w-0 flex-1'>
              <div className='mb-1 text-sm font-medium text-red-800'>
                Execution Error
              </div>
              <div className='text-sm break-words text-red-700'>
                {execution.error}
              </div>
            </div>
          </div>
        </div>
      )}

      <ExecutionLogsPannel executionId={execution.id} />
    </div>
  )
}

function ExecutionActions({
  execution,
  onUpdate,
}: {
  execution: Execution
  onUpdate?: (status: string) => void
}) {
  const canPause = execution.status === 'running'
  const canResume = execution.status === 'paused'
  const canStop =
    execution.status === 'running' || execution.status === 'paused'
  const canRetry =
    execution.status === 'failed' || execution.status === 'cancelled'

  const dropdownItems = []
  if (canPause) {
    dropdownItems.push({
      label: 'Pause',
      icon: Pause,
      onClick: () => onUpdate?.('paused'),
    })
  }
  if (canResume) {
    dropdownItems.push({
      label: 'Resume',
      icon: Play,
      onClick: () => onUpdate?.('running'),
    })
  }
  if (canStop) {
    dropdownItems.push({
      label: 'Stop',
      icon: Square,
      onClick: () => onUpdate?.('cancelled'),
    })
  }
  if (canRetry) {
    dropdownItems.push({
      label: 'Retry',
      icon: Play,
      onClick: () => onUpdate?.('pending'),
    })
  }

  if (dropdownItems.length === 0) {
    return null
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' size='sm' className='h-8 w-8 p-0'>
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        {dropdownItems.map((item) => (
          <DropdownMenuItem key={item.label} onClick={item.onClick}>
            <item.icon className='mr-2 h-4 w-4' />
            {item.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
