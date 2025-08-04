import * as React from 'react'
import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { X, ChevronLeft, ChevronRight, Target, CheckCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface OnboardingStep {
  id: string
  title: string
  description: string
  content?: React.ReactNode
  target?: string // CSS selector for element to highlight
  position?: 'top' | 'bottom' | 'left' | 'right' | 'center'
  showSkip?: boolean
  action?: {
    label: string
    onClick: () => void
  }
}

interface OnboardingTourProps {
  steps: OnboardingStep[]
  isActive: boolean
  onComplete: () => void
  onSkip: () => void
  className?: string
  showProgress?: boolean
  allowSkip?: boolean
}

export function OnboardingTour({
  steps,
  isActive,
  onComplete,
  onSkip,
  className,
  showProgress = true,
  allowSkip = true,
}: OnboardingTourProps) {
  const [currentStepIndex, setCurrentStepIndex] = useState(0)
  const [targetElement, setTargetElement] = useState<HTMLElement | null>(null)
  const [tooltipPosition, setTooltipPosition] = useState({ top: 0, left: 0 })

  const currentStep = steps[currentStepIndex]
  const isFirstStep = currentStepIndex === 0
  const isLastStep = currentStepIndex === steps.length - 1
  const progress = ((currentStepIndex + 1) / steps.length) * 100

  // Update target element and position when step changes
  useEffect(() => {
    if (!isActive || !currentStep?.target) {
      setTargetElement(null)
      return
    }

    const element = document.querySelector(currentStep.target) as HTMLElement
    if (element) {
      setTargetElement(element)
      updateTooltipPosition(element, currentStep.position || 'bottom')
      
      // Scroll element into view
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
        inline: 'nearest',
      })
      
      // Add highlight class
      element.classList.add('onboarding-highlight')
    }

    return () => {
      // Clean up highlight
      if (element) {
        element.classList.remove('onboarding-highlight')
      }
    }
  }, [currentStep, isActive])

  // Calculate tooltip position
  const updateTooltipPosition = (element: HTMLElement, position: string) => {
    const rect = element.getBoundingClientRect()
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop
    const scrollLeft = window.pageXOffset || document.documentElement.scrollLeft

    let top = 0
    let left = 0

    switch (position) {
      case 'top':
        top = rect.top + scrollTop - 20
        left = rect.left + scrollLeft + rect.width / 2
        break
      case 'bottom':
        top = rect.bottom + scrollTop + 20
        left = rect.left + scrollLeft + rect.width / 2
        break
      case 'left':
        top = rect.top + scrollTop + rect.height / 2
        left = rect.left + scrollLeft - 20
        break
      case 'right':
        top = rect.top + scrollTop + rect.height / 2
        left = rect.right + scrollLeft + 20
        break
      case 'center':
      default:
        top = window.innerHeight / 2 + scrollTop
        left = window.innerWidth / 2 + scrollLeft
        break
    }

    setTooltipPosition({ top, left })
  }

  // Navigation functions
  const goToNext = useCallback(() => {
    if (isLastStep) {
      onComplete()
    } else {
      setCurrentStepIndex(prev => prev + 1)
    }
  }, [isLastStep, onComplete])

  const goToPrevious = useCallback(() => {
    if (!isFirstStep) {
      setCurrentStepIndex(prev => prev - 1)
    }
  }, [isFirstStep])

  const handleSkip = useCallback(() => {
    onSkip()
  }, [onSkip])

  // Handle keyboard navigation
  useEffect(() => {
    if (!isActive) return

    const handleKeyDown = (event: KeyboardEvent) => {
      switch (event.key) {
        case 'ArrowRight':
        case 'Enter':
          event.preventDefault()
          goToNext()
          break
        case 'ArrowLeft':
          event.preventDefault()
          goToPrevious()
          break
        case 'Escape':
          event.preventDefault()
          handleSkip()
          break
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isActive, goToNext, goToPrevious, handleSkip])

  if (!isActive || !currentStep) {
    return null
  }

  return (
    <>
      {/* Overlay */}
      <div className="fixed inset-0 z-40 bg-black/50" />
      
      {/* Spotlight for highlighted element */}
      {targetElement && (
        <div
          className="fixed z-50 pointer-events-none"
          style={{
            top: targetElement.getBoundingClientRect().top + window.pageYOffset - 4,
            left: targetElement.getBoundingClientRect().left + window.pageXOffset - 4,
            width: targetElement.offsetWidth + 8,
            height: targetElement.offsetHeight + 8,
            boxShadow: '0 0 0 4px rgba(59, 130, 246, 0.5), 0 0 0 9999px rgba(0, 0, 0, 0.5)',
            borderRadius: '4px',
          }}
        />
      )}

      {/* Tooltip */}
      <Card
        className={cn(
          'fixed z-50 w-80 max-w-sm shadow-lg',
          currentStep.position === 'center' && 'transform -translate-x-1/2 -translate-y-1/2',
          currentStep.position === 'top' && 'transform -translate-x-1/2 -translate-y-full',
          currentStep.position === 'bottom' && 'transform -translate-x-1/2',
          currentStep.position === 'left' && 'transform -translate-x-full -translate-y-1/2',
          currentStep.position === 'right' && 'transform -translate-y-1/2',
          className
        )}
        style={{
          top: tooltipPosition.top,
          left: tooltipPosition.left,
        }}
      >
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <Target className="h-4 w-4 text-primary" />
                <CardTitle className="text-lg">{currentStep.title}</CardTitle>
              </div>
              <CardDescription>{currentStep.description}</CardDescription>
            </div>
            {allowSkip && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSkip}
                className="h-6 w-6 p-0"
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>
        </CardHeader>

        {currentStep.content && (
          <CardContent className="py-0">
            {currentStep.content}
          </CardContent>
        )}

        <CardFooter className="flex flex-col space-y-4 pt-4">
          {/* Progress */}
          {showProgress && (
            <div className="w-full space-y-2">
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Step {currentStepIndex + 1} of {steps.length}</span>
                <span>{Math.round(progress)}%</span>
              </div>
              <Progress value={progress} className="h-1" />
            </div>
          )}

          {/* Actions */}
          <div className="flex w-full justify-between">
            <Button
              variant="outline"
              size="sm"
              onClick={goToPrevious}
              disabled={isFirstStep}
            >
              <ChevronLeft className="mr-1 h-3 w-3" />
              Previous
            </Button>

            <div className="flex gap-2">
              {currentStep.action && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={currentStep.action.onClick}
                >
                  {currentStep.action.label}
                </Button>
              )}
              
              <Button size="sm" onClick={goToNext}>
                {isLastStep ? (
                  <>
                    <CheckCircle className="mr-1 h-3 w-3" />
                    Complete
                  </>
                ) : (
                  <>
                    Next
                    <ChevronRight className="ml-1 h-3 w-3" />
                  </>
                )}
              </Button>
            </div>
          </div>
        </CardFooter>
      </Card>
    </>
  )
}

