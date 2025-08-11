# Frontend Build Issues

This document lists the remaining build issues after fixing unused imports. Issues are categorized by type for systematic resolution.

## Issue Categories

### 1. TypeScript Interface/Property Mismatch Issues

These issues occur when components try to use properties that don't exist in their type definitions:

#### WebSocket Context Issues
- `src/components/debug/websocket-debug-panel.tsx`
  - Missing `WebSocketMessage` export from `@/services/websocketService`
  - Missing `send`, `queuedMessageCount`, `optimisticUpdateCount`, `clearOptimisticUpdates` properties in WebSocketContextValue

- `src/components/examples/websocket-example.tsx`
  - Missing `useWebSocketTaskUpdates` export from `@/context/websocket-context`
  - Missing `send`, `queuedMessageCount`, `clearMessageQueue` properties in WebSocketContextValue

#### Component Property Issues
- `src/components/debug/websocket-debug-panel.tsx` - Switch component doesn't support `size` property
- `src/components/executions/execution-list.tsx` - FilterButton component doesn't support `status` property  
- `src/components/kanban/task-card.tsx` - Badge component doesn't support `title` property
- `src/pages/projects/ProjectDetail.tsx` - RealTimeProjectStats component doesn't accept `projectId` property

#### Parameter Type Issues
- `src/components/examples/websocket-example.tsx` - Parameters `task` and `changes` have implicit 'any' type
- `src/hooks/use-pull-requests.ts` - Parameter `prev` has implicit 'any' type

### 2. Routing Issues

TanStack Router type issues where routes don't exist in the route tree:

- `src/components/profile-dropdown.tsx` - Route `"/settings"` not defined
- `src/components/project-selector.tsx` - Route `"/projects/create"` not defined  
- `src/features/settings/notifications/notifications-form.tsx` - Route `"/settings"` not defined

### 3. Missing Function/Variable References

Code referencing functions or variables that don't exist:

- `src/components/executions/execution-item.tsx` - Missing `onViewLogs` function reference
- `src/hooks/use-pull-requests.ts` - Missing `useState` import (was removed but still needed)

### 4. Unused Variables/Parameters (Minor)

Variables declared but not used (less critical, could be intentional):

#### Component Props/Parameters
- `src/components/collaboration/user-presence.tsx` - `_projectId` parameter
- `src/components/debug/websocket-debug-panel.tsx` - `metrics` variable
- `src/components/examples/websocket-example.tsx` - `subscribeToProject`, `unsubscribeFromProject`
- `src/components/executions/execution-item.tsx` - `onDelete`, `onViewLogs`, `onViewDetails` parameters
- `src/components/executions/execution-list.tsx` - Multiple unused props
- `src/components/stats/real-time-project-stats.tsx` - `projectId` parameter
- Multiple other files with similar unused parameter issues

### 5. Logic/Comparison Issues

- `src/components/ui/connection-status.tsx` - Comparison between incompatible types `'"error" | "disconnected"'` and `'"connecting"'`
- `src/context/websocket-context.tsx` - Function call with wrong number of arguments

### 6. Potential Null/Undefined Issues

- `src/pages/projects/ProjectList.tsx` - Possible undefined access to `archivedProjectsData.projects.length` and `archivedProjectsData`

## Recommended Resolution Order

1. **High Priority**: Fix WebSocket context and service type definitions
2. **High Priority**: Fix missing route definitions in TanStack Router
3. **Medium Priority**: Fix component property mismatches
4. **Medium Priority**: Add missing function implementations
5. **Low Priority**: Clean up unused variables/parameters
6. **Low Priority**: Fix minor logic issues

## Notes

- Many issues appear to be related to incomplete WebSocket integration
- Some components may be incomplete or in development
- Route definitions need to be added to the TanStack Router configuration
- Consider adding proper TypeScript types for WebSocket-related functionality

Total Issues: ~85 (after unused import fixes)