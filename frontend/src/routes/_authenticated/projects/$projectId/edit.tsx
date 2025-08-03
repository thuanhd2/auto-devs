import { createFileRoute } from '@tanstack/react-router'
import { EditProject } from '@/pages/projects/EditProject'

export const Route = createFileRoute('/_authenticated/projects/$projectId/edit')({
  component: EditProject,
})