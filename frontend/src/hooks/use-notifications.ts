import { useState, useCallback, useRef } from 'react'
import { toast } from 'sonner'

export type NotificationType = 'success' | 'error' | 'warning' | 'info' | 'loading'

export interface NotificationItem {
  id: string
  type: NotificationType
  title: string
  description?: string
  timestamp: Date
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
  data?: any
}

export interface NotificationOptions {
  duration?: number
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center'
  dismissible?: boolean
  action?: {
    label: string
    onClick: () => void
  }
  onDismiss?: () => void
  richColors?: boolean
  closeButton?: boolean
}

const DEFAULT_DURATIONS = {
  success: 4000,
  error: 6000,
  warning: 5000,
  info: 4000,
  loading: Infinity,
}

export function useNotifications() {
  const [history, setHistory] = useState<NotificationItem[]>([])
  const notificationIdRef = useRef(0)

  const generateId = () => `notification-${++notificationIdRef.current}`

  const addToHistory = useCallback((notification: NotificationItem) => {
    setHistory(prev => [notification, ...prev].slice(0, 100)) // Keep last 100 notifications
  }, [])

  const showNotification = useCallback((
    type: NotificationType,
    title: string,
    description?: string,
    options: NotificationOptions = {}
  ) => {
    const id = generateId()
    const duration = options.duration ?? DEFAULT_DURATIONS[type]
    const timestamp = new Date()

    const notification: NotificationItem = {
      id,
      type,
      title,
      description,
      timestamp,
      duration: duration === Infinity ? undefined : duration,
      action: options.action,
    }

    // Add to history
    addToHistory(notification)

    // Show toast notification
    const toastOptions = {
      id,
      duration: duration === Infinity ? Infinity : duration,
      dismissible: options.dismissible ?? true,
      closeButton: options.closeButton ?? true,
      description,
      action: options.action ? {
        label: options.action.label,
        onClick: options.action.onClick,
      } : undefined,
      onDismiss: options.onDismiss,
    }

    switch (type) {
      case 'success':
        return toast.success(title, toastOptions)
      case 'error':
        return toast.error(title, toastOptions)
      case 'warning':
        return toast.warning(title, toastOptions)
      case 'info':
        return toast.info(title, toastOptions)
      case 'loading':
        return toast.loading(title, toastOptions)
      default:
        return toast(title, toastOptions)
    }
  }, [addToHistory])

  const success = useCallback((title: string, description?: string, options?: NotificationOptions) => {
    return showNotification('success', title, description, options)
  }, [showNotification])

  const error = useCallback((title: string, description?: string, options?: NotificationOptions) => {
    return showNotification('error', title, description, options)
  }, [showNotification])

  const warning = useCallback((title: string, description?: string, options?: NotificationOptions) => {
    return showNotification('warning', title, description, options)
  }, [showNotification])

  const info = useCallback((title: string, description?: string, options?: NotificationOptions) => {
    return showNotification('info', title, description, options)
  }, [showNotification])

  const loading = useCallback((title: string, description?: string, options?: NotificationOptions) => {
    return showNotification('loading', title, description, options)
  }, [showNotification])

  const dismiss = useCallback((toastId: string) => {
    toast.dismiss(toastId)
  }, [])

  const dismissAll = useCallback(() => {
    toast.dismiss()
  }, [])

  const promise = useCallback(<T,>(
    promise: Promise<T>,
    {
      loading: loadingMessage,
      success: successMessage,
      error: errorMessage,
    }: {
      loading: string
      success: string | ((data: T) => string)
      error: string | ((error: any) => string)
    },
    options?: NotificationOptions
  ) => {
    return toast.promise(promise, {
      loading: loadingMessage,
      success: (data) => {
        const message = typeof successMessage === 'function' ? successMessage(data) : successMessage
        addToHistory({
          id: generateId(),
          type: 'success',
          title: message,
          timestamp: new Date(),
          duration: options?.duration ?? DEFAULT_DURATIONS.success,
          action: options?.action,
        })
        return message
      },
      error: (error) => {
        const message = typeof errorMessage === 'function' ? errorMessage(error) : errorMessage
        addToHistory({
          id: generateId(),
          type: 'error',
          title: message,
          timestamp: new Date(),
          duration: options?.duration ?? DEFAULT_DURATIONS.error,
          action: options?.action,
        })
        return message
      },
    })
  }, [addToHistory])

  const clearHistory = useCallback(() => {
    setHistory([])
  }, [])

  const removeFromHistory = useCallback((id: string) => {
    setHistory(prev => prev.filter(item => item.id !== id))
  }, [])

  return {
    // Core notification methods
    success,
    error,
    warning,
    info,
    loading,
    promise,
    
    // Control methods
    dismiss,
    dismissAll,
    
    // History management
    history,
    clearHistory,
    removeFromHistory,
    
    // Utility
    showNotification,
  }
}