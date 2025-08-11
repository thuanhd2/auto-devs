import { useRef, useCallback, useState, useEffect } from 'react'

// Utility functions for debouncing and throttling
function debounce<T extends (...args: any[]) => any>(func: T, wait: number): T {
  let timeout: NodeJS.Timeout
  return ((...args: any[]) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }) as T
}

function throttle<T extends (...args: any[]) => any>(func: T, limit: number): T {
  let inThrottle: boolean
  return ((...args: any[]) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }) as T
}

/**
 * Hook for optimizing real-time update performance
 */
export function useRealTimePerformance() {
  const updateCountRef = useRef(0)
  const lastUpdateTimeRef = useRef(Date.now())
  const batchedUpdatesRef = useRef<Map<string, any>>(new Map())

  // Debounced function to batch multiple rapid updates
  const debouncedBatchUpdate = useCallback(
    debounce((callback: () => void) => {
      callback()
      batchedUpdatesRef.current.clear()
    }, 100),
    []
  )

  // Throttled function to limit update frequency
  const throttledUpdate = useCallback(
    throttle((callback: () => void) => {
      callback()
    }, 50),
    []
  )

  // Batch similar updates together
  const batchUpdate = useCallback(
    (key: string, data: any, callback: (batchedData: Map<string, any>) => void) => {
      batchedUpdatesRef.current.set(key, data)
      
      debouncedBatchUpdate(() => {
        callback(new Map(batchedUpdatesRef.current))
      })
    },
    [debouncedBatchUpdate]
  )

  // Track update frequency for monitoring
  const trackUpdate = useCallback(() => {
    updateCountRef.current++
    lastUpdateTimeRef.current = Date.now()
  }, [])

  // Get performance metrics
  const getMetrics = useCallback(() => {
    const now = Date.now()
    const timeSinceLastUpdate = now - lastUpdateTimeRef.current
    
    return {
      updateCount: updateCountRef.current,
      lastUpdateTime: lastUpdateTimeRef.current,
      timeSinceLastUpdate,
      updatesPerSecond: updateCountRef.current / ((now - lastUpdateTimeRef.current) / 1000),
    }
  }, [])

  return {
    batchUpdate,
    throttledUpdate,
    trackUpdate,
    getMetrics,
  }
}

/**
 * Hook for managing WebSocket message rate limiting
 */
export function useMessageRateLimit(maxMessagesPerSecond = 10) {
  const messageTimestamps = useRef<number[]>([])
  
  const canSendMessage = useCallback(() => {
    const now = Date.now()
    const oneSecondAgo = now - 1000
    
    // Remove old timestamps
    messageTimestamps.current = messageTimestamps.current.filter(
      timestamp => timestamp > oneSecondAgo
    )
    
    return messageTimestamps.current.length < maxMessagesPerSecond
  }, [maxMessagesPerSecond])
  
  const recordMessage = useCallback(() => {
    messageTimestamps.current.push(Date.now())
  }, [])
  
  return { canSendMessage, recordMessage }
}

/**
 * Hook for optimizing task list rendering with virtual scrolling considerations
 */
export function useVirtualizedTaskUpdates(tasks: any[], itemHeight = 100, containerHeight = 600) {
  const visibleItemCount = Math.ceil(containerHeight / itemHeight) + 2 // Buffer items
  
  const getVisibleRange = useCallback((scrollTop: number) => {
    const startIndex = Math.floor(scrollTop / itemHeight)
    const endIndex = Math.min(startIndex + visibleItemCount, tasks.length)
    
    return { startIndex, endIndex }
  }, [itemHeight, visibleItemCount, tasks.length])
  
  const getVisibleTasks = useCallback((scrollTop: number) => {
    const { startIndex, endIndex } = getVisibleRange(scrollTop)
    return tasks.slice(startIndex, endIndex)
  }, [tasks, getVisibleRange])
  
  return {
    getVisibleRange,
    getVisibleTasks,
    visibleItemCount,
  }
}