// Hook for managing onboarding state
export function useOnboarding(steps: OnboardingStep[], storageKey = 'onboarding-complete') {
  const [isActive, setIsActive] = useState(false)
  const [isCompleted, setIsCompleted] = useState(false)

  // Check if onboarding was already completed
  useEffect(() => {
    const completed = localStorage.getItem(storageKey) === 'true'
    setIsCompleted(completed)
  }, [storageKey])

  const startOnboarding = useCallback(() => {
    setIsActive(true)
  }, [])

  const completeOnboarding = useCallback(() => {
    setIsActive(false)
    setIsCompleted(true)
    localStorage.setItem(storageKey, 'true')
  }, [storageKey])

  const skipOnboarding = useCallback(() => {
    setIsActive(false)
    setIsCompleted(true)
    localStorage.setItem(storageKey, 'true')
  }, [storageKey])

  const resetOnboarding = useCallback(() => {
    setIsCompleted(false)
    setIsActive(false)
    localStorage.removeItem(storageKey)
  }, [storageKey])

  return {
    isActive,
    isCompleted,
    startOnboarding,
    completeOnboarding,
    skipOnboarding,
    resetOnboarding,
  }
}

// Welcome modal component
interface WelcomeModalProps {
  isOpen: boolean
  onStart: () => void
  onSkip: () => void
  title?: string
  description?: string
  features?: string[]
}

export function WelcomeModal({
  isOpen,
  onStart,
  onSkip,
  title = 'Welcome to the App!',
  description = 'Let us show you around and help you get started.',
  features = [],
}: WelcomeModalProps) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md mx-4">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">{title}</CardTitle>
          <CardDescription className="text-base">{description}</CardDescription>
        </CardHeader>

        {features.length > 0 && (
          <CardContent>
            <h4 className="mb-3 font-medium">What you'll learn:</h4>
            <ul className="space-y-2">
              {features.map((feature, index) => (
                <li key={index} className="flex items-center gap-2 text-sm">
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  {feature}
                </li>
              ))}
            </ul>
          </CardContent>
        )}

        <CardFooter className="flex gap-2">
          <Button variant="outline" onClick={onSkip} className="flex-1">
            Skip Tour
          </Button>
          <Button onClick={onStart} className="flex-1">
            Start Tour
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}