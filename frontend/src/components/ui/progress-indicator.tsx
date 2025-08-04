import * as React from 'react'
import { Progress } from './progress'
import { cn } from '@/lib/utils'
import { CheckCircle, Circle, Clock, AlertCircle } from 'lucide-react'

export interface ProgressStep {
  id: string
  title: string
  description?: string
  status: 'pending' | 'current' | 'completed' | 'error'
}

interface ProgressIndicatorProps {
  steps: ProgressStep[]
  currentStep?: string
  variant?: 'linear' | 'circular' | 'dots'
  className?: string
  showLabels?: boolean
  size?: 'sm' | 'md' | 'lg'
}

export function ProgressIndicator({
  steps,
  currentStep,
  variant = 'linear',
  className,
  showLabels = true,
  size = 'md',
}: ProgressIndicatorProps) {
  const completedSteps = steps.filter(step => step.status === 'completed').length
  const totalSteps = steps.length
  const progressValue = (completedSteps / totalSteps) * 100

  const sizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  }

  const iconSize = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-6 w-6',
  }

  if (variant === 'circular') {
    return (
      <div className={cn('flex flex-col items-center space-y-4', className)}>
        <div className="relative">
          <svg className="h-20 w-20 transform -rotate-90" viewBox="0 0 100 100">
            <circle
              cx="50"
              cy="50"
              r="40"
              stroke="currentColor"
              strokeWidth="8"
              fill="none"
              className="text-muted-foreground/20"
            />
            <circle
              cx="50"
              cy="50"
              r="40"
              stroke="currentColor"
              strokeWidth="8"
              fill="none"
              strokeDasharray={`${2 * Math.PI * 40}`}
              strokeDashoffset={`${2 * Math.PI * 40 * (1 - progressValue / 100)}`}
              className="text-primary transition-all duration-500 ease-in-out"
              strokeLinecap="round"
            />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="text-2xl font-bold">{Math.round(progressValue)}%</span>
          </div>
        </div>
        {showLabels && (
          <div className="text-center">
            <p className="text-sm font-medium">
              {completedSteps} of {totalSteps} completed
            </p>
          </div>
        )}
      </div>
    )
  }

  if (variant === 'dots') {
    return (
      <div className={cn('flex items-center space-x-2', className)}>
        {steps.map((step, index) => (
          <div
            key={step.id}
            className={cn(
              'h-2 w-2 rounded-full transition-colors',
              step.status === 'completed' && 'bg-green-500',
              step.status === 'current' && 'bg-blue-500',
              step.status === 'pending' && 'bg-muted-foreground/30',
              step.status === 'error' && 'bg-red-500'
            )}
          />
        ))}
      </div>
    )
  }

  // Linear variant (default)
  return (
    <div className={cn('w-full space-y-4', className)}>
      <Progress value={progressValue} className="h-2" />
      
      {showLabels && (
        <div className="space-y-3">
          {steps.map((step, index) => {
            const StepIcon = ({ status }: { status: ProgressStep['status'] }) => {
              switch (status) {
                case 'completed':
                  return <CheckCircle className={cn(iconSize[size], 'text-green-500')} />
                case 'current':
                  return <Clock className={cn(iconSize[size], 'text-blue-500')} />
                case 'error':
                  return <AlertCircle className={cn(iconSize[size], 'text-red-500')} />
                default:
                  return <Circle className={cn(iconSize[size], 'text-muted-foreground')} />
              }
            }

            return (
              <div
                key={step.id}
                className={cn(
                  'flex items-start space-x-3',
                  step.status === 'current' && 'font-medium',
                  step.status === 'completed' && 'text-muted-foreground',
                  step.status === 'error' && 'text-red-600'
                )}
              >
                <StepIcon status={step.status} />
                <div className="flex-1 min-w-0">
                  <p className={cn(sizeClasses[size], 'font-medium')}>
                    {step.title}
                  </p>
                  {step.description && (
                    <p className={cn(sizeClasses[size], 'text-muted-foreground mt-1')}>
                      {step.description}
                    </p>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

// Task-specific progress indicator
interface TaskProgressProps {
  status: 'TODO' | 'PLANNING' | 'PLAN_REVIEWING' | 'IMPLEMENTING' | 'CODE_REVIEWING' | 'DONE' | 'CANCELLED'
  className?: string
}

export function TaskProgress({ status, className }: TaskProgressProps) {
  const steps: ProgressStep[] = [
    {
      id: 'todo',
      title: 'To Do',
      status: status === 'TODO' ? 'current' : 'pending',
    },
    {
      id: 'planning',
      title: 'Planning',
      status: status === 'PLANNING' ? 'current' : 
             ['PLAN_REVIEWING', 'IMPLEMENTING', 'CODE_REVIEWING', 'DONE'].includes(status) ? 'completed' : 'pending',
    },
    {
      id: 'plan_reviewing',
      title: 'Plan Review',
      status: status === 'PLAN_REVIEWING' ? 'current' : 
             ['IMPLEMENTING', 'CODE_REVIEWING', 'DONE'].includes(status) ? 'completed' : 'pending',
    },
    {
      id: 'implementing',
      title: 'Implementation',
      status: status === 'IMPLEMENTING' ? 'current' : 
             ['CODE_REVIEWING', 'DONE'].includes(status) ? 'completed' : 'pending',
    },
    {
      id: 'code_reviewing',
      title: 'Code Review',
      status: status === 'CODE_REVIEWING' ? 'current' : 
             status === 'DONE' ? 'completed' : 'pending',
    },
    {
      id: 'done',
      title: 'Done',
      status: status === 'DONE' ? 'completed' : 'pending',
    },
  ]

  if (status === 'CANCELLED') {
    steps.forEach(step => {
      if (step.status === 'current') {
        step.status = 'error'
      }
    })
  }

  return (
    <ProgressIndicator
      steps={steps}
      variant="linear"
      size="sm"
      className={className}
    />
  )
}