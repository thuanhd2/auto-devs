import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectFilters,
} from '@/types/project'
import { toast } from 'sonner'
import { projectsApi } from '@/lib/api/projects'

const QUERY_KEYS = {
  projects: ['projects'] as const,
  project: (id: string) => ['projects', id] as const,
  statistics: (id: string) => ['projects', id, 'statistics'] as const,
}

export function useProjects(filters?: ProjectFilters) {
  return useQuery({
    queryKey: [...QUERY_KEYS.projects, filters],
    queryFn: () => projectsApi.getProjects(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function useProject(projectId: string) {
  return useQuery({
    queryKey: QUERY_KEYS.project(projectId),
    queryFn: () => projectsApi.getProject(projectId),
    enabled: !!projectId,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function useProjectStatistics(projectId: string) {
  return useQuery({
    queryKey: QUERY_KEYS.statistics(projectId),
    queryFn: () => projectsApi.getProjectStatistics(projectId),
    enabled: !!projectId,
    staleTime: 2 * 60 * 1000, // 2 minutes
  })
}

export function useCreateProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (project: CreateProjectRequest) =>
      projectsApi.createProject(project),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.projects })
      toast.success('Project created successfully!')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to create project')
    },
  })
}

export function useUpdateProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      projectId,
      updates,
    }: {
      projectId: string
      updates: UpdateProjectRequest
    }) => projectsApi.updateProject(projectId, updates),
    onSuccess: (updatedProject: Project) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.projects })
      queryClient.setQueryData(
        QUERY_KEYS.project(updatedProject.id),
        updatedProject
      )
      toast.success('Project updated successfully!')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to update project')
    },
  })
}

export function useDeleteProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (projectId: string) => projectsApi.deleteProject(projectId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.projects })
      toast.success('Project deleted successfully!')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to delete project')
    },
  })
}

export function useArchiveProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (projectId: string) => projectsApi.archiveProject(projectId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.projects })
      toast.success('Project archived successfully!')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to archive project')
    },
  })
}

export function useRestoreProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (projectId: string) => projectsApi.restoreProject(projectId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.projects })
      toast.success('Project restored successfully!')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to restore project')
    },
  })
}

export function useReinitGitRepository() {
  return useMutation({
    mutationFn: (projectId: string) =>
      projectsApi.reinitGitRepository(projectId),
    onSuccess: () => {
      toast.success('Git repository reinitialized successfully!')
    },
    onError: (error: any) => {
      toast.error(
        error.response?.data?.message || 'Failed to reinitialize Git repository'
      )
    },
  })
}
