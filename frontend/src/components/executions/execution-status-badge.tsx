import type { ExecutionStatus } from '@/types/execution'
import { EXECUTION_STATUS_COLORS } from '@/types/execution'
import {
  Clock,
  Play,
  Pause,
  CheckCircle,
  XCircle,
  StopCircle,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'

interface ExecutionStatusBadgeProps {
  status: ExecutionStatus
  size?: 'sm' | 'md' | 'lg'
  showIcon?: boolean
  className?: string
}

const statusIcons: Record<
  ExecutionStatus,
  React.ComponentType<{ className?: string }>
> = {
  PENDING: Clock,
  RUNNING: Play,
  PAUSED: Pause,
  COMPLETED: CheckCircle,
  FAILED: XCircle,
  CANCELLED: StopCircle,
}

const sizeClasses = {
  sm: 'text-xs px-1.5 py-0.5',
  md: 'text-sm px-2 py-1',
  lg: 'text-base px-3 py-1.5',
}

const iconSizes = {
  sm: 'h-3 w-3',
  md: 'h-4 w-4',
  lg: 'h-5 w-5',
}

export function ExecutionStatusBadge({
  status,
  size = 'md',
  showIcon = true,
  className,
}: ExecutionStatusBadgeProps) {
  const Icon = statusIcons[status]
  const colorClasses = EXECUTION_STATUS_COLORS[status]

  return (
    <Badge
      variant='secondary'
      className={cn(
        'inline-flex items-center gap-1 font-medium',
        sizeClasses[size],
        colorClasses,
        className
      )}
    >
      {showIcon && <Icon className={cn(iconSizes[size])} />}
      <span className='capitalize'>{status}</span>
    </Badge>
  )
}