/**
 * Hook for debouncing search input and other rapid user interactions
 */
export function useDebouncedValue<T>(value: T, delay = 300) {
  const [debouncedValue, setDebouncedValue] = useState(value)
  
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)
    
    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])
  
  return debouncedValue
}

/**
 * Hook for optimistic update management with conflict resolution
 */
export function useOptimisticUpdateQueue() {
  const queueRef = useRef<Map<string, any>>(new Map())
  const conflictsRef = useRef<Map<string, any>>(new Map())
  
  const addOptimisticUpdate = useCallback((id: string, data: any) => {
    queueRef.current.set(id, {
      ...data,
      timestamp: Date.now(),
      pending: true,
    })
  }, [])
  
  const confirmUpdate = useCallback((id: string, serverData?: any) => {
    const pending = queueRef.current.get(id)
    if (pending) {
      queueRef.current.delete(id)
      
      // Check for conflicts
      if (serverData && hasConflict(pending, serverData)) {
        conflictsRef.current.set(id, {
          local: pending,
          server: serverData,
          timestamp: Date.now(),
        })
      }
    }
  }, [])
  
  const revertUpdate = useCallback((id: string) => {
    queueRef.current.delete(id)
    conflictsRef.current.delete(id)
  }, [])
  
  const getConflicts = useCallback(() => {
    return Array.from(conflictsRef.current.entries())
  }, [])
  
  const resolveConflict = useCallback((id: string, resolution: 'local' | 'server') => {
    conflictsRef.current.delete(id)
  }, [])
  
  return {
    addOptimisticUpdate,
    confirmUpdate,
    revertUpdate,
    getConflicts,
    resolveConflict,
    pendingUpdates: queueRef.current,
  }
}

/**
 * Utility function to detect conflicts between local and server data
 */
function hasConflict(localData: any, serverData: any): boolean {
  if (!localData || !serverData) return false
  
  // Compare timestamps
  const localTime = new Date(localData.updated_at || localData.timestamp)
  const serverTime = new Date(serverData.updated_at)
  
  // If server data is significantly newer, there might be a conflict
  return Math.abs(serverTime.getTime() - localTime.getTime()) > 5000 // 5 seconds
}

/**
 * Hook for managing connection quality and adapting update strategies
 */
export function useConnectionQuality() {
  const [quality, setQuality] = useState<'excellent' | 'good' | 'poor' | 'offline'>('good')
  const latencyRef = useRef<number[]>([])
  
  const measureLatency = useCallback((startTime: number) => {
    const latency = Date.now() - startTime
    latencyRef.current.push(latency)
    
    // Keep only last 10 measurements
    if (latencyRef.current.length > 10) {
      latencyRef.current.shift()
    }
    
    // Calculate average latency
    const avgLatency = latencyRef.current.reduce((a, b) => a + b, 0) / latencyRef.current.length
    
    // Update quality based on latency
    if (avgLatency < 100) {
      setQuality('excellent')
    } else if (avgLatency < 300) {
      setQuality('good')
    } else if (avgLatency < 1000) {
      setQuality('poor')
    } else {
      setQuality('offline')
    }
    
    return avgLatency
  }, [])
  
  const getUpdateStrategy = useCallback(() => {
    switch (quality) {
      case 'excellent':
        return { batchDelay: 50, throttleMs: 16, enableAnimations: true }
      case 'good':
        return { batchDelay: 100, throttleMs: 33, enableAnimations: true }
      case 'poor':
        return { batchDelay: 300, throttleMs: 100, enableAnimations: false }
      case 'offline':
        return { batchDelay: 1000, throttleMs: 500, enableAnimations: false }
    }
  }, [quality])
  
  return {
    quality,
    measureLatency,
    getUpdateStrategy,
    averageLatency: latencyRef.current.length > 0 
      ? latencyRef.current.reduce((a, b) => a + b, 0) / latencyRef.current.length 
      : 0,
  }
}