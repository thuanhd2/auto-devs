import { useState } from 'react'
import type { Task } from '@/types/task'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { PlanReview } from './plan-review'
import { PlanEditor } from './plan-editor'
import { PlanPreview } from './plan-preview'

// Sample task data for demonstration
const sampleTask: Task = {
  id: 'demo-task-1',
  project_id: 'demo-project',
  title: 'Implement User Authentication System',
  description: 'Create a comprehensive user authentication system with login, registration, password reset, and session management.',
  status: 'PLAN_REVIEWING',
  plan: `# Implementation Plan: User Authentication System

## Overview
This plan outlines the implementation of a secure user authentication system for the application.

## Phase 1: Backend Infrastructure
### Database Schema
- Create \`users\` table with fields:
  - \`id\` (UUID, primary key)
  - \`email\` (unique, not null)
  - \`password_hash\` (bcrypt hashed)
  - \`created_at\`, \`updated_at\`
  - \`email_verified\` (boolean)
  - \`last_login\`

### Authentication Endpoints
- **POST /api/auth/register** - User registration
- **POST /api/auth/login** - User login
- **POST /api/auth/logout** - User logout
- **POST /api/auth/refresh** - Token refresh
- **POST /api/auth/forgot-password** - Password reset request
- **POST /api/auth/reset-password** - Password reset confirmation

## Phase 2: Security Implementation
### Password Security
- Use \`bcrypt\` with salt rounds >= 12
- Implement password strength validation
- Rate limiting on authentication endpoints

### JWT Token Management
- Access tokens (15 min expiry)
- Refresh tokens (7 days expiry)
- Token blacklisting for logout
- Secure httpOnly cookies for token storage

### Session Management
- Redis-based session store
- Automatic session cleanup
- Cross-device session management

## Phase 3: Frontend Integration
### Login/Registration Forms
- Form validation with proper error handling
- Social login integration (Google, GitHub)
- Remember me functionality
- Password strength indicator

### Authentication State Management
- Context-based auth state
- Automatic token refresh
- Route protection
- Persistent login state

### Security Features
- CSRF protection
- Rate limiting display
- Account lockout notifications
- Email verification flow

## Phase 4: Testing & Security
### Unit Tests
- Authentication middleware tests
- Token validation tests
- Password hashing tests

### Integration Tests
- Complete authentication flow
- Password reset flow
- Session management

### Security Testing
- Penetration testing
- SQL injection prevention
- XSS protection validation

## Estimated Timeline: 3-4 weeks
- Phase 1: 1 week
- Phase 2: 1 week  
- Phase 3: 1 week
- Phase 4: 1 week

## Dependencies
- \`bcryptjs\` - Password hashing
- \`jsonwebtoken\` - JWT implementation
- \`express-rate-limit\` - Rate limiting
- \`nodemailer\` - Email service
- \`redis\` - Session storage`,
  branch_name: 'feature/user-authentication',
  pr_url: '',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

const sampleEmptyTask: Task = {
  id: 'demo-task-empty',
  project_id: 'demo-project',
  title: 'Task Without Plan',
  description: 'This task has no implementation plan yet.',
  status: 'PLANNING',
  plan: '',
  branch_name: '',
  pr_url: '',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

const sampleShortPlan: Task = {
  id: 'demo-task-short',
  project_id: 'demo-project',
  title: 'Fix Navigation Bug',
  description: 'Quick fix for the navigation menu bug.',
  status: 'PLAN_REVIEWING',
  plan: `# Quick Fix: Navigation Bug

## Problem
The navigation menu is not collapsing on mobile devices.

## Solution
1. Update CSS media queries for mobile breakpoints
2. Fix JavaScript event handlers for menu toggle
3. Test across different screen sizes

## Files to modify:
- \`src/components/Navigation.css\`
- \`src/components/Navigation.js\`

**Estimated time:** 2 hours`,
  branch_name: 'fix/navigation-mobile',
  pr_url: '',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

export function PlanReviewDemo() {
  const [selectedTask, setSelectedTask] = useState<Task>(sampleTask)
  const [tasks] = useState<Task[]>([sampleTask, sampleEmptyTask, sampleShortPlan])

  const handleTaskUpdate = (updatedTask: Task) => {
    console.log('Task updated:', updatedTask)
    setSelectedTask(updatedTask)
  }

  const handleStatusChange = (taskId: string, newStatus: Task['status']) => {
    console.log('Status changed:', taskId, newStatus)
    setSelectedTask(prev => ({ ...prev, status: newStatus }))
  }

  return (
    <div className="w-full max-w-7xl mx-auto p-6 space-y-8">
      <div>
        <h1 className="text-3xl font-bold mb-2">Plan Review Interface Demo</h1>
        <p className="text-gray-600 mb-6">
          Demonstration of the plan review, editing, and preview components.
        </p>
      </div>

      {/* Task Selector */}
      <Card>
        <CardHeader>
          <CardTitle>Select Demo Task</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {tasks.map((task) => (
              <div
                key={task.id}
                onClick={() => setSelectedTask(task)}
                className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                  selectedTask.id === task.id 
                    ? 'border-blue-500 bg-blue-50' 
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <h3 className="font-medium mb-2">{task.title}</h3>
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="outline" className="text-xs">
                    {task.status.replace('_', ' ').toLowerCase()}
                  </Badge>
                </div>
                <p className="text-sm text-gray-500 line-clamp-2">
                  {task.description}
                </p>
                <div className="mt-2 text-xs text-gray-400">
                  Plan: {task.plan ? `${task.plan.length} characters` : 'No plan'}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Demo Components */}
      <Tabs defaultValue="full-review" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="full-review">Full Review</TabsTrigger>
          <TabsTrigger value="editor-only">Editor Only</TabsTrigger>
          <TabsTrigger value="preview-only">Preview Only</TabsTrigger>
          <TabsTrigger value="responsive">Responsive</TabsTrigger>
        </TabsList>

        <TabsContent value="full-review" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Complete Plan Review Interface</CardTitle>
              <p className="text-sm text-gray-600">
                Full interface with plan display, editing, approval/rejection controls.
              </p>
            </CardHeader>
            <CardContent>
              <PlanReview
                task={selectedTask}
                onPlanUpdate={handleTaskUpdate}
                onStatusChange={handleStatusChange}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="editor-only" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Plan Editor Component</CardTitle>
              <p className="text-sm text-gray-600">
                Standalone editor with auto-save, history, and export features.
              </p>
            </CardHeader>
            <CardContent className="p-0">
              <PlanEditor
                initialValue={selectedTask.plan}
                onSave={async (content) => {
                  console.log('Saving plan:', content)
                  handleTaskUpdate({ ...selectedTask, plan: content })
                }}
                onCancel={() => console.log('Edit cancelled')}
                autoSave={true}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="preview-only" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Plan Preview Component</CardTitle>
              <p className="text-sm text-gray-600">
                Markdown preview with copy, print, and export functionality.
              </p>
            </CardHeader>
            <CardContent>
              <PlanPreview
                content={selectedTask.plan}
                showActions={true}
                printFriendly={false}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="responsive" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Responsive Design Test</CardTitle>
              <p className="text-sm text-gray-600">
                Test the components at different screen sizes.
              </p>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Mobile simulation */}
                <div className="border rounded-lg p-4">
                  <h3 className="font-medium mb-3 text-center">Mobile View (375px)</h3>
                  <div className="w-[375px] mx-auto border rounded-lg overflow-hidden">
                    <div className="scale-75 origin-top">
                      <div className="w-[500px] h-[600px] overflow-auto">
                        <PlanReview
                          task={selectedTask}
                          onPlanUpdate={handleTaskUpdate}
                          onStatusChange={handleStatusChange}
                        />
                      </div>
                    </div>
                  </div>
                </div>

                {/* Tablet simulation */}
                <div className="border rounded-lg p-4">
                  <h3 className="font-medium mb-3 text-center">Tablet View (768px)</h3>
                  <div className="w-full max-w-[600px] mx-auto border rounded-lg overflow-hidden">
                    <div className="scale-90 origin-top">
                      <div className="w-[800px] h-[600px] overflow-auto">
                        <PlanReview
                          task={selectedTask}
                          onPlanUpdate={handleTaskUpdate}
                          onStatusChange={handleStatusChange}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}