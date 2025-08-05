import { formatDistanceToNow } from 'date-fns'
import { Link, useParams, useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Settings,
  GitBranch,
  Calendar,
  BarChart3,
  CheckCircle2,
  Clock,
  AlertCircle,
  XCircle,
  Trash2,
} from 'lucide-react'
import { useProject, useProjectStatistics, useDeleteProject } from '@/hooks/use-projects'
import { useTasks } from '@/hooks/use-tasks'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  UserPresence,
  UserPresenceCompact,
} from '@/components/collaboration/user-presence'
import { GitStatusCard } from '@/components/kanban/git-status-card'
import { ProjectBoard } from '@/components/kanban/project-board'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { RealTimeNotifications } from '@/components/notifications/real-time-notifications'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search as SearchComponent } from '@/components/search'
import {
  RealTimeProjectStats,
  CompactProjectStats,
} from '@/components/stats/real-time-project-stats'
import { ThemeSwitch } from '@/components/theme-switch'
import { SimpleConfirmDialog } from '@/components/simple-confirm-dialog'

const statusConfig = {
  TODO: { label: 'To Do', icon: Clock, color: 'bg-slate-500' },
  PLANNING: { label: 'Planning', icon: BarChart3, color: 'bg-blue-500' },
  PLAN_REVIEWING: {
    label: 'Plan Review',
    icon: AlertCircle,
    color: 'bg-yellow-500',
  },
  IMPLEMENTING: { label: 'Implementing', icon: Clock, color: 'bg-orange-500' },
  CODE_REVIEWING: {
    label: 'Code Review',
    icon: AlertCircle,
    color: 'bg-purple-500',
  },
  DONE: { label: 'Done', icon: CheckCircle2, color: 'bg-green-500' },
  CANCELLED: { label: 'Cancelled', icon: XCircle, color: 'bg-red-500' },
}

