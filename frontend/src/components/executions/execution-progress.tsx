import type { ExecutionStatus } from '@/types/execution'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Progress } from '@/components/ui/progress'

interface ExecutionProgressProps {
  progress: number // 0.0 to 1.0
  status: ExecutionStatus
  size?: 'sm' | 'md' | 'lg'
  showPercentage?: boolean
  className?: string
}

const sizeClasses = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-3',
}

const _getProgressColor = (status: ExecutionStatus, progress: number) => {
  switch (status) {
    case 'running':
      return progress > 0.8 ? 'bg-green-500' : 'bg-blue-500'
    case 'completed':
      return 'bg-green-500'
    case 'failed':
      return 'bg-red-500'
    case 'cancelled':
      return 'bg-gray-400'
    case 'paused':
      return 'bg-yellow-500'
    case 'pending':
      return 'bg-gray-300'
    default:
      return 'bg-blue-500'
  }
}

const getProgressIcon = (status: ExecutionStatus, _progress: number) => {
  if (status === 'completed') return TrendingUp
  if (status === 'failed' || status === 'cancelled') return TrendingDown
  if (status === 'paused') return Minus
  return null
}

export function ExecutionProgress({
  progress,
  status,
  size = 'md',
  showPercentage = true,
  className,
}: ExecutionProgressProps) {
  const percentage = Math.round(progress * 100)

  const Icon = getProgressIcon(status, progress)

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className='flex-1'>
        <Progress
          value={percentage}
          className={cn('w-full', sizeClasses[size])}
        />
      </div>

      {showPercentage && (
        <div className='text-muted-foreground flex min-w-[3rem] items-center gap-1 text-sm font-medium'>
          {Icon && <Icon className='h-3 w-3' />}
          <span>{percentage}%</span>
        </div>
      )}
    </div>
  )
}
