import { useState, useEffect, useMemo } from 'react'
import type { Task, TaskStatus } from '@/types/task'
import {
  CheckCircle,
  Clock,
  Users,
  AlertCircle,
  TrendingUp,
  Activity,
} from 'lucide-react'
import { AnimationUtils } from '@/utils/animations'
import { useWebSocketContext } from '@/context/websocket-context'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'

interface ProjectStatsProps {
  projectId: string
  tasks: Task[]
  className?: string
}

interface TaskStatistics {
  total: number
  byStatus: Record<TaskStatus, number>
  completionRate: number
  todayCreated: number
  todayCompleted: number
  overdue: number
  activeUsers: number
}

/**
 * Real-time project statistics component with live updates and animations
 */
export function RealTimeProjectStats({
  projectId,
  tasks,
  className = '',
}: ProjectStatsProps) {
  const [previousStats, setPreviousStats] = useState<TaskStatistics | null>(
    null
  )
  const [animatingStats, setAnimatingStats] = useState<Set<string>>(new Set())
  const { onlineUsers } = useWebSocketContext()

  // Calculate current statistics
  const stats = useMemo(() => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)

    const byStatus: Record<TaskStatus, number> = {
      TODO: 0,
      PLANNING: 0,
      PLAN_REVIEWING: 0,
      IMPLEMENTING: 0,
      CODE_REVIEWING: 0,
      DONE: 0,
      CANCELLED: 0,
    }

    let todayCreated = 0
    let todayCompleted = 0
    let overdue = 0

    tasks.forEach((task) => {
      byStatus[task.status]++

      const createdDate = new Date(task.created_at)
      if (createdDate >= today) {
        todayCreated++
      }

      if (task.status === 'DONE') {
        const completedDate = new Date(task.updated_at)
        if (completedDate >= today) {
          todayCompleted++
        }
      }

      // Check if task is overdue (simplified logic)
      if (
        (task as any).due_date &&
        new Date((task as any).due_date) < new Date() &&
        task.status !== 'DONE'
      ) {
        overdue++
      }
    })

    const completedTasks = byStatus.DONE
    const totalTasks = tasks.length
    const completionRate =
      totalTasks > 0 ? (completedTasks / totalTasks) * 100 : 0

    return {
      total: totalTasks,
      byStatus,
      completionRate,
      todayCreated,
      todayCompleted,
      overdue,
      activeUsers: onlineUsers.size,
    }
  }, [tasks, onlineUsers])

  // Animate stat changes
  useEffect(() => {
    if (previousStats) {
      const changedStats = new Set<string>()

      // Check which stats have changed
      if (stats.total !== previousStats.total) changedStats.add('total')
      if (stats.completionRate !== previousStats.completionRate)
        changedStats.add('completion')
      if (stats.todayCreated !== previousStats.todayCreated)
        changedStats.add('todayCreated')
      if (stats.todayCompleted !== previousStats.todayCompleted)
        changedStats.add('todayCompleted')
      if (stats.overdue !== previousStats.overdue) changedStats.add('overdue')
      if (stats.activeUsers !== previousStats.activeUsers)
        changedStats.add('activeUsers')

      Object.keys(stats.byStatus).forEach((status) => {
        if (
          stats.byStatus[status as TaskStatus] !==
          previousStats.byStatus[status as TaskStatus]
        ) {
          changedStats.add(`status-${status}`)
        }
      })

      setAnimatingStats(changedStats)

      // Animate stat elements
      changedStats.forEach((statKey) => {
        const element = document.querySelector(
          `[data-stat="${statKey}"]`
        ) as HTMLElement
        if (element) {
          AnimationUtils.animateCountUpdate(element)
        }
      })

      // Clear animations after delay
      setTimeout(() => setAnimatingStats(new Set()), 2000)
    }

    setPreviousStats(stats)
  }, [stats, previousStats])

  const getStatusColor = (status: TaskStatus): string => {
    const colors = {
      TODO: 'text-gray-600',
      PLANNING: 'text-blue-600',
      PLAN_REVIEWING: 'text-yellow-600',
      IMPLEMENTING: 'text-purple-600',
      CODE_REVIEWING: 'text-orange-600',
      DONE: 'text-green-600',
      CANCELLED: 'text-red-600',
    }
    return colors[status] || 'text-gray-600'
  }

  const getStatusIcon = (status: TaskStatus) => {
    switch (status) {
      case 'DONE':
        return <CheckCircle className='h-4 w-4' />
      case 'TODO':
        return <Clock className='h-4 w-4' />
      default:
        return <Activity className='h-4 w-4' />
    }
  }

  return (
    <div
      className={`grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 ${className}`}
    >
      {/* Total Tasks */}
      <Card className='relative overflow-hidden'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Total Tasks</CardTitle>
          <Activity className='text-muted-foreground h-4 w-4' />
        </CardHeader>
        <CardContent>
          <div
            className={`text-2xl font-bold transition-all duration-300 ${
              animatingStats.has('total') ? 'scale-110 text-blue-600' : ''
            }`}
            data-stat='total'
          >
            {stats.total}
          </div>
          <div className='text-muted-foreground flex items-center gap-2 text-xs'>
            <TrendingUp className='h-3 w-3' />
            <span>+{stats.todayCreated} today</span>
          </div>
        </CardContent>
      </Card>

      {/* Completion Rate */}
      <Card className='relative overflow-hidden'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Completion Rate</CardTitle>
          <CheckCircle className='text-muted-foreground h-4 w-4' />
        </CardHeader>
        <CardContent>
          <div
            className={`text-2xl font-bold transition-all duration-300 ${
              animatingStats.has('completion') ? 'scale-110 text-green-600' : ''
            }`}
            data-stat='completion'
          >
            {stats.completionRate.toFixed(1)}%
          </div>
          <Progress
            value={stats.completionRate}
            className='mt-2 h-2 transition-all duration-500'
          />
          <div className='text-muted-foreground mt-1 text-xs'>
            {stats.byStatus.DONE} of {stats.total} completed
          </div>
        </CardContent>
      </Card>

      {/* Today's Activity */}
      <Card className='relative overflow-hidden'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>
            Today's Activity
          </CardTitle>
          <Clock className='text-muted-foreground h-4 w-4' />
        </CardHeader>
        <CardContent>
          <div className='space-y-2'>
            <div className='flex items-center justify-between'>
              <span className='text-muted-foreground text-sm'>Created</span>
              <span
                className={`font-medium transition-all duration-300 ${
                  animatingStats.has('todayCreated')
                    ? 'scale-110 text-blue-600'
                    : ''
                }`}
                data-stat='todayCreated'
              >
                {stats.todayCreated}
              </span>
            </div>
            <div className='flex items-center justify-between'>
              <span className='text-muted-foreground text-sm'>Completed</span>
              <span
                className={`font-medium transition-all duration-300 ${
                  animatingStats.has('todayCompleted')
                    ? 'scale-110 text-green-600'
                    : ''
                }`}
                data-stat='todayCompleted'
              >
                {stats.todayCompleted}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Alerts & Active Users */}
      <Card className='relative overflow-hidden'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Alerts & Users</CardTitle>
          <Users className='text-muted-foreground h-4 w-4' />
        </CardHeader>
        <CardContent>
          <div className='space-y-2'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-1'>
                <AlertCircle className='h-3 w-3 text-red-500' />
                <span className='text-muted-foreground text-sm'>Overdue</span>
              </div>
              <span
                className={`font-medium transition-all duration-300 ${
                  animatingStats.has('overdue') ? 'scale-110 text-red-600' : ''
                } ${stats.overdue > 0 ? 'text-red-600' : ''}`}
                data-stat='overdue'
              >
                {stats.overdue}
              </span>
            </div>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-1'>
                <Users className='h-3 w-3 text-green-500' />
                <span className='text-muted-foreground text-sm'>Active</span>
              </div>
              <span
                className={`font-medium transition-all duration-300 ${
                  animatingStats.has('activeUsers')
                    ? 'scale-110 text-green-600'
                    : ''
                }`}
                data-stat='activeUsers'
              >
                {stats.activeUsers}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Status Breakdown */}
      <Card className='col-span-full'>
        <CardHeader>
          <CardTitle className='text-lg'>Task Status Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='grid grid-cols-2 gap-4 md:grid-cols-4 lg:grid-cols-7'>
            {Object.entries(stats.byStatus).map(([status, count]) => (
              <div key={status} className='text-center'>
                <div className='mb-2 flex items-center justify-center'>
                  <div className={getStatusColor(status as TaskStatus)}>
                    {getStatusIcon(status as TaskStatus)}
                  </div>
                </div>
                <div
                  className={`text-2xl font-bold transition-all duration-300 ${
                    animatingStats.has(`status-${status}`) ? 'scale-110' : ''
                  } ${getStatusColor(status as TaskStatus)}`}
                  data-stat={`status-${status}`}
                >
                  {count}
                </div>
                <div className='text-muted-foreground text-xs'>
                  {status.replace('_', ' ')}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

/**
 * Compact stats component for sidebar or smaller spaces
 */
export function CompactProjectStats({
  tasks,
}: Pick<ProjectStatsProps, 'tasks'>) {
  const stats = useMemo(() => {
    const completed = tasks.filter((t) => t.status === 'DONE').length
    const total = tasks.length
    const completionRate = total > 0 ? (completed / total) * 100 : 0

    return {
      total,
      completed,
      completionRate,
      inProgress: tasks.filter((t) =>
        ['IMPLEMENTING', 'CODE_REVIEWING'].includes(t.status)
      ).length,
    }
  }, [tasks])

  return (
    <div className='space-y-2'>
      <div className='flex items-center justify-between text-sm'>
        <span className='text-muted-foreground'>Progress</span>
        <span className='font-medium'>{stats.completionRate.toFixed(0)}%</span>
      </div>
      <Progress value={stats.completionRate} className='h-2' />
      <div className='grid grid-cols-3 gap-2 text-xs'>
        <div className='text-center'>
          <div className='font-medium'>{stats.total}</div>
          <div className='text-muted-foreground'>Total</div>
        </div>
        <div className='text-center'>
          <div className='font-medium text-green-600'>{stats.completed}</div>
          <div className='text-muted-foreground'>Done</div>
        </div>
        <div className='text-center'>
          <div className='font-medium text-blue-600'>{stats.inProgress}</div>
          <div className='text-muted-foreground'>Active</div>
        </div>
      </div>
    </div>
  )
}
