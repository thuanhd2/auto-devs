import { createFileRoute } from '@tanstack/react-router'
import { CreateProject } from '@/pages/projects/CreateProject'

export const Route = createFileRoute('/_authenticated/projects/create')({
  component: CreateProject,
})