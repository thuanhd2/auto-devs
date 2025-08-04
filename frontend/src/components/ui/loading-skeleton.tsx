import { cn } from '@/lib/utils'
import { Skeleton } from './skeleton'

interface LoadingSkeletonProps {
  variant?: 'card' | 'list' | 'table' | 'avatar' | 'text' | 'custom'
  rows?: number
  className?: string
  children?: React.ReactNode
}

export function LoadingSkeleton({ 
  variant = 'card', 
  rows = 3, 
  className,
  children 
}: LoadingSkeletonProps) {
  if (children) {
    return <div className={className}>{children}</div>
  }

  switch (variant) {
    case 'card':
      return (
        <div className={cn('space-y-4', className)}>
          {Array.from({ length: rows }).map((_, i) => (
            <div key={i} className="rounded-lg border p-4 space-y-3">
              <div className="flex items-center space-x-3">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
              <Skeleton className="h-20 w-full" />
              <div className="flex space-x-2">
                <Skeleton className="h-8 w-16" />
                <Skeleton className="h-8 w-20" />
              </div>
            </div>
          ))}
        </div>
      )

    case 'list':
      return (
        <div className={cn('space-y-2', className)}>
          {Array.from({ length: rows }).map((_, i) => (
            <div key={i} className="flex items-center space-x-3 p-2">
              <Skeleton className="h-8 w-8 rounded-full" />
              <div className="space-y-1 flex-1">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
              <Skeleton className="h-6 w-16" />
            </div>
          ))}
        </div>
      )

    case 'table':
      return (
        <div className={cn('space-y-2', className)}>
          <div className="grid grid-cols-4 gap-4 p-4 border-b">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
          </div>
          {Array.from({ length: rows }).map((_, i) => (
            <div key={i} className="grid grid-cols-4 gap-4 p-4">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
            </div>
          ))}
        </div>
      )

    case 'avatar':
      return (
        <div className={cn('flex items-center space-x-3', className)}>
          <Skeleton className="h-12 w-12 rounded-full" />
          <div className="space-y-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-3 w-24" />
          </div>
        </div>
      )

    case 'text':
      return (
        <div className={cn('space-y-2', className)}>
          {Array.from({ length: rows }).map((_, i) => (
            <Skeleton 
              key={i} 
              className={cn(
                'h-4',
                i === rows - 1 ? 'w-3/4' : 'w-full'
              )} 
            />
          ))}
        </div>
      )

    default:
      return (
        <div className={cn('space-y-3', className)}>
          {Array.from({ length: rows }).map((_, i) => (
            <Skeleton key={i} className="h-4 w-full" />
          ))}
        </div>
      )
  }
}

// Specialized skeleton components
export function TaskCardSkeleton() {
  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <div className="flex items-start justify-between">
        <Skeleton className="h-5 w-3/4" />
        <Skeleton className="h-6 w-16 rounded-full" />
      </div>
      <Skeleton className="h-16 w-full" />
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Skeleton className="h-6 w-6 rounded-full" />
          <Skeleton className="h-4 w-20" />
        </div>
        <Skeleton className="h-4 w-12" />
      </div>
    </div>
  )
}

export function ProjectCardSkeleton() {
  return (
    <div className="rounded-lg border bg-card p-6 space-y-4">
      <div className="flex items-center space-x-3">
        <Skeleton className="h-12 w-12 rounded-lg" />
        <div className="space-y-2 flex-1">
          <Skeleton className="h-6 w-2/3" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      </div>
      <Skeleton className="h-20 w-full" />
      <div className="flex items-center justify-between">
        <div className="flex space-x-2">
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-6 w-20" />
        </div>
        <Skeleton className="h-4 w-24" />
      </div>
    </div>
  )
}

export function KanbanColumnSkeleton() {
  return (
    <div className="w-80 space-y-3">
      <div className="flex items-center justify-between p-4 border-b">
        <Skeleton className="h-5 w-24" />
        <Skeleton className="h-6 w-8 rounded-full" />
      </div>
      <div className="space-y-3 p-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <TaskCardSkeleton key={i} />
        ))}
      </div>
    </div>
  )
}