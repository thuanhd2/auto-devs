import Cookies from 'js-cookie'
import { Outlet } from '@tanstack/react-router'
import { cn } from '@/lib/utils'
import { SearchProvider } from '@/context/search-context'
import { SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import SkipToMain from '@/components/skip-to-main'
import { WebSocketProvider } from '@/context/websocket-context'
import { WebSocketDebugPanel } from '@/components/debug/websocket-debug-panel'

interface Props {
  children?: React.ReactNode
}

export function AuthenticatedLayout({ children }: Props) {
  const defaultOpen = Cookies.get('sidebar_state') !== 'false'
  
  // Mock auth token - in real app this would come from auth context
  const authToken = Cookies.get('auth_token') || 'mock-token'
  
  return (
    <SearchProvider>
      <WebSocketProvider 
        authToken={authToken}
        autoConnect={true}
        onTaskCreated={(task) => {
          // Global task creation handler - can be used for notifications
          console.log('Task created via WebSocket:', task)
        }}
        onTaskUpdated={(task, changes) => {
          // Global task update handler
          console.log('Task updated via WebSocket:', task, changes)
        }}
        onTaskDeleted={(taskId) => {
          // Global task deletion handler
          console.log('Task deleted via WebSocket:', taskId)
        }}
        onProjectUpdated={(project, changes) => {
          // Global project update handler
          console.log('Project updated via WebSocket:', project, changes)
        }}
        onConnectionError={(error) => {
          console.error('WebSocket connection error:', error)
        }}
        onAuthRequired={() => {
          console.log('WebSocket authentication required')
          // In real app, redirect to login or refresh token
        }}
      >
        <SidebarProvider defaultOpen={defaultOpen}>
          <SkipToMain />
          <AppSidebar />
          <div
            id='content'
            className={cn(
              'ml-auto w-full max-w-full',
              'peer-data-[state=collapsed]:w-[calc(100%-var(--sidebar-width-icon)-1rem)]',
              'peer-data-[state=expanded]:w-[calc(100%-var(--sidebar-width))]',
              'sm:transition-[width] sm:duration-200 sm:ease-linear',
              'flex h-svh flex-col',
              'group-data-[scroll-locked=1]/body:h-full',
              'has-[main.fixed-main]:group-data-[scroll-locked=1]/body:h-svh'
            )}
          >
            {children ? children : <Outlet />}
          </div>
          {/* WebSocket Debug Panel - only in development */}
          {import.meta.env.DEV && <WebSocketDebugPanel />}
        </SidebarProvider>
      </WebSocketProvider>
    </SearchProvider>
  )
}
