import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Link } from '@tanstack/react-router'
import type { ProjectFilters } from '@/types/project'
import {
  Plus,
  Search,
  Calendar,
  GitBranch,
  Activity,
  Archive,
  RotateCcw,
} from 'lucide-react'
import { useProjects, useRestoreProject } from '@/hooks/use-projects'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ProjectCreateModal } from '@/components/project-create-modal'
import { Search as SearchComponent } from '@/components/search'
import { SimpleConfirmDialog } from '@/components/simple-confirm-dialog'
import { ThemeSwitch } from '@/components/theme-switch'

export function ProjectList() {
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState<'created_at' | 'updated_at' | 'name'>(
    'updated_at'
  )
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [activeTab, setActiveTab] = useState('active')

  const activeFilters: ProjectFilters = {
    search: search || undefined,
    sortBy,
    sortOrder,
    archived: false,
  }

  const archivedFilters: ProjectFilters = {
    search: search || undefined,
    sortBy,
    sortOrder,
    archived: true,
  }

  const {
    data: activeProjectsData,
    isLoading: activeLoading,
    error: activeError,
  } = useProjects(activeFilters)
  const {
    data: archivedProjectsData,
    isLoading: archivedLoading,
    error: archivedError,
  } = useProjects(archivedFilters)
  const restoreProjectMutation = useRestoreProject()

  const handleRestoreProject = async (projectId: string) => {
    try {
      await restoreProjectMutation.mutateAsync(projectId)
    } catch (error) {
      // Error handling is done in the mutation hook
    }
  }

  if (activeError || archivedError) {
    return (
      <div className='flex h-full items-center justify-center'>
        <div className='text-center'>
          <h3 className='text-lg font-semibold'>Error loading projects</h3>
          <p className='text-muted-foreground'>
            {activeError instanceof Error
              ? activeError.message
              : 'An unexpected error occurred'}
          </p>
        </div>
      </div>
    )
  }

  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <div className='ml-auto flex items-center space-x-4'>
          <SearchComponent />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main>
        <div className='space-y-0.5'>
          <div className='flex items-center justify-between'>
            <h1 className='text-2xl font-bold tracking-tight'>Projects</h1>
            <Button onClick={() => setCreateModalOpen(true)}>
              <Plus className='mr-2 h-4 w-4' />
              New Project
            </Button>
          </div>
          <p className='text-muted-foreground'>
            Manage your development projects and track their progress.
          </p>
        </div>
        <Separator className='my-4 lg:my-6' />
        <div className='h-full space-y-6'>
          {/* Filters */}
          <div className='flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between'>
            <div className='flex flex-1 items-center space-x-2'>
              <div className='relative max-w-sm flex-1'>
                <Search className='text-muted-foreground absolute top-2.5 left-2 h-4 w-4' />
                <Input
                  placeholder='Search projects...'
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className='pl-8'
                />
              </div>
            </div>
            <div className='flex items-center space-x-2'>
              <Select
                value={sortBy}
                onValueChange={(value: any) => setSortBy(value)}
              >
                <SelectTrigger className='w-[180px]'>
                  <SelectValue placeholder='Sort by' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='name'>Name</SelectItem>
                  <SelectItem value='created_at'>Created</SelectItem>
                  <SelectItem value='updated_at'>Updated</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={sortOrder}
                onValueChange={(value: any) => setSortOrder(value)}
              >
                <SelectTrigger className='w-[100px]'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='asc'>Asc</SelectItem>
                  <SelectItem value='desc'>Desc</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Tabs */}
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value='active'>Active Projects</TabsTrigger>
              <TabsTrigger value='archived' className='flex items-center gap-2'>
                <Archive className='h-4 w-4' />
                Archived
                {archivedProjectsData?.projects.length > 0 && (
                  <Badge variant='secondary' className='ml-1'>
                    {archivedProjectsData.projects.length}
                  </Badge>
                )}
              </TabsTrigger>
            </TabsList>

            <TabsContent value='active' className='space-y-6'>
              {/* Active Project Grid */}
              {activeLoading ? (
                <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3'>
                  {Array.from({ length: 6 }).map((_, i) => (
                    <ProjectCardSkeleton key={i} />
                  ))}
                </div>
              ) : (
                <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3'>
                  {activeProjectsData?.projects.map((project) => (
                    <ProjectCard key={project.id} project={project} />
                  ))}
                </div>
              )}

              {/* Empty State for Active Projects */}
              {!activeLoading && activeProjectsData?.projects.length === 0 && (
                <div className='flex h-[400px] items-center justify-center'>
                  <div className='text-center'>
                    <div className='bg-muted mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full'>
                      <GitBranch className='text-muted-foreground h-6 w-6' />
                    </div>
                    <h3 className='text-lg font-semibold'>No projects found</h3>
                    <p className='text-muted-foreground mb-4'>
                      {search
                        ? 'Try adjusting your search terms'
                        : 'Get started by creating your first project'}
                    </p>
                    <Button onClick={() => setCreateModalOpen(true)}>
                      <Plus className='mr-2 h-4 w-4' />
                      Create Project
                    </Button>
                  </div>
                </div>
              )}

              {/* Results Info for Active Projects */}
              {!activeLoading &&
                activeProjectsData &&
                activeProjectsData.projects.length > 0 && (
                  <div className='text-muted-foreground text-center text-sm'>
                    Showing {activeProjectsData.projects.length} of{' '}
                    {activeProjectsData.total} active projects
                  </div>
                )}
            </TabsContent>

            <TabsContent value='archived' className='space-y-6'>
              {/* Archived Project Grid */}
              {archivedLoading ? (
                <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3'>
                  {Array.from({ length: 6 }).map((_, i) => (
                    <ProjectCardSkeleton key={i} />
                  ))}
                </div>
              ) : (
                <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3'>
                  {archivedProjectsData?.projects.map((project) => (
                    <ArchivedProjectCard
                      key={project.id}
                      project={project}
                      onRestore={handleRestoreProject}
                    />
                  ))}
                </div>
              )}

              {/* Empty State for Archived Projects */}
              {!archivedLoading &&
                archivedProjectsData?.projects.length === 0 && (
                  <div className='flex h-[400px] items-center justify-center'>
                    <div className='text-center'>
                      <div className='bg-muted mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full'>
                        <Archive className='text-muted-foreground h-6 w-6' />
                      </div>
                      <h3 className='text-lg font-semibold'>
                        No archived projects
                      </h3>
                      <p className='text-muted-foreground mb-4'>
                        {search
                          ? 'Try adjusting your search terms'
                          : 'Archived projects will appear here'}
                      </p>
                    </div>
                  </div>
                )}

              {/* Results Info for Archived Projects */}
              {!archivedLoading &&
                archivedProjectsData &&
                archivedProjectsData.projects.length > 0 && (
                  <div className='text-muted-foreground text-center text-sm'>
                    Showing {archivedProjectsData.projects.length} of{' '}
                    {archivedProjectsData.total} archived projects
                  </div>
                )}
            </TabsContent>
          </Tabs>
        </div>
      </Main>

      <ProjectCreateModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
      />
    </>
  )
}

