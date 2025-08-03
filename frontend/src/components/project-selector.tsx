import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Check, ChevronsUpDown, Plus, Folder, Search } from 'lucide-react'
import { useProjects } from '@/hooks/use-projects'
import { Button } from '@/components/ui/button'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Project } from '@/types/project'

interface ProjectSelectorProps {
  currentProjectId?: string
  className?: string
}

export function ProjectSelector({ currentProjectId, className }: ProjectSelectorProps) {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const { data: projectsData, isLoading } = useProjects()

  const currentProject = projectsData?.projects.find(p => p.id === currentProjectId)

  const onSelect = (project: Project) => {
    setOpen(false)
    navigate({ to: '/projects/$projectId', params: { projectId: project.id } })
  }

  const onCreateNew = () => {
    setOpen(false)
    navigate({ to: '/projects/create' })
  }

  const onViewAll = () => {
    setOpen(false)
    navigate({ to: '/projects' })
  }

  if (isLoading) {
    return <ProjectSelectorSkeleton className={className} />
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn("justify-between", className)}
        >
          <div className="flex items-center gap-2 min-w-0">
            <Folder className="h-4 w-4 shrink-0" />
            <span className="truncate">
              {currentProject?.name || "Select project..."}
            </span>
          </div>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      
      <PopoverContent className="w-[300px] p-0">
        <Command>
          <CommandInput placeholder="Search projects..." />
          <CommandList>
            <CommandEmpty>No projects found.</CommandEmpty>
            
            <CommandGroup heading="Recent Projects">
              {projectsData?.projects.slice(0, 5).map((project) => (
                <CommandItem
                  key={project.id}
                  value={project.id}
                  onSelect={() => onSelect(project)}
                  className="cursor-pointer"
                >
                  <div className="flex items-center gap-2 min-w-0 flex-1">
                    <Folder className="h-4 w-4 shrink-0" />
                    <div className="min-w-0 flex-1">
                      <div className="truncate font-medium">{project.name}</div>
                      {project.description && (
                        <div className="truncate text-xs text-muted-foreground">
                          {project.description}
                        </div>
                      )}
                    </div>
                    {project.id === currentProjectId && (
                      <Check className="h-4 w-4 shrink-0" />
                    )}
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>

            <CommandGroup>
              <CommandItem onSelect={onViewAll} className="cursor-pointer">
                <Search className="mr-2 h-4 w-4" />
                <span>View all projects</span>
                {projectsData && projectsData.total > 5 && (
                  <Badge variant="secondary" className="ml-auto">
                    {projectsData.total}
                  </Badge>
                )}
              </CommandItem>
              
              <CommandItem onSelect={onCreateNew} className="cursor-pointer">
                <Plus className="mr-2 h-4 w-4" />
                <span>Create new project</span>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

function ProjectSelectorSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("flex items-center justify-between h-10 px-3 py-2 border rounded-md", className)}>
      <div className="flex items-center gap-2 min-w-0">
        <Skeleton className="h-4 w-4" />
        <Skeleton className="h-4 w-32" />
      </div>
      <Skeleton className="h-4 w-4" />
    </div>
  )
}