export function ProjectDetail() {
  const { projectId } = useParams({
    from: '/_authenticated/projects/$projectId',
  })
  const navigate = useNavigate()
  const {
    data: project,
    isLoading: projectLoading,
    error: projectError,
  } = useProject(projectId)
  const { data: statistics } = useProjectStatistics(projectId)
  const { data: tasksResponse } = useTasks(projectId, {})
  const deleteProjectMutation = useDeleteProject()

  const handleDeleteProject = async () => {
    try {
      await deleteProjectMutation.mutateAsync(projectId)
      navigate({ to: '/projects' })
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  if (projectError) {
    return (
      <div className='flex h-full items-center justify-center'>
        <div className='text-center'>
          <h3 className='text-lg font-semibold'>Error loading project</h3>
          <p className='text-muted-foreground mb-4'>
            {projectError instanceof Error
              ? projectError.message
              : 'An unexpected error occurred'}
          </p>
          <Link to='/projects'>
            <Button variant='outline'>
              <ArrowLeft className='mr-2 h-4 w-4' />
              Back to Projects
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  if (projectLoading) {
    return <ProjectDetailSkeleton />
  }

  if (!project) {
    return (
      <div className='flex h-full items-center justify-center'>
        <div className='text-center'>
          <h3 className='text-lg font-semibold'>Project not found</h3>
          <p className='text-muted-foreground mb-4'>
            The project you're looking for doesn't exist or has been deleted.
          </p>
          <Link to='/projects'>
            <Button variant='outline'>
              <ArrowLeft className='mr-2 h-4 w-4' />
              Back to Projects
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  const tasks = tasksResponse?.tasks || []
  const totalTasks = statistics?.total_tasks || 0
  const completedTasks = statistics?.tasks_by_status.DONE || 0
  const progress = totalTasks > 0 ? (completedTasks / totalTasks) * 100 : 0

  return (
    <>
      {/* Real-time notifications */}
      <RealTimeNotifications
        projectId={projectId}
        enableToastNotifications={true}
        enableBrowserNotifications={false}
        enableSound={false}
      />
      {/* ===== Top Heading ===== */}
      <Main>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-4'>
            <Link to='/projects'>
              <Button variant='ghost' size='icon'>
                <ArrowLeft className='h-4 w-4' />
              </Button>
            </Link>
            <div>
              <div className='flex items-center gap-3'>
                <h1 className='text-3xl font-bold'>{project.name}</h1>
                <UserPresence
                  projectId={projectId}
                  showDetails={false}
                  maxAvatars={3}
                />
              </div>
              <p className='text-muted-foreground'>
                {project.description || 'No description provided'}
              </p>
            </div>
          </div>
          <div className='flex items-center gap-2'>
            <UserPresenceCompact projectId={projectId} />
            <Link to='/projects/$projectId/edit' params={{ projectId }}>
              <Button variant='outline'>
                <Settings className='mr-2 h-4 w-4' />
                Settings
              </Button>
            </Link>
            <SimpleConfirmDialog
              title="Delete Project"
              description={`Are you sure you want to delete "${project?.name}"? This action cannot be undone. The project will be moved to the trash and can be restored later.`}
              onConfirm={handleDeleteProject}
              destructive={true}
              confirmText="Delete Project"
              cancelText="Cancel"
            >
              <Button variant='destructive'>
                <Trash2 className='mr-2 h-4 w-4' />
                Delete
              </Button>
            </SimpleConfirmDialog>
          </div>
        </div>
        <Separator className='my-4 lg:my-6' />
        <div className='h-full space-y-6'>
          {/* Header */}

          <Tabs defaultValue='overview' className='h-full'>
            <TabsList>
              <TabsTrigger value='overview'>Overview</TabsTrigger>
              <TabsTrigger value='tasks'>Tasks</TabsTrigger>
            </TabsList>

            <TabsContent value='overview' className='space-y-6'>
              {/* Real-time Project Statistics */}
              <RealTimeProjectStats
                projectId={projectId}
                tasks={tasks}
                className='mb-6'
              />

              {/* Project Info */}
              <div className='grid gap-6 md:grid-cols-2'>
                <Card>
                  <CardHeader>
                    <CardTitle>Project Information</CardTitle>
                  </CardHeader>
                  <CardContent className='space-y-4'>
                    {project.repository_url && (
                      <div className='flex items-center gap-2 text-sm'>
                        <GitBranch className='text-muted-foreground h-4 w-4' />
                        <span className='text-muted-foreground'>
                          Repository:
                        </span>
                        <span className='truncate font-mono'>
                          {project.repository_url}
                        </span>
                      </div>
                    )}

                    <div className='flex items-center gap-2 text-sm'>
                      <Calendar className='text-muted-foreground h-4 w-4' />
                      <span className='text-muted-foreground'>Created:</span>
                      <span>
                        {formatDistanceToNow(new Date(project.created_at), {
                          addSuffix: true,
                        })}
                      </span>
                    </div>
                    <div className='flex items-center gap-2 text-sm'>
                      <Calendar className='text-muted-foreground h-4 w-4' />
                      <span className='text-muted-foreground'>Updated:</span>
                      <span>
                        {formatDistanceToNow(new Date(project.updated_at), {
                          addSuffix: true,
                        })}
                      </span>
                    </div>
                  </CardContent>
                </Card>

                {/* Git Status Card */}
                <GitStatusCard
                  projectId={projectId}
                  gitEnabled={!!project.repository_url}
                />

                <Card>
                  <CardHeader>
                    <CardTitle>Real-time Activity</CardTitle>
                    <CardDescription>
                      Live project activity and collaboration
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='space-y-4'>
                    <CompactProjectStats projectId={projectId} tasks={tasks} />

                    <div className='border-t pt-2'>
                      <UserPresence
                        projectId={projectId}
                        showDetails={true}
                        maxAvatars={5}
                      />
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Fallback to original statistics if real-time data is not available */}
              {!tasks.length && statistics && (
                <Card>
                  <CardHeader>
                    <CardTitle>Task Distribution</CardTitle>
                    <CardDescription>
                      Overview of tasks by status
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className='grid gap-4 md:grid-cols-4 lg:grid-cols-7'>
                      {Object.entries(statistics.tasks_by_status).map(
                        ([status, count]) => {
                          const config =
                            statusConfig[status as keyof typeof statusConfig]
                          const Icon = config.icon

                          return (
                            <div key={status} className='space-y-2 text-center'>
                              <div
                                className={`mx-auto h-8 w-8 rounded-full ${config.color} flex items-center justify-center`}
                              >
                                <Icon className='h-4 w-4 text-white' />
                              </div>
                              <div className='text-2xl font-bold'>{count}</div>
                              <div className='text-muted-foreground text-xs'>
                                {config.label}
                              </div>
                            </div>
                          )
                        }
                      )}
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            <TabsContent value='tasks' className='h-full'>
              <ProjectBoard projectId={projectId} />
            </TabsContent>
          </Tabs>
        </div>
      </Main>
    </>
  )
}

function ProjectDetailSkeleton() {
  return (
    <div className='h-full space-y-6'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-4'>
          <Skeleton className='h-10 w-10' />
          <div className='space-y-2'>
            <Skeleton className='h-8 w-64' />
            <Skeleton className='h-4 w-96' />
          </div>
        </div>
        <Skeleton className='h-10 w-24' />
      </div>

      <div className='grid gap-6 md:grid-cols-2'>
        <Card>
          <CardHeader>
            <Skeleton className='h-6 w-40' />
          </CardHeader>
          <CardContent className='space-y-4'>
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className='flex items-center gap-2'>
                <Skeleton className='h-4 w-4' />
                <Skeleton className='h-4 w-20' />
                <Skeleton className='h-4 flex-1' />
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <Skeleton className='h-6 w-32' />
            <Skeleton className='h-4 w-48' />
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='space-y-2'>
              <div className='flex justify-between'>
                <Skeleton className='h-4 w-20' />
                <Skeleton className='h-4 w-12' />
              </div>
              <Skeleton className='h-2 w-full' />
            </div>
            <Skeleton className='h-4 w-64' />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
