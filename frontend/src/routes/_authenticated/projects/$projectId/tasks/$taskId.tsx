import { createFileRoute } from '@tanstack/react-router'
import { ProjectDetail } from '@/pages/projects/ProjectDetail'

export const Route = createFileRoute('/_authenticated/projects/$projectId/tasks/$taskId')({
  component: ProjectDetail,
})