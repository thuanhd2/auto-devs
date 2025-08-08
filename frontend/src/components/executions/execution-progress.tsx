import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'
import type { ExecutionStatus } from '@/types/execution'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

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
      <div className="flex-1">
        <Progress 
          value={percentage}
          className={cn('w-full', sizeClasses[size])}
        />
      </div>
      
      {showPercentage && (
        <div className="flex items-center gap-1 text-sm font-medium text-muted-foreground min-w-[3rem]">
          {Icon && <Icon className="h-3 w-3" />}
          <span>{percentage}%</span>
        </div>
      )}
    </div>
  )
}

// Animated progress component for active executions
export function AnimatedExecutionProgress({
  progress,
  status,
  size = 'md',
  showPercentage = true,
  className,
}: ExecutionProgressProps) {
  const isActive = status === 'running' || status === 'pending'
  
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className="flex-1 relative">
        <Progress 
          value={Math.round(progress * 100)}
          className={cn(
            'w-full transition-all duration-300',
            sizeClasses[size],
            isActive && 'animate-pulse'
          )}
        />
        
        {/* Animated gradient overlay for running executions */}
        {status === 'running' && (
          <div 
            className={cn(
              'absolute inset-0 rounded-full opacity-30',
              'bg-gradient-to-r from-transparent via-white to-transparent',
              'animate-shimmer'
            )}
            style={{
              backgroundSize: '200% 100%',
              animation: 'shimmer 2s infinite linear',
            }}
          />
        )}
      </div>
      
      {showPercentage && (
        <div className="flex items-center gap-1 text-sm font-medium text-muted-foreground min-w-[3rem]">
          <span className={cn(
            'transition-all duration-300',
            isActive && 'text-blue-600'
          )}>
            {Math.round(progress * 100)}%
          </span>
        </div>
      )}
    </div>
  )
}

