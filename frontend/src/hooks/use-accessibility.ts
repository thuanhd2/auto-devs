import { useEffect, useRef, useState, useCallback } from 'react'

// Hook for managing focus
export function useFocus<T extends HTMLElement = HTMLElement>() {
  const ref = useRef<T>(null)

  const focus = useCallback(() => {
    ref.current?.focus()
  }, [])

  const blur = useCallback(() => {
    ref.current?.blur()
  }, [])

  const isFocused = useState(false)

  useEffect(() => {
    const element = ref.current
    if (!element) return

    const handleFocus = () => isFocused[1](true)
    const handleBlur = () => isFocused[1](false)

    element.addEventListener('focus', handleFocus)
    element.addEventListener('blur', handleBlur)

    return () => {
      element.removeEventListener('focus', handleFocus)
      element.removeEventListener('blur', handleBlur)
    }
  }, [isFocused])

  return {
    ref,
    focus,
    blur,
    isFocused: isFocused[0],
  }
}

// Hook for keyboard navigation
export function useKeyboardNavigation(options: {
  onEscape?: () => void
  onEnter?: () => void
  onArrowUp?: () => void
  onArrowDown?: () => void
  onArrowLeft?: () => void
  onArrowRight?: () => void
  onTab?: () => void
  onShiftTab?: () => void
  preventDefault?: boolean
}) {
  const {
    onEscape,
    onEnter,
    onArrowUp,
    onArrowDown,
    onArrowLeft,
    onArrowRight,
    onTab,
    onShiftTab,
    preventDefault = true,
  } = options

  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    const { key, shiftKey } = event

    switch (key) {
      case 'Escape':
        if (onEscape) {
          if (preventDefault) event.preventDefault()
          onEscape()
        }
        break
      case 'Enter':
        if (onEnter) {
          if (preventDefault) event.preventDefault()
          onEnter()
        }
        break
      case 'ArrowUp':
        if (onArrowUp) {
          if (preventDefault) event.preventDefault()
          onArrowUp()
        }
        break
      case 'ArrowDown':
        if (onArrowDown) {
          if (preventDefault) event.preventDefault()
          onArrowDown()
        }
        break
      case 'ArrowLeft':
        if (onArrowLeft) {
          if (preventDefault) event.preventDefault()
          onArrowLeft()
        }
        break
      case 'ArrowRight':
        if (onArrowRight) {
          if (preventDefault) event.preventDefault()
          onArrowRight()
        }
        break
      case 'Tab':
        if (shiftKey && onShiftTab) {
          if (preventDefault) event.preventDefault()
          onShiftTab()
        } else if (!shiftKey && onTab) {
          if (preventDefault) event.preventDefault()
          onTab()
        }
        break
    }
  }, [onEscape, onEnter, onArrowUp, onArrowDown, onArrowLeft, onArrowRight, onTab, onShiftTab, preventDefault])

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  return handleKeyDown
}

// Hook for managing ARIA announcements
export function useAnnouncer() {
  const [announcement, setAnnouncement] = useState('')
  const timeoutRef = useRef<NodeJS.Timeout>()

  const announce = useCallback((message: string, priority: 'polite' | 'assertive' = 'polite') => {
    // Clear any existing timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    // Set the announcement
    setAnnouncement(message)

    // Clear the announcement after a short delay to allow screen readers to pick it up
    timeoutRef.current = setTimeout(() => {
      setAnnouncement('')
    }, 1000)
  }, [])

  const announceSuccess = useCallback((message: string) => {
    announce(`Success: ${message}`, 'polite')
  }, [announce])

  const announceError = useCallback((message: string) => {
    announce(`Error: ${message}`, 'assertive')
  }, [announce])

  const announceInfo = useCallback((message: string) => {
    announce(`Info: ${message}`, 'polite')
  }, [announce])

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  return {
    announcement,
    announce,
    announceSuccess,
    announceError,
    announceInfo,
  }
}

// Hook for focus trap
export function useFocusTrap(active: boolean = true) {
  const containerRef = useRef<HTMLElement>(null)

  useEffect(() => {
    if (!active) return

    const container = containerRef.current
    if (!container) return

    const focusableElements = container.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )

    const firstElement = focusableElements[0] as HTMLElement
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return

      if (event.shiftKey) {
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement?.focus()
        }
      } else {
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement?.focus()
        }
      }
    }

    container.addEventListener('keydown', handleKeyDown)
    
    // Focus the first element when trap becomes active
    firstElement?.focus()

    return () => {
      container.removeEventListener('keydown', handleKeyDown)
    }
  }, [active])

  return containerRef
}

// Hook for reduced motion preference
export function useReducedMotion() {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false)

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    setPrefersReducedMotion(mediaQuery.matches)

    const handleChange = (event: MediaQueryListEvent) => {
      setPrefersReducedMotion(event.matches)
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  return prefersReducedMotion
}

// Hook for high contrast mode
export function useHighContrast() {
  const [prefersHighContrast, setPrefersHighContrast] = useState(false)

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-contrast: high)')
    setPrefersHighContrast(mediaQuery.matches)

    const handleChange = (event: MediaQueryListEvent) => {
      setPrefersHighContrast(event.matches)
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  return prefersHighContrast
}

// Hook for managing ARIA live regions
export function useLiveRegion() {
  const [liveMessage, setLiveMessage] = useState('')
  const [priority, setPriority] = useState<'polite' | 'assertive'>('polite')

  const announceToLiveRegion = useCallback((message: string, newPriority: 'polite' | 'assertive' = 'polite') => {
    setPriority(newPriority)
    setLiveMessage(message)

    // Clear the message after announcement
    setTimeout(() => {
      setLiveMessage('')
    }, 1000)
  }, [])

  return {
    liveMessage,
    priority,
    announceToLiveRegion,
  }
}

// Hook for skip links
export function useSkipLinks() {
  const skipToMain = useCallback(() => {
    const mainElement = document.querySelector('main') || document.querySelector('#main-content')
    if (mainElement) {
      (mainElement as HTMLElement).focus()
      mainElement.scrollIntoView()
    }
  }, [])

  const skipToNavigation = useCallback(() => {
    const navElement = document.querySelector('nav') || document.querySelector('#navigation')
    if (navElement) {
      (navElement as HTMLElement).focus()
      navElement.scrollIntoView()
    }
  }, [])

  const skipToContent = useCallback((selector: string) => {
    const element = document.querySelector(selector)
    if (element) {
      (element as HTMLElement).focus()
      element.scrollIntoView()
    }
  }, [])

  return {
    skipToMain,
    skipToNavigation,
    skipToContent,
  }
}

// Hook for managing aria-expanded state
export function useAriaExpanded(initialExpanded = false) {
  const [expanded, setExpanded] = useState(initialExpanded)

  const toggle = useCallback(() => {
    setExpanded(prev => !prev)
  }, [])

  const expand = useCallback(() => {
    setExpanded(true)
  }, [])

  const collapse = useCallback(() => {
    setExpanded(false)
  }, [])

  return {
    expanded,
    toggle,
    expand,
    collapse,
    ariaExpanded: expanded.toString(),
  }
}

// Hook for unique IDs (useful for ARIA relationships)
export function useUniqueId(prefix = 'id') {
  const id = useRef<string>()

  if (!id.current) {
    id.current = `${prefix}-${Math.random().toString(36).substr(2, 9)}`
  }

  return id.current
}