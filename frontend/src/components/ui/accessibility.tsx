import * as React from 'react'
import { useLiveRegion } from '@/hooks/use-accessibility'
import { cn } from '@/lib/utils'

// Live Region for screen reader announcements
export function LiveRegion() {
  const { liveMessage, priority } = useLiveRegion()

  return (
    <div
      aria-live={priority}
      aria-atomic="true"
      className="sr-only"
      role="status"
    >
      {liveMessage}
    </div>
  )
}

// Screen reader only text
interface ScreenReaderOnlyProps {
  children: React.ReactNode
  as?: keyof JSX.IntrinsicElements
  className?: string
}

export function ScreenReaderOnly({ 
  children, 
  as: Component = 'span',
  className 
}: ScreenReaderOnlyProps) {
  return (
    <Component className={cn('sr-only', className)}>
      {children}
    </Component>
  )
}

// Skip link component
interface SkipLinkProps {
  href: string
  children: React.ReactNode
  className?: string
}

export function SkipLink({ href, children, className }: SkipLinkProps) {
  return (
    <a
      href={href}
      className={cn(
        'absolute left-0 top-0 z-50 -translate-y-full bg-primary px-4 py-2 text-primary-foreground transition-transform focus:translate-y-0',
        className
      )}
    >
      {children}
    </a>
  )
}

// Focus indicator component
interface FocusIndicatorProps {
  children: React.ReactNode
  className?: string
  visible?: boolean
}

export function FocusIndicator({ 
  children, 
  className,
  visible = true 
}: FocusIndicatorProps) {
  return (
    <div
      className={cn(
        visible && 'focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2',
        className
      )}
    >
      {children}
    </div>
  )
}

// ARIA description component
interface AriaDescriptionProps {
  id: string
  children: React.ReactNode
  className?: string
}

export function AriaDescription({ id, children, className }: AriaDescriptionProps) {
  return (
    <div
      id={id}
      className={cn('sr-only', className)}
      role="note"
    >
      {children}
    </div>
  )
}

// High contrast mode detector
export function HighContrastModeProvider({ children }: { children: React.ReactNode }) {
  const [isHighContrast, setIsHighContrast] = React.useState(false)

  React.useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-contrast: high)')
    setIsHighContrast(mediaQuery.matches)

    const handleChange = (event: MediaQueryListEvent) => {
      setIsHighContrast(event.matches)
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  React.useEffect(() => {
    if (isHighContrast) {
      document.documentElement.classList.add('high-contrast')
    } else {
      document.documentElement.classList.remove('high-contrast')
    }
  }, [isHighContrast])

  return <>{children}</>
}

// Reduced motion provider
export function ReducedMotionProvider({ children }: { children: React.ReactNode }) {
  const [prefersReducedMotion, setPrefersReducedMotion] = React.useState(false)

  React.useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    setPrefersReducedMotion(mediaQuery.matches)

    const handleChange = (event: MediaQueryListEvent) => {
      setPrefersReducedMotion(event.matches)
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  React.useEffect(() => {
    if (prefersReducedMotion) {
      document.documentElement.classList.add('reduce-motion')
    } else {
      document.documentElement.classList.remove('reduce-motion')
    }
  }, [prefersReducedMotion])

  return <>{children}</>
}

// Accessible button with enhanced ARIA support
interface AccessibleButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode
  description?: string
  expanded?: boolean
  pressed?: boolean
  loading?: boolean
}

export const AccessibleButton = React.forwardRef<HTMLButtonElement, AccessibleButtonProps>(
  ({ children, description, expanded, pressed, loading, className, ...props }, ref) => {
    const descriptionId = React.useId()

    return (
      <>
        <button
          ref={ref}
          className={cn(
            'focus-ring inline-flex items-center justify-center',
            loading && 'cursor-not-allowed opacity-50',
            className
          )}
          aria-describedby={description ? descriptionId : undefined}
          aria-expanded={expanded !== undefined ? expanded : undefined}
          aria-pressed={pressed !== undefined ? pressed : undefined}
          aria-busy={loading}
          disabled={loading || props.disabled}
          {...props}
        >
          {children}
          {loading && <ScreenReaderOnly>Loading</ScreenReaderOnly>}
        </button>
        {description && (
          <AriaDescription id={descriptionId}>
            {description}
          </AriaDescription>
        )}
      </>
    )
  }
)

AccessibleButton.displayName = 'AccessibleButton'

// Accessible form field with enhanced labeling
interface AccessibleFieldProps {
  id: string
  label: string
  description?: string
  error?: string
  required?: boolean
  children: React.ReactNode
}

export function AccessibleField({
  id,
  label,
  description,
  error,
  required,
  children,
}: AccessibleFieldProps) {
  const descriptionId = description ? `${id}-description` : undefined
  const errorId = error ? `${id}-error` : undefined
  
  const ariaDescribedBy = [descriptionId, errorId].filter(Boolean).join(' ')

  return (
    <div className="space-y-2">
      <label 
        htmlFor={id} 
        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        {label}
        {required && (
          <span className="ml-1 text-red-500" aria-label="required">
            *
          </span>
        )}
      </label>
      
      {React.cloneElement(children as React.ReactElement, {
        id,
        'aria-describedby': ariaDescribedBy || undefined,
        'aria-invalid': error ? 'true' : undefined,
        'aria-required': required,
      })}
      
      {description && (
        <p id={descriptionId} className="text-sm text-muted-foreground">
          {description}
        </p>
      )}
      
      {error && (
        <p id={errorId} className="text-sm text-red-500" role="alert">
          {error}
        </p>
      )}
    </div>
  )
}

// Accessible modal/dialog
interface AccessibleModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  description?: string
  children: React.ReactNode
}

export function AccessibleModal({
  isOpen,
  onClose,
  title,
  description,
  children,
}: AccessibleModalProps) {
  const titleId = React.useId()
  const descriptionId = React.useId()

  React.useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
      // Announce modal opening
      const announcement = `Dialog opened: ${title}`
      const announcer = document.createElement('div')
      announcer.setAttribute('aria-live', 'assertive')
      announcer.setAttribute('aria-atomic', 'true')
      announcer.className = 'sr-only'
      announcer.textContent = announcement
      document.body.appendChild(announcer)
      
      setTimeout(() => {
        document.body.removeChild(announcer)
      }, 1000)
    } else {
      document.body.style.overflow = ''
    }

    return () => {
      document.body.style.overflow = ''
    }
  }, [isOpen, title])

  if (!isOpen) return null

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
    >
      <div className="bg-background rounded-lg p-6 shadow-lg max-w-md w-full mx-4">
        <h2 id={titleId} className="text-lg font-semibold mb-4">
          {title}
        </h2>
        {description && (
          <p id={descriptionId} className="text-sm text-muted-foreground mb-4">
            {description}
          </p>
        )}
        {children}
        <button
          onClick={onClose}
          className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded focus-ring"
          aria-label="Close dialog"
        >
          Close
        </button>
      </div>
    </div>
  )
}