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
  GitFork,
} from 'lucide-react'
import { useProjects } from '@/hooks/use-projects'
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
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ProjectCreateModal } from '@/components/project-create-modal'
import { Search as SearchComponent } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

export function ProjectList() {
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState<'created_at' | 'updated_at' | 'name'>(
    'updated_at'
  )
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [createModalOpen, setCreateModalOpen] = useState(false)

  const filters: ProjectFilters = {
    search: search || undefined,
    sortBy,
    sortOrder,
  }

  const { data: projectsData, isLoading, error } = useProjects(filters)

  if (error) {
    return (
      <div className='flex h-full items-center justify-center'>
        <div className='text-center'>
          <h3 className='text-lg font-semibold'>Error loading projects</h3>
          <p className='text-muted-foreground'>
            {error instanceof Error
              ? error.message
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
            <div className='relative flex-1 sm:max-w-sm'>
              <Search className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2' />
              <Input
                placeholder='Search projects...'
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className='pl-9'
              />
            </div>

            <div className='flex gap-2'>
              <Select
                value={sortBy}
                onValueChange={(value: any) => setSortBy(value)}
              >
                <SelectTrigger className='w-[150px]'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='updated_at'>Last Updated</SelectItem>
                  <SelectItem value='created_at'>Created Date</SelectItem>
                  <SelectItem value='name'>Name</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={sortOrder}
                onValueChange={(value: any) => setSortOrder(value)}
              >
                <SelectTrigger className='w-[120px]'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='desc'>Descending</SelectItem>
                  <SelectItem value='asc'>Ascending</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Project Grid */}
          {isLoading ? (
            <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3'>
              {Array.from({ length: 6 }).map((_, i) => (
                <ProjectCardSkeleton key={i} />
              ))}
            </div>
          ) : (
            <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3'>
              {projectsData?.projects.map((project) => (
                <ProjectCard key={project.id} project={project} />
              ))}
            </div>
          )}

          {/* Empty State */}
          {!isLoading && projectsData?.projects.length === 0 && (
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

          {/* Results Info */}
          {!isLoading && projectsData && projectsData.projects.length > 0 && (
            <div className='text-muted-foreground text-center text-sm'>
              Showing {projectsData.projects.length} of {projectsData.total}{' '}
              projects
            </div>
          )}
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
      <Card className='hover:bg-muted/50 h-full cursor-pointer transition-colors'>
        <CardHeader>
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
          {project.repository_url && (
            <div className='text-muted-foreground flex items-center gap-2 text-sm'>
              <GitBranch className='h-4 w-4' />
              <span className='truncate'>{project.repository_url}</span>
            </div>
          )}

          <div className='text-muted-foreground flex items-center gap-2 text-sm'>
            <Calendar className='h-4 w-4' />
            <span>
              Updated{' '}
              {formatDistanceToNow(new Date(project.updated_at), {
                addSuffix: true,
              })}
            </span>
          </div>

          <div className='flex gap-2'>
            <Badge variant='outline'>0 tasks</Badge>
            <Badge variant='outline'>0 active</Badge>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}

function ProjectCardSkeleton() {
  return (
    <Card className='h-full'>
      <CardHeader>
        <div className='flex items-start justify-between'>
          <div className='flex-1 space-y-2'>
            <Skeleton className='h-5 w-3/4' />
            <Skeleton className='h-4 w-full' />
            <Skeleton className='h-4 w-2/3' />
          </div>
          <Skeleton className='h-6 w-16' />
        </div>
      </CardHeader>

      <CardContent className='space-y-4'>
        <div className='flex items-center gap-2'>
          <Skeleton className='h-4 w-4' />
          <Skeleton className='h-4 flex-1' />
        </div>

        <div className='flex items-center gap-2'>
          <Skeleton className='h-4 w-4' />
          <Skeleton className='h-4 w-24' />
        </div>

        <div className='flex gap-2'>
          <Skeleton className='h-6 w-16' />
          <Skeleton className='h-6 w-16' />
        </div>
      </CardContent>
    </Card>
  )
}
