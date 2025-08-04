import * as React from 'react'
import { cn } from '@/lib/utils'

interface MicroInteractionProps {
  children: React.ReactNode
  className?: string
  hover?: 'scale' | 'lift' | 'glow' | 'bounce' | 'rotate' | 'pulse'
  click?: 'scale' | 'bounce' | 'ripple' | 'shake'
  focus?: boolean
  disabled?: boolean
}

export function MicroInteraction({
  children,
  className,
  hover = 'scale',
  click = 'scale',
  focus = true,
  disabled = false,
}: MicroInteractionProps) {
  const [isClicked, setIsClicked] = React.useState(false)

  const handleClick = () => {
    if (disabled) return
    setIsClicked(true)
    setTimeout(() => setIsClicked(false), 150)
  }

  const hoverClasses = {
    scale: 'hover:scale-105',
    lift: 'hover:-translate-y-1 hover:shadow-lg',
    glow: 'hover:shadow-lg hover:shadow-primary/25',
    bounce: 'hover:animate-bounce',
    rotate: 'hover:rotate-3',
    pulse: 'hover:animate-pulse',
  }

  const clickClasses = {
    scale: isClicked ? 'scale-95' : '',
    bounce: isClicked ? 'animate-bounce' : '',
    ripple: isClicked ? 'relative overflow-hidden' : '',
    shake: isClicked ? 'animate-shake' : '',
  }

  return (
    <div
      className={cn(
        'transition-all duration-200 ease-in-out',
        !disabled && hoverClasses[hover],
        !disabled && clickClasses[click],
        focus && !disabled && 'focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2',
        disabled && 'opacity-50 cursor-not-allowed',
        className
      )}
      onClick={handleClick}
    >
      {children}
    </div>
  )
}

// Specialized hover card with micro-interactions
interface InteractiveCardProps {
  children: React.ReactNode
  className?: string
  clickable?: boolean
  onClick?: () => void
}

export function InteractiveCard({ 
  children, 
  className, 
  clickable = false, 
  onClick 
}: InteractiveCardProps) {
  return (
    <div
      className={cn(
        'rounded-lg border bg-card p-4 transition-all duration-200',
        'hover:shadow-md hover:-translate-y-0.5',
        clickable && 'hover:shadow-lg hover:border-primary/20 cursor-pointer active:scale-[0.98]',
        className
      )}
      onClick={onClick}
    >
      {children}
    </div>
  )
}

// Floating action button with micro-interactions
interface FloatingButtonProps {
  children: React.ReactNode
  className?: string
  onClick?: () => void
}

export function FloatingButton({ children, className, onClick }: FloatingButtonProps) {
  return (
    <button
      className={cn(
        'fixed bottom-6 right-6 h-14 w-14 rounded-full bg-primary text-primary-foreground shadow-lg',
        'hover:shadow-xl hover:scale-110 active:scale-95',
        'transition-all duration-200 ease-in-out',
        'focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
        className
      )}
      onClick={onClick}
    >
      {children}
    </button>
  )
}

// Animated counter
interface AnimatedCounterProps {
  value: number
  duration?: number
  className?: string
}

export function AnimatedCounter({ value, duration = 1000, className }: AnimatedCounterProps) {
  const [count, setCount] = React.useState(0)

  React.useEffect(() => {
    let startTime: number
    let animationFrame: number

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp
      const progress = Math.min((timestamp - startTime) / duration, 1)
      
      setCount(Math.floor(progress * value))
      
      if (progress < 1) {
        animationFrame = requestAnimationFrame(animate)
      }
    }

    animationFrame = requestAnimationFrame(animate)

    return () => {
      if (animationFrame) {
        cancelAnimationFrame(animationFrame)
      }
    }
  }, [value, duration])

  return <span className={className}>{count}</span>
}

// Staggered fade-in animation for lists
interface StaggeredFadeInProps {
  children: React.ReactNode[]
  delay?: number
  className?: string
}

export function StaggeredFadeIn({ children, delay = 100, className }: StaggeredFadeInProps) {
  return (
    <div className={className}>
      {children.map((child, index) => (
        <div
          key={index}
          className="animate-fade-in opacity-0"
          style={{
            animationDelay: `${index * delay}ms`,
            animationFillMode: 'forwards',
          }}
        >
          {child}
        </div>
      ))}
    </div>
  )
}

// Morphing icon button
interface MorphingIconProps {
  icon1: React.ReactNode
  icon2: React.ReactNode
  toggled: boolean
  className?: string
}

export function MorphingIcon({ icon1, icon2, toggled, className }: MorphingIconProps) {
  return (
    <div className={cn('relative', className)}>
      <div
        className={cn(
          'absolute inset-0 transition-all duration-300',
          toggled ? 'opacity-0 rotate-90 scale-0' : 'opacity-100 rotate-0 scale-100'
        )}
      >
        {icon1}
      </div>
      <div
        className={cn(
          'transition-all duration-300',
          toggled ? 'opacity-100 rotate-0 scale-100' : 'opacity-0 -rotate-90 scale-0'
        )}
      >
        {icon2}
      </div>
    </div>
  )
}