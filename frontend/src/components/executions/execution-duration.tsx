import { cn } from '@/lib/utils'
import { Clock, Timer } from 'lucide-react'
import { useMemo, useState, useEffect } from 'react'

import type { ExecutionStatus } from '@/types/execution'

interface ExecutionDurationProps {
  startedAt: string
  completedAt?: string
  status: ExecutionStatus
  showIcon?: boolean
  showLabel?: boolean
  format?: 'short' | 'long' | 'human'
  className?: string
}

export function ExecutionDuration({
  startedAt,
  completedAt,
  status,
  showIcon = true,
  showLabel = false,
  format = 'short',
  className,
}: ExecutionDurationProps) {
  const [currentTime, setCurrentTime] = useState(Date.now())
  
  const isActive = status === 'running' || status === 'pending'

  // Update current time for active executions
  useEffect(() => {
    if (!isActive) return

    const interval = setInterval(() => {
      setCurrentTime(Date.now())
    }, 1000)

    return () => clearInterval(interval)
  }, [isActive])

  const duration = useMemo(() => {
    const startTime = new Date(startedAt).getTime()
    const endTime = completedAt ? new Date(completedAt).getTime() : currentTime
    
    return endTime - startTime
  }, [startedAt, completedAt, currentTime])

  const formattedDuration = useMemo(() => {
    const durationInSeconds = Math.floor(duration / 1000)
    
    switch (format) {
      case 'short':
        return formatShortDuration(durationInSeconds)
      case 'long':
        return formatLongDuration(durationInSeconds)
      case 'human':
        return formatHumanDuration(durationInSeconds)
      default:
        return formatShortDuration(durationInSeconds)
    }
  }, [duration, format])

  const Icon = isActive ? Timer : Clock

  return (
    <div className={cn(
      'inline-flex items-center gap-1 text-sm text-muted-foreground',
      className
    )}>
      {showIcon && (
        <Icon className={cn(
          'h-3 w-3',
          isActive && 'animate-pulse text-blue-500'
        )} />
      )}
      {showLabel && <span>Duration:</span>}
      <span className={cn(
        'font-mono',
        isActive && 'text-blue-600 font-medium'
      )}>
        {formattedDuration}
      </span>
    </div>
  )
}

// Real-time duration component that updates every second
export function LiveExecutionDuration({
  startedAt,
  completedAt,
  status,
  showIcon = true,
  className,
}: Omit<ExecutionDurationProps, 'format'>) {
  const [currentTime, setCurrentTime] = useState(Date.now())
  
  const isActive = status === 'running' || status === 'pending'

  useEffect(() => {
    if (!isActive) return

    const interval = setInterval(() => {
      setCurrentTime(Date.now())
    }, 1000)

    return () => clearInterval(interval)
  }, [isActive])

  const duration = useMemo(() => {
    const startTime = new Date(startedAt).getTime()
    const endTime = completedAt ? new Date(completedAt).getTime() : currentTime
    
    return Math.floor((endTime - startTime) / 1000)
  }, [startedAt, completedAt, currentTime])

  return (
    <div className={cn(
      'inline-flex items-center gap-1 text-sm',
      isActive ? 'text-blue-600 font-medium' : 'text-muted-foreground',
      className
    )}>
      {showIcon && (
        <Timer className={cn(
          'h-3 w-3',
          isActive && 'animate-pulse'
        )} />
      )}
      <span className="font-mono">
        {formatShortDuration(duration)}
      </span>
      {isActive && (
        <div className="flex items-center">
          <div className="h-1 w-1 rounded-full bg-green-500 animate-ping" />
        </div>
      )}
    </div>
  )
}

// Utility functions for duration formatting
function formatShortDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  } else if (minutes > 0) {
    return `${minutes}m ${remainingSeconds}s`
  } else {
    return `${remainingSeconds}s`
  }
}

function formatLongDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  const parts = []
  if (hours > 0) parts.push(`${hours} hour${hours !== 1 ? 's' : ''}`)
  if (minutes > 0) parts.push(`${minutes} minute${minutes !== 1 ? 's' : ''}`)
  if (remainingSeconds > 0 || parts.length === 0) {
    parts.push(`${remainingSeconds} second${remainingSeconds !== 1 ? 's' : ''}`)
  }

  return parts.join(', ')
}

function formatHumanDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) {
    return `${days}d ago`
  } else if (hours > 0) {
    return `${hours}h ago`
  } else if (minutes > 0) {
    return `${minutes}m ago`
  } else {
    return 'Just now'
  }
}