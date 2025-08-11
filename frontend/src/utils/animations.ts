import { toast } from 'sonner'

interface AnimationConfig {
  duration?: number
  easing?: string
  delay?: number
}

interface TaskAnimationOptions {
  showToast?: boolean
  toastMessage?: string
  highlightDuration?: number
  className?: string
}

/**
 * Animation utilities for real-time UI updates
 */
export class AnimationUtils {
  /**
   * Animate task appearance with a gentle scale and fade effect
   */
  static animateTaskAppear(
    element: HTMLElement,
    config: AnimationConfig = {}
  ): Promise<void> {
    const { duration = 300, easing = 'ease-out', delay = 0 } = config

    return new Promise((resolve) => {
      // Start from invisible and scaled down
      element.style.opacity = '0'
      element.style.transform = 'scale(0.95) translateY(-10px)'
      element.style.transition = `all ${duration}ms ${easing}`

      setTimeout(() => {
        element.style.opacity = '1'
        element.style.transform = 'scale(1) translateY(0)'

        setTimeout(() => {
          element.style.transition = ''
          resolve()
        }, duration)
      }, delay)
    })
  }

  /**
   * Animate task removal with fade and scale down
   */
  static animateTaskRemove(
    element: HTMLElement,
    config: AnimationConfig = {}
  ): Promise<void> {
    const { duration = 250, easing = 'ease-in' } = config

    return new Promise((resolve) => {
      element.style.transition = `all ${duration}ms ${easing}`
      element.style.opacity = '0'
      element.style.transform = 'scale(0.95) translateY(-10px)'
      element.style.height = '0'
      element.style.marginBottom = '0'
      element.style.paddingTop = '0'
      element.style.paddingBottom = '0'

      setTimeout(resolve, duration)
    })
  }

  /**
   * Animate task status change with color pulse
   */
  static animateStatusChange(
    element: HTMLElement,
    options: TaskAnimationOptions = {}
  ): void {
    const {
      highlightDuration = 2000,
      className = 'animate-pulse bg-blue-100 border-blue-300',
    } = options

    // Add highlight class
    const originalClasses = element.className
    element.className += ` ${className}`

    // Remove highlight after duration
    setTimeout(() => {
      element.className = originalClasses
    }, highlightDuration)

    // Show toast if requested
    if (options.showToast && options.toastMessage) {
      toast.info(options.toastMessage)
    }
  }

  /**
   * Animate task update with gentle highlight
   */
  static animateTaskUpdate(
    element: HTMLElement,
    options: TaskAnimationOptions = {}
  ): void {
    const {
      highlightDuration = 1500,
      className = 'animate-pulse bg-green-50 border-green-200',
    } = options

    this.animateStatusChange(element, {
      ...options,
      highlightDuration,
      className,
    })
  }

  /**
   * Animate column count update
   */
  static animateCountUpdate(element: HTMLElement): void {
    element.style.transition = 'transform 150ms ease-out'
    element.style.transform = 'scale(1.1)'

    setTimeout(() => {
      element.style.transform = 'scale(1)'
      setTimeout(() => {
        element.style.transition = ''
      }, 150)
    }, 150)
  }

  /**
   * Create a ripple effect for interactive elements
   */
  static createRippleEffect(element: HTMLElement, x: number, y: number): void {
    const ripple = document.createElement('div')
    const rect = element.getBoundingClientRect()
    const size = Math.max(rect.width, rect.height)
    const radius = size / 2

    ripple.style.width = ripple.style.height = `${size}px`
    ripple.style.left = `${x - rect.left - radius}px`
    ripple.style.top = `${y - rect.top - radius}px`
    ripple.style.borderRadius = '50%'
    ripple.style.position = 'absolute'
    ripple.style.background = 'rgba(255, 255, 255, 0.6)'
    ripple.style.transform = 'scale(0)'
    ripple.style.animation = 'ripple 600ms linear'
    ripple.style.pointerEvents = 'none'

    element.style.position = 'relative'
    element.style.overflow = 'hidden'
    element.appendChild(ripple)

    setTimeout(() => {
      ripple.remove()
    }, 600)
  }

  /**
   * Smooth height transition for collapsible elements
   */
  static animateHeight(
    element: HTMLElement,
    targetHeight: number,
    duration = 300
  ): Promise<void> {
    return new Promise((resolve) => {
      const startHeight = element.offsetHeight
      const startTime = performance.now()

      const animate = (currentTime: number) => {
        const elapsed = currentTime - startTime
        const progress = Math.min(elapsed / duration, 1)

        // Easing function (ease-out)
        const easeOut = 1 - Math.pow(1 - progress, 3)
        const currentHeight =
          startHeight + (targetHeight - startHeight) * easeOut

        element.style.height = `${currentHeight}px`

        if (progress < 1) {
          requestAnimationFrame(animate)
        } else {
          element.style.height = ''
          resolve()
        }
      }

      requestAnimationFrame(animate)
    })
  }
}

/**
 * Hook for managing task animations in React components
 */
export function useTaskAnimations() {
  const animateTaskCreated = (
    taskId: string,
    options?: TaskAnimationOptions
  ) => {
    const element = document.querySelector(
      `[data-task-id="${taskId}"]`
    ) as HTMLElement
    if (element) {
      AnimationUtils.animateTaskAppear(element)
      if (options?.showToast) {
        toast.success(options.toastMessage || 'New task created')
      }
    }
  }

  const animateTaskUpdated = (
    taskId: string,
    options?: TaskAnimationOptions
  ) => {
    const element = document.querySelector(
      `[data-task-id="${taskId}"]`
    ) as HTMLElement
    if (element) {
      AnimationUtils.animateTaskUpdate(element, options)
    }
  }

  const animateTaskStatusChanged = (
    taskId: string,
    newStatus: string,
    options?: TaskAnimationOptions
  ) => {
    const element = document.querySelector(
      `[data-task-id="${taskId}"]`
    ) as HTMLElement
    if (element) {
      AnimationUtils.animateStatusChange(element, {
        ...options,
        toastMessage: options?.toastMessage || `Task moved to ${newStatus}`,
        showToast: options?.showToast ?? true,
      })
    }
  }

  const animateTaskDeleted = (
    taskId: string,
    options?: TaskAnimationOptions
  ) => {
    const element = document.querySelector(
      `[data-task-id="${taskId}"]`
    ) as HTMLElement
    if (element) {
      return AnimationUtils.animateTaskRemove(element).then(() => {
        if (options?.showToast) {
          toast.info(options.toastMessage || 'Task deleted')
        }
      })
    }
    return Promise.resolve()
  }

  const animateColumnCountUpdate = (status: string) => {
    const element = document.querySelector(
      `[data-column="${status}"] .task-count`
    ) as HTMLElement
    if (element) {
      AnimationUtils.animateCountUpdate(element)
    }
  }

  return {
    animateTaskCreated,
    animateTaskUpdated,
    animateTaskStatusChanged,
    animateTaskDeleted,
    animateColumnCountUpdate,
  }
}