function ProjectCard({ project }: { project: any }) {
  return (
    <Link to='/projects/$projectId' params={{ projectId: project.id }}>
      <Card className='group transition-shadow hover:shadow-md'>
        <CardHeader className='pb-3'>
          <div className='flex items-start justify-between'>
            <div className='space-y-1'>
              <CardTitle className='line-clamp-1'>{project.name}</CardTitle>
              <CardDescription className='line-clamp-2'>
                {project.description || 'No description provided'}
              </CardDescription>
            </div>
            <Badge variant='secondary' className='shrink-0'>
              <Activity className='mr-1 h-3 w-3' />
              Active
            </Badge>
          </div>
        </CardHeader>

        <CardContent className='space-y-4'>
          <div className='text-muted-foreground flex items-center gap-2 text-sm'>
            <GitBranch className='h-4 w-4' />
            <span className='truncate'>
              {project.repository_url || 'No repository URL'}
            </span>
          </div>

          <div className='text-muted-foreground flex items-center justify-between text-sm'>
            <div className='flex items-center gap-2'>
              <Calendar className='h-4 w-4' />
              <span>
                {formatDistanceToNow(new Date(project.created_at), {
                  addSuffix: true,
                })}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}

function ArchivedProjectCard({
  project,
  onRestore,
}: {
  project: any
  onRestore: (projectId: string) => void
}) {
  return (
    <Card className='group border-dashed transition-shadow hover:shadow-md'>
      <CardHeader className='pb-3'>
        <div className='flex items-start justify-between'>
          <div className='space-y-1'>
            <CardTitle className='text-muted-foreground text-lg'>
              {project.name}
            </CardTitle>
            <CardDescription className='line-clamp-2'>
              {project.description || 'No description provided'}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className='pt-0'>
        <div className='text-muted-foreground flex items-center justify-between text-sm'>
          <div className='flex items-center gap-2'>
            <Calendar className='h-4 w-4' />
            <span>
              {formatDistanceToNow(new Date(project.created_at), {
                addSuffix: true,
              })}
            </span>
          </div>
          <SimpleConfirmDialog
            title='Restore Project'
            description={`Are you sure you want to restore "${project.name}"? This will make it active again.`}
            onConfirm={() => onRestore(project.id)}
            confirmText='Restore'
            cancelText='Cancel'
          >
            <Button variant='outline' size='sm'>
              <RotateCcw className='mr-2 h-4 w-4' />
              Restore
            </Button>
          </SimpleConfirmDialog>
        </div>
      </CardContent>
    </Card>
  )
}

function ProjectCardSkeleton() {
  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='space-y-2'>
          <Skeleton className='h-5 w-3/4' />
          <Skeleton className='h-4 w-full' />
        </div>
      </CardHeader>
      <CardContent className='pt-0'>
        <div className='flex items-center justify-between'>
          <Skeleton className='h-4 w-24' />
          <Skeleton className='h-4 w-4' />
        </div>
      </CardContent>
    </Card>
  )
}
